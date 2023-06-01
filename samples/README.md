# Samples

This directory contains samples for most CloudEvents sdk-go features. 
You can grab them and copy-paste in your project to start using sdk-go.

* AMQP
  * [AMQP Sender](./amqp/sender): Send events using the CloudEvents Client. To run the tests look at [AMQP samples README](./amqp/README.md).
  * [AMQP Receiver](./amqp/receiver): Receive events using the CloudEvents Client. To run the tests look at [AMQP samples README](./amqp/README.md).
* Go channels
  * [Go channels example](./gochan): Send and receive events using the CloudEvents Client. Useful for mocking purpose.
* HTTP:
  * [Receiver](./http/receiver): Receive events using the CloudEvents Client.
  * [Direct receiver](./http/receiver-direct): Create an `http.Handler` to receive events without the CloudEvents Client.
  * [Gorilla receiver](./http/receiver-gorilla): Receive events using [Gorilla](https://www.gorillatoolkit.org/).
  * [Sleepy receiver](./http/receiver-sleepy): Receive events for 5 seconds, then stop the receiver. 
  * [Traced receiver](./http/receiver-traced): Receive events enabling tracing.
  * [Requester](./http/requester): Request/Response events using the CloudEvents Client, creating them with different data content types and different encodings.
  * [Requester with custom client](./http/requester-with-custom-client): Request/Response events with a custom `http.Transport` with TLS configured.
  * [Responder](./http/responder): Receive and reply to events using the CloudEvents Client.
  * [Sender](./http/sender): Send events using the CloudEvents Client.
  * [Sender with retries](./http/sender-retry): Send events, retrying in case of a failure.
  * [Receiver & Requester with metrics enabled](./http/metrics): Request events and handle events with metrics enabled.
* Kafka
  * [Receiver](./kafka/receiver): Receive events using the CloudEvents Client. To run the tests look at [Kafka samples README](./kafka/README.md).
  * [Sender](./kafka/sender): Receive events using the CloudEvents Client. To run the tests look at [Kafka samples README](./kafka/README.md).
  * [Sender & Receiver](./kafka/sender-receiver): Send and receive events using the same Kafka client. To run the tests look at [Kafka samples README](./kafka/README.md).
* Message
  * [Message interoperability](./nats/message-interoperability): Pipe a message from an HTTP receiver directly to NATS using directly the `Protocol`s implementations.
  * [Handle non CloudEvents](./kafka/message-handle-non-cloudevents): Pipe messages from one Kafka topic to another and transform non CloudEvents to valid CloudEvents.
* NATS
  * [Receiver](./nats/receiver): Receive events using the CloudEvents Client.
  * [Sender](./nats/sender): Receive events using the CloudEvents Client.
* STAN
  * [Receiver](./stan/receiver): Receive events using the CloudEvents Client.
  * [Sender](./stan/sender): Receive events using the CloudEvents Client.
  * [Sender & Receiver](./stan/sender-receiver): Send and receive events using the same NATS client.
* WebSockets
  * [Client](./ws/client): Sends and receive events, from client side, using the CloudEvents Client.