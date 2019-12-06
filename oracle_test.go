package goraecle

import (
	"testing"

	"github.com/aeternity/aepp-sdk-go/v7/account"
	"github.com/aeternity/aepp-sdk-go/v7/binary"
)

func TestParseQuery(t *testing.T) {
	testAccount, err := account.New()
	if err != nil {
		t.Fatalf("expected to pass, but got: %v", err)
	}

	testHandler := func(r *Request) (string, error) {
		return "success", nil
	}
	HandleFunc("test_process_query", testHandler)

	o, err := NewOracle(testAccount.SigningKeyToHexString(), AeternityTestnet)
	if err != nil {
		t.Fatalf("expected to pass, but got: %v", err)
	}

	answ, err := o.processQuery(binary.Encode(binary.PrefixOracleQueryID, []byte("test_process_query:name=Arjan van Eersel")))
	if err != nil {
		t.Fatalf("expected to pass, but got: %v", err)
	}

	if answ != "success" {
		t.Fatal("incorrect answer")
	}
}
