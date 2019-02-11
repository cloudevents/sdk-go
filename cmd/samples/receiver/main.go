package main

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/kelseyhightower/envconfig"
	"log"
	"net/http"
	"os"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port string `envconfig:"PORT" default:"8080"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	os.Exit(_main(os.Args[1:], env))
}

type Receiver struct{}

func (r *Receiver) Receive(event cloudevents.Event) {
	fmt.Printf("Got Event: %v\n", event)
	fmt.Printf("----------------------------\n")
}

//
//func (r *Receiver) ServeHTTP(w http.ResponseWriter, req *http.Request) {
//	marshaller := v02.NewDefaultHTTPMarshaller()
//	// req is *http.Request
//	event, err := marshaller.FromRequest(req)
//	if err != nil {
//		log.Printf("Unable to parse event from http Request: " + err.Error())
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(`Invalid request`))
//		return
//	}
//	if t, ok := event.Get("type"); ok {
//		fmt.Printf("type: %s\n", t)
//	}
//	if s, ok := event.GetURL("source"); ok {
//		fmt.Printf("source: %s\n", s.RequestURI())
//	}
//	if t, ok := event.GetTime("time"); ok {
//		fmt.Printf("time: %s\n", t)
//	}
//	if d, ok := event.GetBinary("data"); ok {
//		fmt.Printf("data as binary: %s\n", string(d))
//	}
//	if d, ok := event.GetMap("data"); ok {
//		fmt.Printf("data as map:\n")
//		for k, v := range d {
//			fmt.Printf("\t%q: %v\n", k, v)
//		}
//
//	}
//	fmt.Printf("----------------------------\n")
//	w.WriteHeader(http.StatusNoContent)
//}

func _main(args []string, env envConfig) int {
	r := &Receiver{}
	t := &cloudeventshttp.Transport{Receiver: r}

	log.Printf("listening on port %s\n", env.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env.Port), t))

	return 0
}
