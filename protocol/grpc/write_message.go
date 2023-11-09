/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package grpc

import (
	"bytes"
	"context"
	"fmt"
	"io"

	pb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WritePubMessage fills the provided pubMessage with the message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WritePubMessage(ctx context.Context, m binding.Message, pbEvt *pb.CloudEvent, transformers ...binding.Transformer) error {
	structuredWriter := (*pbEventWriter)(pbEvt)
	binaryWriter := (*pbEventWriter)(pbEvt)

	_, err := binding.Write(
		ctx,
		m,
		structuredWriter,
		binaryWriter,
		transformers...,
	)
	return err
}

type pbEventWriter pb.CloudEvent

var (
	_ binding.StructuredWriter = (*pbEventWriter)(nil)
	_ binding.BinaryWriter     = (*pbEventWriter)(nil)
)

func (b *pbEventWriter) SetStructuredEvent(ctx context.Context, f format.Format, event io.Reader) error {
	if b.Attributes == nil {
		b.Attributes = make(map[string]*pb.CloudEventAttributeValue)
	}

	b.Attributes[contenttype], _ = attributeFor(f.MediaType())

	var buf bytes.Buffer
	_, err := io.Copy(&buf, event)
	if err != nil {
		return err
	}

	// TODO: check the data content type and set the right data format
	b.Data = &pb.CloudEvent_BinaryData{
		BinaryData: buf.Bytes(),
	}

	return nil
}

func (b *pbEventWriter) Start(ctx context.Context) error {
	if b.Attributes == nil {
		b.Attributes = make(map[string]*pb.CloudEventAttributeValue)
	}

	return nil
}

func (b *pbEventWriter) End(ctx context.Context) error {
	return nil
}

func (b *pbEventWriter) SetData(reader io.Reader) error {
	buf, ok := reader.(*bytes.Buffer)
	if !ok {
		buf = new(bytes.Buffer)
		_, err := io.Copy(buf, reader)
		if err != nil {
			return err
		}
	}

	b.Data = &pb.CloudEvent_BinaryData{
		BinaryData: buf.Bytes(),
	}

	return nil
}

func (b *pbEventWriter) SetAttribute(attribute spec.Attribute, value interface{}) error {
	fmt.Printf("setting attribute: %s, value: %v\n", attribute.Name(), value)
	switch attribute.Kind() {
	case spec.SpecVersion:
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid SpecVersion type, expected string got %T", value)
		}
		b.SpecVersion = val
		return nil
	case spec.ID:
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid ID type, expected string got %T", value)
		}
		b.Id = val
		return nil
	case spec.Source:
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid Source type, expected string got %T", value)
		}
		b.Source = val
		return nil
	case spec.Type:
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid Type type, expected string got %T", value)
		}
		b.Type = val
		return nil
	case spec.DataContentType:
		if value == nil {
			delete(b.Attributes, contenttype)
		}
		b.Attributes[contenttype], _ = attributeFor(value)
		return nil
	case spec.DataSchema:
		if value == nil {
			delete(b.Attributes, prefix+dataSchema)
		}
		b.Attributes[prefix+dataSchema], _ = attributeFor(value)
		return nil
	case spec.Subject:
		if value == nil {
			delete(b.Attributes, prefix+subject)
		}
		b.Attributes[prefix+subject], _ = attributeFor(value)
		return nil
	case spec.Time:
		if value == nil {
			delete(b.Attributes, prefix+time)
		}
		b.Attributes[prefix+time], _ = attributeFor(value)
		return nil
	default:
		if value == nil {
			delete(b.Attributes, prefix+attribute.Name())
		}
		b.Attributes[prefix+attribute.Name()], _ = attributeFor(value)
	}

	return nil
}

func (b *pbEventWriter) SetExtension(name string, value interface{}) (err error) {
	fmt.Printf("setting extension: %s, value: %v\n", name, value)
	if value == nil {
		delete(b.Attributes, prefix+name)
	}

	b.Attributes[prefix+name], err = attributeFor(value)

	return
}

func attributeFor(v interface{}) (*pb.CloudEventAttributeValue, error) {
	vv, err := types.Validate(v)
	if err != nil {
		return nil, err
	}
	attr := &pb.CloudEventAttributeValue{}
	switch vt := vv.(type) {
	case bool:
		attr.Attr = &pb.CloudEventAttributeValue_CeBoolean{
			CeBoolean: vt,
		}
	case int32:
		attr.Attr = &pb.CloudEventAttributeValue_CeInteger{
			CeInteger: vt,
		}
	case string:
		attr.Attr = &pb.CloudEventAttributeValue_CeString{
			CeString: vt,
		}
	case []byte:
		attr.Attr = &pb.CloudEventAttributeValue_CeBytes{
			CeBytes: vt,
		}
	case types.URI:
		attr.Attr = &pb.CloudEventAttributeValue_CeUri{
			CeUri: vt.String(),
		}
	case types.URIRef:
		attr.Attr = &pb.CloudEventAttributeValue_CeUriRef{
			CeUriRef: vt.String(),
		}
	case types.Timestamp:
		attr.Attr = &pb.CloudEventAttributeValue_CeTimestamp{
			CeTimestamp: timestamppb.New(vt.Time),
		}
	default:
		return nil, fmt.Errorf("unsupported attribute type: %T", v)
	}
	return attr, nil
}
