/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/kelseyhightower/envconfig"
)

type RawHTTP struct {
	Port int `envconfig:"PORT" default:"8080"`
}

func (raw *RawHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if reqBytes, err := httputil.DumpRequest(r, true); err == nil {
		log.Printf("Raw HTTP Request:\n%+v", string(reqBytes))
		_, _ = w.Write(reqBytes)
	} else {
		log.Printf("Failed to call DumpRequest: %s", err)
	}
	fmt.Println("------------------------------")
}

func main() {
	var env RawHTTP
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}
	log.Printf("Starting listening on :%d\n", env.Port)
	log.Println(http.ListenAndServe(fmt.Sprintf(":%d", env.Port), &env))
}
