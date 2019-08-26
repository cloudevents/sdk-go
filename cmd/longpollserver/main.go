package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// This is a simple example of a long poll server.

var timeout = time.Second * 5 //30
var maxOpMS = 8000            //60 * 1000 // 60 seconds -> 50% chance of StatusNotModified.

func init() {
	rand.Seed(time.Now().Unix())
}

func operation(ctx context.Context, ch chan<- string) {
	wait := time.Millisecond * time.Duration(rand.Intn(maxOpMS))
	fmt.Printf(" -> Wait %v\n", wait)
	select {
	case <-time.After(wait):
		ch <- fmt.Sprintf(`{"waited":"%v"}`, wait)
	case <-ctx.Done():
		fmt.Println(" -> Op canceled")
		return
	}
	fmt.Println(" -> Op done")
}

func handler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Request")
	ch := make(chan string) // TODO: send real struct.
	go operation(req.Context(), ch)

	select {
	case result := <-ch:
		fmt.Println("Response")
		_, _ = w.Write([]byte(result))
	case <-time.After(timeout):
		fmt.Println("Timeout")
		w.WriteHeader(http.StatusNotModified)
	case <-req.Context().Done():
		fmt.Println("Client has disconnected")
	}
	close(ch)
	fmt.Println("Done")
}

func main() {
	addr := ":8181"
	http.HandleFunc("/", handler)
	server := &http.Server{Addr: addr, Handler: nil}

	log.Printf("Starting on %s\n", addr)
	log.Print(server.ListenAndServe())
}
