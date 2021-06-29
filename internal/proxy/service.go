/*
Copyright © 2018-2021 Neil Hemming
*/

//Package proxy contains the implementation of the oauthproxy service
package proxy

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/oauth2"
)

type (
	// doneChan is used to flag when a request has been processed
	doneChan chan struct{}

	// downstreamRequest is an active request to the downstream system for a token
	downstreamRequest struct {
		tr   tokenRequest
		w    http.ResponseWriter
		done doneChan
	}

	// entry is an entry in the cache
	entry struct {
		token      []byte
		header     http.Header
		statusCode int
		expiry     time.Time
	}

	// tokenCache is the token cache
	tokenCache map[tokenRequest]entry

	// runtime contains all the service running state
	runtime struct {
		ctx                 context.Context
		requestTimeout      time.Duration
		endpoint            string
		houseKeeperPeriod   time.Duration
		logger              LoggerFunc
		err                 error
		cancel              context.CancelFunc
		cache               tokenCache
		rwLock              sync.RWMutex
		ttl                 time.Duration
		downstream          chan downstreamRequest
		downstreamWaitGroup sync.WaitGroup
		isStopping          bool
	}
)

// Run will setup and run the server until the passed context is cancelled.
// If the server cannot be run, or fails to shut gracefully an error will be returned.
func Run(ctx context.Context, settings Settings) error {

	// Validate settings
	if err := settings.validateSettings(); err != nil {
		return err
	}

	// Create a runtime instance, this does most of the work
	rt := newRuntime(ctx, settings)
	defer rt.close()

	// Create the http server (possible add support for https here too)
	srv := http.Server{
		Addr:    settings.HTTPListenAddr,
		Handler: http.HandlerFunc(rt.handleRequest),
		//ErrorLog: &log.Logger{},
	}

	go func() {
		rt.logInfo("http listening on %s for downstream %s", settings.HTTPListenAddr, settings.Endpoint)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			rt.criticalError(err)
		}
	}()

	// Wait for exit signal
	<-rt.done()

	// Mark that we are stopping
	rt.logInfo("shutting down ...")
	rt.isStopping = true

	// Check for fatal runtime error
	if err := rt.err; err != nil {
		return err
	}

	// Commence shutdown
	ctxShutDown, cancel := context.WithTimeout(context.Background(), settings.ShutdownGracePeriod)
	defer cancel()

	return srv.Shutdown(ctxShutDown)
}

// newRuntime creates the internal runtime object used to handle the service
func newRuntime(ctx context.Context, settings Settings) *runtime {
	runningCtx, cancel := context.WithCancel(ctx)

	rt := &runtime{
		cancel:            cancel,
		ctx:               runningCtx,
		cache:             make(tokenCache),
		downstream:        make(chan downstreamRequest),
		ttl:               settings.CacheTTL,
		requestTimeout:    settings.RequestTimeout,
		endpoint:          settings.Endpoint,
		houseKeeperPeriod: settings.CacheTTL,
		logger:            settings.Logger,
	}

	// Add the house keeping to the service waitgroup
	// Ensures all services provided by runtime are completed before Run exits
	rt.downstreamWaitGroup.Add(1)
	go rt.housekeeper()

	// Start backend, will exit once shutdown complete
	for p := 0; p < settings.PoolSize; p++ {
		rt.downstreamWaitGroup.Add(1)
		go rt.downstreamService()
	}

	return rt
}

