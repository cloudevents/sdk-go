package binding

import (
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

// Translate creates a new message with the same content and mode as 'in',
// using the given makeBinary or makeStruct functions.
// TODO(slinkydeveloper) This function still makes sense to exists?
func Translate(in Message,
	makeBinary func(ce.Event) (Message, error),
	makeStruct func(string, []byte) (Message, error),
) (Message, error) {
	//if f, b, err := in.Structured(); err == nil {
	//	if f != "" && len(b) > 0 {
	//		return makeStruct(f, b)
	//	}
	//} else {
	//	return nil, err
	//}
	//e, err := in.Event()
	//if err != nil {
	//	return nil, err
	//}
	//return makeBinary(e)
	return nil, nil
}
