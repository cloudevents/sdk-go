package main

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gin-gonic/gin"
)

func receive(event cloudevents.Event) {
	fmt.Printf("Got an Event: %s", event)
}

func index(c *gin.Context) {
	c.JSON(http.StatusOK, "Welcome to CloudEvents")
}

func healthz(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func cloudEventsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		p, err := cloudevents.NewHTTP()
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to create protocol")
		}

		ceh, err := cloudevents.NewHTTPReceiveHandler(c, p, receive)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("failed to create handler")
		}

		ceh.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.GET("/", index)
	r.GET("/healthz", healthz)
	r.POST("/", cloudEventsHandler())

	log.Fatal().
		Err(http.ListenAndServe(":8080", r))
}
