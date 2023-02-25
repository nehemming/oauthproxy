module github.com/nehemming/oauthproxy

go 1.16

require (
	github.com/hashicorp/go-multierror v1.1.1
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nehemming/cirocket v0.1.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602
	golang.org/x/sys v0.1.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

// Clean up of Vulnerable Dependencies
// go list -json -m all | nancy sleuth --quiet
replace (
	github.com/coreos/bbolt v1.3.6 => go.etcd.io/bbolt v1.3.6
	github.com/coreos/etcd v3.3.13+incompatible => go.etcd.io/bbolt v1.3.6
	github.com/coreos/etcd/v3 v3.5.0 => go.etcd.io/etcd/v3 v3.5.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible => github.com/golang-jwt/jwt v3.2.1+incompatible
)
