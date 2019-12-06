# Goraecle
Copyright © 2019 UAB "Sonemas" https://sonemas.com

# About this project
Library for rapid development of oracles for the aeternity blockchain in Go.

The purpose of this project is to write aeternity oracles in Go in the same manner as other go server by implementing a Handler interface.

Example:

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

# State 
This project is currently under development and not ready for production use.

# Licensing

    Licensed under the ISC license (the "License");

    Permission to use, copy, modify, and/or distribute this software for any
    purpose with or without fee is hereby granted, provided that the above
    copyright notice and this permission notice appear in all copies.

    THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
    REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
    AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
    INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
    LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
    OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
    PERFORMANCE OF THIS SOFTWARE.
