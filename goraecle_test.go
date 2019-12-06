package goraecle_test

import (
	"log"
	"testing"

	"github.com/sonemas/goraecle"
)

// var senderPrivateKey = os.Getenv("SENDER_PRIVATE_KEY")
// TODO: Improve testing

func TestDefaults(t *testing.T) {
	testHandler := func(r *goraecle.Request) (string, error) {
		log.Println("entered test handler")
		return "success", nil
	}
	goraecle.HandleFunc("test_defaults", testHandler)

	queryParser := goraecle.DefaultQueryParser

	req, err := queryParser.ParseOracleQuery("test_defaults:name=Arjan van Eersel")
	if err != nil {
		t.Fatalf("expected to pass, but got: %v", err)
	}

	if !req.HasArgs() {
		t.Fatalf("expected to have args")
	}

	if req.Query != "test_defaults" {
		t.Fatalf("expected req.Query to be test_defaults, but got: %v", req.Query)
	}

	got, ok := req.Args["name"]
	if !ok {
		t.Fatalf("expected to have a name arg, but got: %+v", req.Args)
	}

	if v := got.(string); v != "Arjan van Eersel" {
		t.Fatalf("expected v to be Arjan van Eersel, but got %v", v)
	}

	answ, err := goraecle.DefaultOracleMux.ServeOracle(req)
	if err != nil {
		t.Fatalf("expected to pass, but got: %v", err)
	}

	if answ != "success" {
		t.Fatalf("expected answer to be success, but got: %v", answ)
	}
}
