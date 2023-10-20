#!/bin/bash

# Run the services needed by the integration test in our local docker install

if [[ "$1" == "stop" ]]; then
	docker rm -f kafka nats amqp mqtt
	exit 0
fi

# Kafka
docker run --name kafka -dti -e ADV_HOST=localhost -p 9091:9091 -p 9092:9092 \
	lensesio/fast-data-dev

# NATS
docker run --name nats -dti -p 4222:4222 nats-streaming:0.22.1

# AMQP
docker run --name amqp -dti -e QDROUTERD_CONFIG_OPTIONS='
  router {
    mode: standalone
    id: ZTg2NDQ0N2Q1YjU1OGE1N2NkNzY4NDFk
    workerThreads: 4
  }
  log {
    module: DEFAULT
    enable: trace+
    timestamp: true
  }
  listener {
    role: normal
    host: 0.0.0.0
    port: amqp
    saslMechanisms: ANONYMOUS
  }' -p 5672:5672 scholzj/qpid-dispatch

# MQTT
docker run --name mqtt -dti -p 1883:1883 eclipse-mosquitto:1.6

