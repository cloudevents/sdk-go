package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

const (
	count = 1000
)

type envConfig struct {
	// Target URL where to send cloudevents
	Target string `envconfig:"TARGET" default:"https://localhost:8080" required:"true"`
	// Path for the client certificate used to publish to an HTTPS endpoint
	ClientCert string `envconfig:"CLIENT_CERT" default:"client.crt" required:"true"`
	// Path for the client key used to publish to an HTTPS endpoint
	ClientKey string `envconfig:"CLIENT_KEY" default:"client.key" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	os.Exit(_main(os.Args[1:], env))
}

// Example is a basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func _main(args []string, env envConfig) int {
	source, err := url.Parse("https://github.com/cloudevents/sdk-go/v2/cmd/samples/sender")
	if err != nil {
		log.Printf("failed to parse source url, %v", err)
		return 1
	}

	cert, err := tls.LoadX509KeyPair(env.ClientCert, env.ClientKey)
	if err != nil {
		log.Fatalln("unable to load certs", err)
	}
	clientCACert, err := ioutil.ReadFile(env.ClientCert)
	if err != nil {
		log.Fatal("unable to open cert", err)
	}

	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(clientCACert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            clientCertPool,
		InsecureSkipVerify: true,
	}

	tlsConfig.BuildNameToCertificate()

	seq := 0
	for _, dataContentType := range []string{"application/json", "application/xml"} {
		for _, encoding := range []cloudevents.Encoding{cloudevents.EncodingBinary, cloudevents.EncodingStructured} {

			ctx := context.Background()

			p, err := cloudevents.NewHTTP(cloudevents.WithTarget(env.Target),
				cloudevents.WithRoundTripper(&http.Transport{TLSClientConfig: tlsConfig}))
			if err != nil {
				log.Printf("failed to create protocol, %v", err)
				return 1
			}

			c, err := cloudevents.NewClient(p,
				cloudevents.WithTimeNow(),
			)
			if err != nil {
				log.Printf("failed to create client, %v", err)
				return 1
			}

			message := fmt.Sprintf("Hello, %d!", encoding)

			for i := 0; i < count; i++ {
				event := cloudevents.Event{
					Context: cloudevents.EventContextV03{
						ID:     uuid.New().String(),
						Type:   "com.cloudevents.sample.sent",
						Source: cloudevents.URIRef{URL: *source},
					}.AsV03(),
				}
				_ = event.SetData(dataContentType, &Example{
					Sequence: i,
					Message:  message,
				})

				switch encoding {
				case cloudevents.EncodingBinary:
					ctx = cloudevents.WithEncodingBinary(ctx)
				case cloudevents.EncodingStructured:
					ctx = cloudevents.WithEncodingStructured(ctx)
				}

				if resp, err := c.Request(ctx, event); err != nil {
					log.Printf("failed to send: %v", err)
				} else if resp != nil {
					fmt.Printf("Response:\n%s\n", resp)
					fmt.Printf("Got Event Response Context: %+v\n", resp.Context)
					data := &Example{}
					if err := resp.DataAs(data); err != nil {
						fmt.Printf("Got Data Error: %s\n", err.Error())
					}
					fmt.Printf("Got Response Data: %+v\n", data)
					fmt.Printf("----------------------------\n")
				} else {
					log.Printf("event sent at %s", time.Now())
				}

				seq++
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	return 0
}
