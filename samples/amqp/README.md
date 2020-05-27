# AMQP Samples

These samples require an AMQP 1.0 broker or router to be running.

One option is http://qpid.apache.org/components/dispatch-router/index.html 
It can be installed via dnf or apt, or from source: https://qpid.apache.org/packages.html
Run `qdrouterd` and the samples will work without any additional configuration.

## Sample configuration

The environment variable AMQP_URL can be set to indicate the location of the broker
and the AMQP node address. It has this form:

    amqp://user:pass@host:port/node

The default if AMQP_URL is not set is:

    amqp://localhost:5672/test

*Note*: setting `user:pass` in a URL is not recommended in production,
it is done here to simplify using the samples.


