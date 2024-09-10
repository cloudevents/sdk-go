/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import "fmt"

const (
	// field names
	fieldURL              = "URL"
	fieldConsumerConfig   = "ConsumerConfig"
	fieldSendSubject      = "SendSubject"
	fieldPullConsumerOpts = "PullConsumerOptions"
	fieldPublishOptions   = "PublishOptions"
	fieldFilterSubjects   = "FilterSubjects"

	// error messages
	messageNoConnection                 = "URL or nats connection must be given."
	messageConflictingConnection        = "URL and nats connection were both given."
	messageNoConsumerConfig             = "No consumer config was given."
	messageNoFilterSubjects             = "No filter subjects were given."
	messageMoreThanOneStream            = "More than one stream for given filter subjects."
	messageNoSendSubject                = "Cannot send without a NATS subject defined."
	messageMoreThanOneConsumerConfig    = "More than one consumer config given."
	messageReceiverOptionsWithoutConfig = "Receiver options given without consumer config."
	messageSenderOptionsWithoutSubject  = "Sender options given without send subject."
)

// validateOptions runs after all options have been applied and makes sure needed options were set correctly.
func validateOptions(p *Protocol) error {
	if p.url == "" && p.conn == nil {
		return newValidationError(fieldURL, messageNoConnection)
	}

	if p.url != "" && p.conn != nil {
		return newValidationError(fieldURL, messageConflictingConnection)
	}

	consumerConfigOptions := 0
	if p.consumerConfig != nil {
		consumerConfigOptions++
	}
	if p.orderedConsumerConfig != nil {
		consumerConfigOptions++
	}

	if consumerConfigOptions > 1 {
		return newValidationError(fieldConsumerConfig, messageMoreThanOneConsumerConfig)
	}

	if len(p.pullConsumeOpts) > 0 && consumerConfigOptions == 0 {
		return newValidationError(fieldPullConsumerOpts, messageReceiverOptionsWithoutConfig)

	}

	if len(p.publishOpts) > 0 && p.sendSubject == "" {
		return newValidationError(fieldPublishOptions, messageSenderOptionsWithoutSubject)
	}

	return nil
}

// validationError is returned when an invalid option is given
type validationError struct {
	field   string
	message string
}

// Error returns a message indicating an error condition, with the nil value representing no error.
func (v validationError) Error() string {
	return fmt.Sprintf("invalid parameters provided: %q: %s", v.field, v.message)
}

// newValidationError creates a validation error
func newValidationError(field, message string) validationError {
	return validationError{
		field:   field,
		message: message,
	}
}
