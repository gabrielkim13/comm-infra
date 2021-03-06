package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	uuid "github.com/satori/go.uuid"

	utils "github.com/gabrielkim13/comm-infra/utils"
)

type PluginEvent struct {
	ClientId  string `json:"client_id"`
	Code      string `json:"code"`
	Value     int    `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

type PublishPluginEventsOptions struct {
	Code     string
	Min      int
	Max      int
	Interval time.Duration
}

var ClientId, Username, Password string

func init() {
	rand.Seed(time.Now().UnixNano())

	envUsername, isUsernameDefined := os.LookupEnv("COMM_INFRA_PRODUCER_USERNAME")
	envPassword, isPasswordDefined := os.LookupEnv("COMM_INFRA_PRODUCER_PASSWORD")

	if isUsernameDefined && isPasswordDefined {
		Username = envUsername
		Password = envPassword

		ClientId = Username
	} else {
		Username = "guest"
		Password = "guest"

		ClientId = uuid.NewV4().String()
	}
}

func main() {
	fmt.Printf("comm-infra-producer: MQTT client for testing RabbitMQ\n\n")
	fmt.Printf("Starting producer: %s\n\n", ClientId)

	client := createMqttClient("localhost", 1883)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		utils.FailOnError(token.Error(), "Failed to connect to broker")
	}

	go publishPluginEvents(client, &PublishPluginEventsOptions{
		Code:     "0001",
		Min:      0,
		Max:      1023,
		Interval: 10 * time.Second,
	})

	go publishPluginEvents(client, &PublishPluginEventsOptions{
		Code:     "0002",
		Min:      127,
		Max:      255,
		Interval: 20 * time.Second,
	})

	go publishPluginEvents(client, &PublishPluginEventsOptions{
		Code:     "0003",
		Min:      -32768,
		Max:      32767,
		Interval: 30 * time.Second,
	})

	utils.WaitForCtrlC()

	fmt.Printf("\nExiting...")
}

func createMqttClient(host string, port uint16) mqtt.Client {
	broker := fmt.Sprintf("tcp://%s:%d", host, port)

	options := mqtt.NewClientOptions()

	options.AddBroker(broker)
	options.SetClientID(ClientId)
	options.SetUsername(Username)
	options.SetPassword(Password)

	options.OnConnect = func(_ mqtt.Client) {
		fmt.Printf("Connected\n\n")
	}

	options.OnConnectionLost = func(_ mqtt.Client, err error) {
		fmt.Printf("Connection lost: %s\n\n", err)
	}

	client := mqtt.NewClient(options)

	return client
}

func publishPluginEvents(client mqtt.Client, options *PublishPluginEventsOptions) {
	topic := fmt.Sprintf("plugins/%s", options.Code)

	for {
		value := utils.GetRandomIntRange(options.Min, options.Max)

		event := PluginEvent{
			ClientId:  ClientId,
			Code:      options.Code,
			Value:     value,
			Timestamp: time.Now().UnixMilli(),
		}
		payload := fmt.Sprintf("%s,%s,%d,%d", event.ClientId, event.Code, event.Value, event.Timestamp)

		token := client.Publish(topic, 1, false, payload)
		token.Wait()

		fmt.Printf("Published event %v\n", event)

		time.Sleep(options.Interval)
	}
}
