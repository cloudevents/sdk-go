package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/cloudevents/sdk-go/pkg/event"
)

// ResultsFull is called when passed in to send
type ResultsFull func(context.Context, event.Event)

type resultsFn struct {
	numIn   int
	fnValue reflect.Value

	hasContextIn bool
	hasEventIn   bool
}

const (
	resultsInParamUsage  = "expected a function taking (context.Context, event.Event) ordered"
	resultsOutParamUsage = "expected a function returning nothing"
)

// parseResultsFn creates a resultsFn wrapper class that is used by the client to
// validate and invoke the provided function.
// Valid fn signatures are:
// * func()
// * func(context.Context)
// * func(*event.Event)
// * func(context.Context, *event.Event)
//
func parseResultsFn(fn interface{}) (*resultsFn, error) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return nil, errors.New("must pass a function to handle events")
	}

	r := &resultsFn{
		fnValue: reflect.ValueOf(fn),
		numIn:   fnType.NumIn(),
	}
	if err := r.validate(fnType); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *resultsFn) invoke(ctx context.Context, event *event.Event) {
	args := make([]reflect.Value, 0, r.numIn)

	if r.numIn > 0 {
		if r.hasContextIn {
			args = append(args, reflect.ValueOf(ctx))
		}
		if r.hasEventIn {
			args = append(args, reflect.ValueOf(event))
		}
	}
	// Not expecting a return.
	_ = r.fnValue.Call(args)
}

// Verifies that the inputs to a function have a valid signature
// Valid input is to be [0, all] of
// context.Context, event.Event, *event.EventResponse in this order.
func (r *resultsFn) validateInParamSignature(fnType reflect.Type) error {
	r.hasContextIn = false
	r.hasEventIn = false

	switch fnType.NumIn() {
	case 2:
		// can be event.Event
		if fnType.In(1).ConvertibleTo(eventTypePtr) {
			r.hasEventIn = true
		} else {
			return fmt.Errorf("%s; cannot convert parameter 1 from %s to event.Event", resultsInParamUsage, fnType.In(1))
		}
		fallthrough
	case 1:
		// context.Context or can be event.Event
		if fnType.In(0).ConvertibleTo(contextType) {
			r.hasContextIn = true
		} else {
			if !fnType.In(0).ConvertibleTo(eventTypePtr) {
				return fmt.Errorf("%s; cannot convert parameter 0 from %s to context.Context, event.Event", resultsInParamUsage, fnType.In(0))
			} else if r.hasEventIn {
				return fmt.Errorf("%s; duplicate parameter of type event.Event", resultsInParamUsage)
			} else {
				r.hasEventIn = true
			}
		}
		fallthrough
	case 0:
		return nil
	default:
		return fmt.Errorf("%s; function has too many parameters (%d)", resultsInParamUsage, fnType.NumIn())
	}
}

// Verifies that the outputs of a function have a valid signature
// Valid output signatures:
// (), (error)
func (r *resultsFn) validateOutParamSignature(fnType reflect.Type) error {
	switch fnType.NumOut() {
	case 0:
		return nil
	default:
		return fmt.Errorf("%s; function has too many return types (%d)", resultsOutParamUsage, fnType.NumOut())
	}
}

// validateReceiverFn validates that a function has the right number of in and
// out params and that they are of allowed types.
func (r *resultsFn) validate(fnType reflect.Type) error {
	if err := r.validateInParamSignature(fnType); err != nil {
		return err
	}
	if err := r.validateOutParamSignature(fnType); err != nil {
		return err
	}
	return nil
}
