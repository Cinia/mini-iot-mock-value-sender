package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

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
		for {
			select {
			case <-ctx.Done():
				log.Info("Context cancelled, quitting sender thread")
				return
			default:
				log.Info("Sending message..")
				if token := client.Publish("mini-iot/mock-sender/temperature", 0, false, "25"); token.Wait() && token.Error() != nil {
					log.Fatal(token.Error())
				}
			}
		}
	}(ctx)

	wg.Wait()
}
