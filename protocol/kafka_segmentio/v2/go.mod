module github.com/cloudevents/sdk-go/protocol/kafka_segmentio/v2

go 1.15

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.0.0
	github.com/frankban/quicktest v1.11.3 // indirect
	github.com/golang/snappy v0.0.2 // indirect
	github.com/klauspost/compress v1.11.7 // indirect
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/segmentio/kafka-go v0.4.12
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777 // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
