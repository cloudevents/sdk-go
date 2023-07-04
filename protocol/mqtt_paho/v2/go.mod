module github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2

go 1.1

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.0.0-00010101000000-000000000000
	github.com/eclipse/paho.golang v0.11.0
)

require golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
