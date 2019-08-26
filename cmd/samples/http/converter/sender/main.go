package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// Target URL where to send post
	Target string `envconfig:"TARGET" default:"http://localhost:8080" required:"true"`
}

// Basic data struct.
type Example struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	data := &Example{
		ID:      123,
		Message: "Hello, World!",
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = http.Post(env.Target, "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Fatalln(err)
	}
}