// parseRequest checks the request is for a token and extract the details
func (rt *runtime) parseRequest(w http.ResponseWriter, r *http.Request) (tokenRequest, bool) {

	// Create a token request
	tr := tokenRequest{
		path: r.URL.Path,
	}

	// Basic routing, only interested in token requests
	if r.Method != "POST" || !strings.HasSuffix(tr.path, "/token") {
		replyNotFound(w)
		return tr, false
	}

	// Parse credentials form
	if err := r.ParseForm(); err != nil {
		rt.logError("parse request error %s", err)
		replyInvalid(w)
		return tr, false
	}

	// Valiudate we are using password flow
	if grantType := r.PostFormValue("grant_type"); grantType != "password" {
		rt.logError("invlaid grant type: %s", grantType)
		replyInvalid(w)
		return tr, false
	}

	// Test for auth in header
	if u, p, ok := r.BasicAuth(); ok {

		tr.authMode = authInHeader

		if tr.clientID == "" {
			tr.clientID = u
		}
		if tr.clientSecret == "" {
			tr.clientSecret = p
		}
	} else {
		tr.authMode = authInBody
		tr.clientID = r.PostFormValue("client_id")
		tr.clientSecret = r.PostFormValue("client_secret")
	}

	// Grab details from the form data
	tr.username = r.PostFormValue("username")
	tr.password = r.PostFormValue("password")
	tr.scopes = r.PostFormValue("scope")

	return tr, true
}

// close terminates the service. It can only be called once
// use rt.cancel to initiate shutdown
func (rt *runtime) close() {

	// Cancel the context, will close house keeping
	rt.cancel()
	rt.isStopping = true

	// Close the channel, will cause downstreamService to exit
	close(rt.downstream)

	//	Wait for all downstreamService have closed
	rt.downstreamWaitGroup.Wait()

	rt.logInfo("shutdown complete")
}

// criticalError captures a any errors that can terminate the service
// Using this method avoids the need for log.Fatal type calls to exit
// the process.  Allows Run contract to be respected.
func (rt *runtime) criticalError(err error) {
	rt.err = err
	rt.cancel()
}

// logInfo logs a info message for the service
func (rt *runtime) logInfo(format string, args ...interface{}) {
	if rt.logger != nil {
		rt.logger(false, format, args...)
	}
}

// logError logs a error message for the service
func (rt *runtime) logError(format string, args ...interface{}) {
	if rt.logger != nil {
		rt.logger(true, format, args...)
	}
}

// done captures the running contexts exit channel
func (rt *runtime) done() <-chan struct{} {
	return rt.ctx.Done()
}

// handleRequest handles the incoming http token request
func (rt *runtime) handleRequest(w http.ResponseWriter, r *http.Request) {

	// Check the request isa a valid token request
	tr, matched := rt.parseRequest(w, r)
	if !matched {
		// parse handles client responses
		return
	}

	// Check to see if the token request is already in the cache
	entry := rt.lookup(tr)

	// If thee entry is not valid request a token from the down stream service.
	if entry.token == nil || entry.expiry.Before(time.Now().UTC()) {
		//	Not found or expied, request new token
		rt.requestFromDownstream(tr, w)
		return
	}

	// Found here, reply without bothering downstream service
	rt.reply(w, entry)
}

// housekeeper runs the house keeping service
func (rt *runtime) housekeeper() {
	// Mark closure in work group
	defer rt.downstreamWaitGroup.Done()

	for {

		// Set up a context to time out after the house keeping period
		wait, cancel := context.WithTimeout(rt.ctx, rt.houseKeeperPeriod)

		// Wait for timeout or the process to exit
		<-wait.Done()
		cancel()

		// Time to exit?
		if rt.ctx.Err() != nil {
			return
		}

		// Do some house keeping
		rt.clean(time.Now().UTC())
	}
}

// downstreamService executes the main down stream request processing
func (rt *runtime) downstreamService() {

	// Mark closure in work group
	defer rt.downstreamWaitGroup.Done()

	// Consume the channel queue
	for dReq := range rt.downstream {
		rt.processDownstreamRequest(dReq)
	}
}

// processDownstreamRequest handles an individual down sstream request
func (rt *runtime) processDownstreamRequest(dReq downstreamRequest) {

	// Close request
	defer close(dReq.done)

	// Check if we have started stopping
	if rt.isStopping {
		// Not available to service
		replyServiceUnavailable(dReq.w)
		return
	}

	// Double check if token exists
	entry := rt.lookup(dReq.tr)
	if entry.token != nil && entry.expiry.After(time.Now().UTC()) {
		//	Already have, return here
		rt.reply(dReq.w, entry)
		return
	}

	// Process the down stream request
	rt.getDownstreamToken(dReq.tr, dReq.w)
}

