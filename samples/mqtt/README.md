MQTT samples

To run the samples, you need a running MQTT broker.

To run a sample MQTT broker using docker:

```bash
echo "listener 1883 0.0.0.0
allow_anonymous true" > samples/mosquitto.conf
docker run --rm --name mosquitto -p 1883:1883 -v "$(pwd)/samples/mosquitto.conf:/mosquitto/config/mosquitto.conf" eclipse-mosquitto
```