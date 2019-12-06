package goraecle

import (
	"fmt"
	"sync"
)

// Request contains the query and arguments of incoming query requests.
type Request struct {
	mu sync.RWMutex

	// Query is the invoked query entry.
	Query string

	// Args contains the arguments to this query.
	Args map[string]interface{}
}

// SetArg safely sets a key/value pair.
func (r *Request) SetArg(k string, v string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Args == nil {
		r.Args = make(map[string]interface{})
	}

	// TODO: Type detection
	r.Args[k] = v
}

// HasArgs returns true if the query has any arguments
func (r *Request) HasArgs() bool {
	return r.Args != nil && len(r.Args) != 0
}

// Handler responds to an oracle query request
//
// ServeOracle received a request containing the query entry
// and arguments.
// TODO: Handler panics
type Handler interface {
	ServeOracle(*Request) (string, error)
}

// HandlerFunc is a wrapper which allows functions to be used as Oracle handlers.
// The function needs comply with the signature, for example:
// func helloHandler(r *Request) (string, error) {}
type HandlerFunc func(*Request) (string, error)

// ServeOracle implemnets the Handler interface for HandlerFunc.
func (f HandlerFunc) ServeOracle(r *Request) (string, error) {
	return f(r)
}

// OracleMux is a muxer for oracles
type OracleMux struct {
	mu      sync.RWMutex
	entries map[string]Handler
}

// NewOracleMux created a new instance of OracleMux
func NewOracleMux() *OracleMux { return new(OracleMux) }

// Handle registers a new Hendler(h) for the provider entry(q).
func (mux *OracleMux) Handle(q string, h Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	// q may not be empty
	if q == "" {
		panic("goraecle: empty query")
	}

	// nil handlers can't be accepted
	if h == nil {
		panic("goraecle: nil handler")
	}

	// Initialize the entry map if needed
	if mux.entries == nil {
		mux.entries = make(map[string]Handler)
	}

	// make sure query entries are only registered once.
	if _, ok := mux.entries[q]; ok {
		panic("goraecle: query is already registered")
	}

	mux.entries[q] = h
}

// HandleFunc registers the provided function as a HandlerFunc Handler at the muxer.
// f should comply with the HandlerFunc signature.
func (mux *OracleMux) HandleFunc(q string, f func(*Request) (string, error)) {
	if f == nil {
		panic("goraecle: nil handler")
	}

	mux.Handle(q, HandlerFunc(f))
}

// ServeOracle implements the Handler interface for OracleMux
func (mux *OracleMux) ServeOracle(r *Request) (string, error) {
	// Make sure the query entry exists.
	h, ok := mux.entries[r.Query]
	if !ok {
		return "", fmt.Errorf("invalid query")
	}

	// Pass the request to the entry handler.
	return h.ServeOracle(r)
}

var (
	defaultOracleMux OracleMux

	// DefaultOracleMux is the default OracleMux instance for this package.
	DefaultOracleMux = &defaultOracleMux
)

// Handle registers a new Handler at DefaultOracleMux.
func Handle(q string, h Handler) { DefaultOracleMux.Handle(q, h) }

// HandleFunc registers a new HandlerFunc Handler at DefaultOracleMux.
func HandleFunc(q string, h func(*Request) (string, error)) { DefaultOracleMux.HandleFunc(q, h) }