// requestFromDownstream is called when a client request needs to get a new token
func (rt *runtime) requestFromDownstream(tr tokenRequest, w http.ResponseWriter) {
	if rt.isStopping {
		replyServiceUnavailable(w)
		return
	}

	rt.logInfo("passing on downstream request for %s", tr.path)

	// Send the request to the downs stream queue
	// As HTTP Handlers need to wait for competition before exiting, so
	// we will wait on a don channel.
	c := make(doneChan)
	rt.downstream <- downstreamRequest{tr, w, c}

	// Wait
	<-c
}

// getDownstreamToken handles downstream requests
func (rt *runtime) getDownstreamToken(tr tokenRequest, w http.ResponseWriter) {

	// create a request
	req, err := tr.prepareRequest(rt.endpoint)
	if err != nil {
		// Problem creating request
		rt.logError("prepare request: %s", err)
		replyInvalid(w)
		return
	}

	rt.logInfo("downstream request for %s", req.URL)

	// Create a context to timeout in case of no response
	ctxTimeout, cancel := context.WithTimeout(rt.ctx, rt.requestTimeout)
	defer cancel()

	//	Round trip request
	resp, err := ctxhttp.Do(ctxTimeout, nil, req)
	if err != nil {
		rt.logError("send request: %s", err)
		replyInvalid(w)
		return
	}

	// Get the body
	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	if err != nil {
		// Bad read, error
		rt.logError("read body error: %s", err)
		replyInvalid(w)
		return
	}

	// Handle reply
	header := http.Header{}

	//	Set headers from downstream
	for key := range resp.Header {
		w.Header().Set(key, resp.Header.Get(key))
		header.Set(key, resp.Header.Get(key))
	}

	// Write header and body
	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	// If reply was a 500+ error don't cache the result
	if resp.StatusCode < http.StatusInternalServerError &&
		resp.StatusCode != http.StatusTooManyRequests {
		rt.update(tr, header, body, resp.StatusCode)
	}
}

// reply to a upstream request with an existing entry
func (rt *runtime) reply(w http.ResponseWriter, entry entry) {

	// Send the reply back, set standard headers
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	// Set headers from downstream
	for key := range entry.header {
		w.Header().Set(key, entry.header.Get(key))
	}

	w.WriteHeader(entry.statusCode)
	w.Write(entry.token)
}

// lookup checks the cache for an existing user
func (rt *runtime) lookup(tr tokenRequest) entry {

	rt.rwLock.RLock()
	defer rt.rwLock.RUnlock()

	return rt.cache[tr]
}

// update updates entries in the cache
func (rt *runtime) update(tr tokenRequest, header http.Header, body []byte, statusCode int) {

	rt.logInfo("update cache for %s with status %d", tr.path, statusCode)

	now := time.Now().UTC()
	expiry := now.Add(rt.ttl)

	if statusCode == http.StatusOK {
		// request succeeded, try and get expiry time from the request
		authToken := oauth2.Token{}

		// If the expiry in the token is shorter than our ttl reduce the time
		if err := json.Unmarshal(body, &authToken); err == nil {
			if authToken.Expiry.After(now) && authToken.Expiry.Before(expiry) {
				expiry = authToken.Expiry
			}
		}
	}

	e := entry{
		statusCode: statusCode,
		expiry:     expiry,
		header:     header,
		token:      body,
	}

	rt.rwLock.Lock()
	defer rt.rwLock.Unlock()
	rt.cache[tr] = e
}

func (rt *runtime) clean(now time.Time) {

	rt.logInfo("running housekeeping")

	rt.rwLock.Lock()
	defer rt.rwLock.Unlock()

	for k, entry := range rt.cache {
		if entry.expiry.Before(now) {
			delete(rt.cache, k)
		}
	}
}
