package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/sonemas/goraecle"
)

type ComplexObject struct {
	number int
}

func (o ComplexObject) ServeOracle(r *goraecle.Request) (string, error) {
	return fmt.Sprintf("The number is: %d", o.number), nil
}

func helloQueryHandler(r *goraecle.Request) (string, error) {
	name, ok := r.Args["name"] // Get name or set to "stranger" if not provided
	if !ok {
		name = "stranger"
	}

	// Error handling
	if ok && strings.Contains(strings.ToLower(name.(string)), "vitalik") {
		return "", fmt.Errorf("Go away!")
	}

	// Return an answer
	answer := fmt.Sprintf("Hello, %s", name)
	fmt.Println(answer)
	return answer, nil
}

func main() {
	// goraecle.AeDebugging = true
	goraecle.DebugMode = true

	goraecle.HandleFunc("hello", helloQueryHandler) // Register handler for hello query

	complex := ComplexObject{number: 42}
	goraecle.Handle("complex", complex)

	// Start the oracle
	if err := goraecle.ListenAndServe(ownerPrivateKey, goraecle.AeternityTestnet); err != nil {
		log.Fatalf("Failed to start oracle: %v", err)
	}
}
