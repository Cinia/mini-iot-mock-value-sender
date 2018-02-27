# mini-iot-mock-value-sender

Mock data-source to be used in place of https://github.com/Cinia/mini-iot/tree/master/rpi-sense-hat

## Running

To run program needs one argument
* ```broker``` - connection string to MQTT-broker

Argument can be provided as command line flags, for example:

    ./mini-iot-mock-value-sender -broker=tcp:/localhost:1883

or environment variable:

    BROKER=tcp://localhost:1883 ./mini-iot-mock-value-sender

## Running with Docker

    docker run -e BROKER="tcp://mqtt-broker:1883" cinia/mini-iot-mock-value-sender