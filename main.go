package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type sensor struct {
	temperature float32
	humidity    float32
}

func main() {
	// allow setting values also with commandline flags
	pflag.String("broker", "tcp://localhost:1883", "MQTT-broker url")
	viper.BindPFlag("broker", pflag.Lookup("broker"))

	// parse values from environment variables
	viper.AutomaticEnv()

	brokerURL := viper.GetString("broker")

	log.Infof("Using broker %s", brokerURL)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID("mock-rpi-sender")

	var wg sync.WaitGroup

	// Handle interrupt and term signals gracefully
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		wg.Add(1)
		select {
		case <-sigs:
			cancel()
		case <-ctx.Done():
		}
		wg.Done()
	}()

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	wg.Add(1)
	go func(ctx context.Context) {
		log.Info("Starting the sender thread")
		defer wg.Done()

		// generate intial values for sensors
		s := createRandomSensorValues()
		for {
			select {
			case <-ctx.Done():
				log.Info("Context cancelled, quitting sender thread")
				return
			default:
				s = changeSensorValues(s)
				log.Infof("Publishing temperature %.2f", s.temperature)
				if token := client.Publish("mini-iot/mock-sender/temperature", 0, false, strconv.FormatFloat(float64(s.temperature), 'f', 2, 32)); token.Wait() && token.Error() != nil {
					log.Fatal(token.Error())
				}
				log.Infof("Publishing humidity %.2f", s.humidity)
				if token := client.Publish("mini-iot/mock-sender/humidity", 0, false, strconv.FormatFloat(float64(s.humidity), 'f', 2, 32)); token.Wait() && token.Error() != nil {
					log.Fatal(token.Error())
				}
			}
			time.Sleep(5 * time.Second)
		}
	}(ctx)

	wg.Wait()
}

func createRandomSensorValues() sensor {
	return sensor{
		temperature: ((rand.Float32() * 20) + 15), // value between 15 and 35
		humidity:    ((rand.Float32() * 90) + 10), // value between 10 and 100
	}
}

func changeSensorValues(originalSensor sensor) sensor {
	// change previous value by a value between -0.5 - 0.5
	// and make sure it does not go over boundaries
	originalSensor.temperature += (rand.Float32() - 0.5)
	if originalSensor.temperature < 15 {
		originalSensor.temperature++
	} else if originalSensor.temperature > 35 {
		originalSensor.temperature--
	}

	originalSensor.humidity += (rand.Float32() - 0.5)
	if originalSensor.humidity < 10 {
		originalSensor.humidity++
	} else if originalSensor.humidity > 100 {
		originalSensor.humidity--
	}
	return originalSensor
}
