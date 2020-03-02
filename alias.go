package cloudevents

// Package cloudevents alias' common functions and types to improve discoverability and reduce
// the number of imports for simple HTTP clients.

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/observability"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/types"
)

// Client

type ClientOption client.Option
type Client = client.Client
type ConvertFn = client.ConvertFn

// Event

type Event = event.Event
type EventResponse = event.EventResponse

// Context

type EventContext = event.EventContext
type EventContextV1 = event.EventContextV1
type EventContextV01 = event.EventContextV01
type EventContextV02 = event.EventContextV02
type EventContextV03 = event.EventContextV03

// Custom Types

type Timestamp = types.Timestamp
type URLRef = types.URLRef

// HTTP Transport

type HTTPOption http.Option
type HTTPTransport = http.Transport
type HTTPTransportContext = http.TransportContext
type HTTPTransportResponseContext = http.TransportResponseContext
type HTTPEncoding = http.Encoding

const (
	// Encoding

	ApplicationXML                  = event.ApplicationXML
	ApplicationJSON                 = event.ApplicationJSON
	ApplicationCloudEventsJSON      = event.ApplicationCloudEventsJSON
	ApplicationCloudEventsBatchJSON = event.ApplicationCloudEventsBatchJSON
	Base64                          = event.Base64

	// Event Versions

	VersionV1  = event.CloudEventsVersionV1
	VersionV01 = event.CloudEventsVersionV01
	VersionV02 = event.CloudEventsVersionV02
	VersionV03 = event.CloudEventsVersionV03

	// HTTP Transport Encodings

	HTTPBinaryV1      = http.BinaryV1
	HTTPStructuredV1  = http.StructuredV1
	HTTPBatchedV1     = http.BatchedV1
	HTTPBinaryV01     = http.BinaryV01
	HTTPStructuredV01 = http.StructuredV01
	HTTPBinaryV02     = http.BinaryV02
	HTTPStructuredV02 = http.StructuredV02
	HTTPBinaryV03     = http.BinaryV03
	HTTPStructuredV03 = http.StructuredV03
	HTTPBatchedV03    = http.BatchedV03

	// Context HTTP Transport Encodings

	Binary     = http.Binary
	Structured = http.Structured
)

var (
	// ContentType Helpers

	StringOfApplicationJSON                 = event.StringOfApplicationJSON
	StringOfApplicationXML                  = event.StringOfApplicationXML
	StringOfApplicationCloudEventsJSON      = event.StringOfApplicationCloudEventsJSON
	StringOfApplicationCloudEventsBatchJSON = event.StringOfApplicationCloudEventsBatchJSON
	StringOfBase64                          = event.StringOfBase64

	// Client Creation

	NewClient        = client.New
	NewDefaultClient = client.NewDefault

	// Client Options

	WithEventDefaulter      = client.WithEventDefaulter
	WithUUIDs               = client.WithUUIDs
	WithTimeNow             = client.WithTimeNow
	WithConverterFn         = client.WithConverterFn
	WithDataContentType     = client.WithDataContentType
	WithoutTracePropagation = client.WithoutTracePropagation

	// Event Creation

	NewEvent = event.New

	// Tracing

	EnableTracing = observability.EnableTracing

	// Context

	ContextWithTarget   = context.WithTarget
	TargetFromContext   = context.TargetFrom
	ContextWithEncoding = context.WithEncoding
	EncodingFromContext = context.EncodingFrom

	// Custom Types

	ParseTimestamp = types.ParseTimestamp
	ParseURLRef    = types.ParseURLRef
	ParseURIRef    = types.ParseURIRef
	ParseURI       = types.ParseURI

	// HTTP Transport

	NewHTTPTransport = http.New

	// HTTP Transport Options

	WithTarget               = http.WithTarget
	WithMethod               = http.WithMethod
	WitHHeader               = http.WithHeader
	WithShutdownTimeout      = http.WithShutdownTimeout
	WithEncoding             = http.WithEncoding
	WithContextBasedEncoding = http.WithContextBasedEncoding
	WithBinaryEncoding       = http.WithBinaryEncoding
	WithStructuredEncoding   = http.WithStructuredEncoding
	WithPort                 = http.WithPort
	WithPath                 = http.WithPath
	WithMiddleware           = http.WithMiddleware
	WithLongPollTarget       = http.WithLongPollTarget
	WithListener             = http.WithListener
	WithHTTPTransport        = http.WithHTTPTransport

	// HTTP Context

	HTTPTransportContextFrom = http.TransportContextFrom
	ContextWithHeader        = http.ContextWithHeader
	SetContextHeaders        = http.SetContextHeaders
)
