/*
Package cloudevents provides primitives to work with CloudEvents specification: https://github.com/cloudevents/spec.


Parsing Event from JSON:
	event, err := cloudEvents.FromJSON(data)
	if err != nil {
		panic("Unable to parse event from JSON: " + err.String())
	}


Creating a minimal CloudEvent in version 0.1:
	event := cloudevents.EventV01{
		EventType:        "com.example.file.created",
		EventTypeVersion: "0.1",
		Source:           "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		EventID:          "ea35b24ede421",
		EventTime:        time.Now(),
	}

Parsing Event from http request:
	import "github.com/dispatchframework/cloudevents-go-sdk/httptransport"
	...
	func ServeHTTP(rw http.ResponseWriter, req *http.Request) {
		event, err := httptransport.FromRequest(req)
		if err != nil {
			// parse error
		}
	}



The goal of this package is to provide support for all released versions of CloudEvents, ideally while maintaining
the same API. It will use semantic versioning with following rules:
* MAJOR version increments when backwards incompatible changes is introduced.
* MINOR version increments when backwards compatible feature is introduced INCLUDING support for new CloudEvents version.
* PATCH version increments when a backwards compatible bug fix is introduced.
*/
package cloudevents
