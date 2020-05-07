package event

import (
	"fmt"
)

type EventValidationError map[string]error

func (e EventValidationError) Error() string {
	return fmt.Sprintf("validation errors: %+v", map[string]error(e))
}

// Validate performs a spec based validation on this event.
// Validation is dependent on the spec version specified in the event context.
func (e Event) Validate() EventValidationError {
	if e.Context == nil {
		return EventValidationError{"specversion": fmt.Errorf("missing Event.Context")}
	}

	errs := map[string]error{}
	if e.FieldErrors != nil {
		for k, v := range errs {
			errs[k] = v
		}
	}

	if fieldErrors := e.Context.Validate(); fieldErrors != nil {
		for k, v := range fieldErrors {
			errs[k] = v
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
