package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

type RawHTTP struct{}

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
	log.Println(http.ListenAndServe(":8282", &RawHTTP{}))
}
