# Kafka samples

To run the samples, you need a running Kafka cluster.

To run a sample Kafka cluster using docker:

```
docker run --rm --net=host -e ADV_HOST=localhost -e SAMPLEDATA=0 lensesio/fast-data-dev
```