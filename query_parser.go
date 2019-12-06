package goraecle

import (
	"fmt"
	"strings"
)

// QueryParser handles the raw incoming query.
//
// Custom QueryParser are useful to change the query langauge.
// QueryParser can be set at Oracle.QueryParser for use.
// ParseOracleQuery receives the raw query as a string and returns q pointer Request or an error.
type QueryParser interface {
	ParseOracleQuery(string) (*Request, error)
}

// QueryParserFunc is a wrapper which allows functions to be used as a QueryParser.
// The function needs to comply with the signature.
type QueryParserFunc func(string) (*Request, error)

// ParseOracleQuery implemented the QueryParser interface for QueryParserFunc.
func (f QueryParserFunc) ParseOracleQuery(raw string) (*Request, error) {
	return f(raw)
}

func defaultQueryParserFunc(raw string) (*Request, error) {
	var r Request

	// Expected format is "query:arg=val,..."
	// Split on : to separate query from args
	p := strings.Split(string(raw), ":")

	switch len(p) {
	case 1:
		// No arguments have been provided.
		r.Query = strings.TrimSpace(p[0])
	case 2:
		// Arguments have been provided
		r.Query = strings.TrimSpace(p[0])

		// Split arguments by comma.
		args := strings.Split(strings.TrimSpace(p[1]), ",")
		for _, arg := range args {
			a := strings.Split(strings.TrimSpace(arg), "=")
			if len(a) != 2 {
				return nil, fmt.Errorf("invalid argument %q, should be in format key=value", arg)
			}

			r.SetArg(strings.TrimSpace(a[0]), strings.TrimSpace(a[1]))
		}
	default:
		return nil, fmt.Errorf("invalid query")
	}

	return &r, nil
}

// DefaultQueryParser is the default query handler.
//
// It expects queries to be formed as: "query:arg1:val1,arg2:val2"
var DefaultQueryParser = QueryParserFunc(defaultQueryParserFunc)
