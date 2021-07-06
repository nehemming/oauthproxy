# oauthproxy

A oauth2 proxy token caching service for password authentication flows.


![Status](https://img.shields.io/badge/Status-ALPHA-red?style=for-the-badge)
[![Build Status](https://img.shields.io/circleci/build/gh/nehemming/oauthproxy/master?style=for-the-badge)](https://github.com/nehemming/oauthproxy) 
[![Release](https://img.shields.io/github/v/release/nehemming/oauthproxy.svg?style=for-the-badge)](https://github.com/nehemming/oauthproxy/releases/latest)
[![Coveralls](https://img.shields.io/coveralls/github/nehemming/oauthproxy?style=for-the-badge)](https://coveralls.io/github/nehemming/oauthproxy)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=for-the-badge)](/LICENSE)
[![GoReportCard](https://goreportcard.com/badge/github.com/nehemming/oauthproxy?test=0&style=for-the-badge)](https://goreportcard.com/report/github.com/nehemming/oauthproxy)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/goreleaser/goreleaser)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![Uses: cirocket](https://img.shields.io/badge/Uses-cirocket-orange?style=for-the-badge)](https://github.com/nehemming/cirocket)
[![Uses: GoReleaser](https://img.shields.io/badge/uses-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)


## Introduction

oauthproxy has been developed to support local testing suites that require multiple independent tests to authenticate with a external token provider using the same credentials.   Many authentication providers implement rate limiting, and where as normally this is a reasonable restriction it can be problematic where multiple tests are all requesting authentication simultaneously.   In these scenarios Oauthproxy acts as a substitute token provider returning a cached copy of the token issues by the down stream provider.   The downstream provider is only called when tokens need to be refreshed.   To use the proxy client applications only need to change their token provider url too the oauthproxy local url. 

oauthproxy stores cached tokens and their authentication credentials in memory and does not persist them to disk.   The service only listens on http, which is not encrypted, for localhost connections.   This is don to help ensure non encrypted token traffic is not sent over a non local network.

> Do not send credentials over networks using the HTTP protocol, always use HTTPS

## Installation

oauthproxy source can be installed by running, it requires go version 1.16 or grater installed.

```sh
go get -u github.com/nehemming/oauthproxy
```

or by cloning this repository

```sh
git clone https://github.com/nehemming/oauthproxy.git
cd oauthproxy
go install ./...
```

## Running the service

To run oauthproxy the proxy server:

```sh
oauthproxy serve --downstreeam <url-of-auth-provider>
```

This will run the server, listening by default on 127.0.0.1:8090 for incoming request and forward these as necessary down stream to the `url-of-auth-provider` provider.

### Request URL's

The service returns Not Found (404) fo all requests except POST requests where the url ends in `/token`.  The request will be rejected if the token request `grant_type` is not `password` too.  If the result of a previous downstream token request is not cached the service will forward the request to the down stream service.   The url of the down stream request is formed by concatenating the inbound request's url path with the `url-of-auth-provider`.  E.g.

```
inbound req request: http://localhost:8090/v1/token
url of auth provider: https://provider.com/tenant

down stream request: https://provider.com/tenant/v1/token
```

oauthproxy supports requests passing the client ID and client secrets in the header or in the POST body.  The inbound convention will be used with the down stream provider.

### What is cached?

The server caches the response and status from the downstream provider.  This includes all status codes below 500, except 429 (too much data).  The reasoning behind this is if the credentials are invalid the response given for them can still be cached.

All downstream responses are cached for a finite period of time.  This limit to this period is defined by the config entry `serve.cacheTTL` or OAP_SERVE_CACHETTL environment variable.  The expiry time is calculated by adding the `cacheTTL` duration onto the current UTC time. 

If a downstream token request is successful (StatusOK) the token's `expiry` value is inspected and this is sooner than the default expiry time, its value is used instead.   Note it cannot extend the cache time beyond the `cacheTTL` time.

A house keeping task runs in the background removing any expired tokens.

## Request Command
In addition to running the proxy service `oauthproxy` can send token requests.  This is intended as a simple method of testing connection credentials prior to using the cache.

```sh
oauthproxy request <secrets-json-file>
```

The `secrets-json-file` contains the credentials to test and must be in the following format:

```json
{
  "api": {
    "tokenUrl": "http://localhost:8090/token",
    "username": "<username>",
    "password": "<password>",
    "clientId": "<clientID>",
    "clientSecret": "<secret>",
    "scopes": [
      "openid", 
      "<other>"
    ]
  }
}
```

The command will output the response from the server.

A standard use case for this tool is to set up the credentials with `tokenUrl` pointing at the downstream providers full URL and test it returns a valid response.

Then the `secrets-json-file` can be edited to update the `tokenUrl` t the local cache.  Rerunning the command should produce the same results.

>Great care should be taken with credentials stored in files.  This facility is intended only for testing authentication. To prevent accidental check in this projects default git ignore and docker ignore rules exclude `.secrets*` files.

## Configuration

oauthproxy supports configuration options being passed bv the command line, environment variables or defined in a configuration file.

The default config file is called `.oauthproxy` and is located by searching in order the current working directory then the user's home folder.

The config file use the YAML format.

### Serve config entries

|entry|env variable|description|
|-|-|-|
|downstream|OAP_SERVE_DOWNSTREAM|URL or the down stream service|
|port|OAP_SERVE_PORT|Port the service listens on localhost for HTTP connections|
|cacheTTL|OAP_SERVE_CACHETTL|Default period to cache responses from the down stream provider.  Value is in Minutes.  The housekeeping service runs every `cacheTTL` minutes as well.|
|timeout|OAP_SERVE_TIMEOUT|Timeout period in seconds to wait for responses from the downstream provider| 
|shutdown|OAP_SERVE_SHUTDOWN|Period of time the service will wait once a `SIGTERM` or `SIGINT` (ctrl-c) signal has been received to complete requests before terminating|
|silent|OAP_SERVE_SILENT|If set to true the service will not output logging information.  This can be useful when running as part of a test suite as a background service.|
|poolSize|OAP_SERVE_POOLSIZE|Specifies the number of threads servicing downstream requests,   The default and recommendation is to set this to 2|

## Contributing

We would welcome contributions to this project.  Please read our [CONTRIBUTION](https://github.com/nehemming/oauthproxy/blob/master/CONTRIBUTING.md) file for further details on how you can participate or report any issues.

## License

This software is licensed under the [Apache License](http://www.apache.org/licenses/). 











