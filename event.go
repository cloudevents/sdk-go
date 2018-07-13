package cloudevents

// TODO: There is a value in providing a general-use function like this, but is it feasible?

// NewEvent takes map of attributes and optional data, and produces a CloudEvent with version that best matches
// the provided input
func NewEvent(context map[string]interface{}, data interface{}) (CloudEvent, error) {
	panic("not implemented")
}

// FromJSON takes the data input and produces an instance of CloudEvent, if possible.
// TODO: might make more sense in a encoding sub-package
func FromJSON(data []byte) (CloudEvent, error) {
	panic("not implemented")
}
