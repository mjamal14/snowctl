package client

import "net/http"

// Authenticator applies authentication to HTTP requests.
type Authenticator interface {
	Apply(req *http.Request) error
}

// BasicAuth authenticates using username and password.
type BasicAuth struct {
	Username string
	Password string
}

// Apply sets the Basic auth header on the request.
func (b *BasicAuth) Apply(req *http.Request) error {
	req.SetBasicAuth(b.Username, b.Password)
	return nil
}
