package cloudevents

// Package cloudevents alias' common functions and types to improve discoverability and reduce
// the number of imports for simple HTTP clients.

import (
	"github.com/cloudevents/sdk-go/pkg/client"
	"github.com/cloudevents/sdk-go/pkg/context"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/observability"
	"github.com/cloudevents/sdk-go/pkg/transport/http"
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
type EventContextV03 = event.EventContextV03

// Custom Types

type Timestamp = types.Timestamp
type URLRef = types.URLRef
type URIRef = types.URIRef

// HTTP Protocol

type HTTPOption http.Option
type HTTPProtocolOption http.ProtocolOption

type HTTPTransport = http.Protocol
type HTTPEncoding = http.Encoding

const (
	// ReadEncoding

	ApplicationXML                  = event.ApplicationXML
	ApplicationJSON                 = event.ApplicationJSON
	ApplicationCloudEventsJSON      = event.ApplicationCloudEventsJSON
	ApplicationCloudEventsBatchJSON = event.ApplicationCloudEventsBatchJSON
	Base64                          = event.Base64

	// Event Versions

	VersionV1  = event.CloudEventsVersionV1
	VersionV03 = event.CloudEventsVersionV03

	HTTPBinaryEncoding     = http.Binary
	HTTPStructuredEncoding = http.Structured
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

	// HTTP Protocol

	NewHTTPTransport = http.New
	NewHTTPProtocol  = http.NewProtocol

	// HTTP Protocol Options

	WithTarget             = http.WithTarget
	WitHHeader             = http.WithHeader
	WithShutdownTimeout    = http.WithShutdownTimeout
	WithEncoding           = http.WithEncoding
	WithStructuredEncoding = http.WithStructuredEncoding
	WithPort               = http.WithPort
	WithPath               = http.WithPath
	WithMiddleware         = http.WithMiddleware
	WithListener           = http.WithListener
	WithHTTPTransport      = http.WithHTTPTransport
)
