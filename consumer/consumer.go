package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"

	uuid "github.com/satori/go.uuid"

	utils "github.com/gabrielkim13/comm-infra/utils"
)

type PluginEvent struct {
	ClientId  string `json:"client_id"`
	Code      string `json:"code"`
	Value     int    `json:"value"`
	CreatedAt string `json:"created_at"`
}

func main() {
	args := parseArgs()

	fmt.Printf("comm-infra-consumer: Consumer client for testing Kafka\n\n")
	fmt.Printf("Starting consumer: %s\n\n", *args["consumerGroup"])

	consumer := createKafkaConsumer(*args["topic"], *args["consumerGroup"])

	defer func(consumer *kafka.Consumer) {
		err := consumer.Close()

		utils.FailOnError(err, "Failed to close consumer")
	}(consumer)

	go readMessages(consumer)

	utils.WaitForCtrlC()

	fmt.Printf("\nExiting...")
}

func createKafkaConsumer(topic string, consumerGroup string) *kafka.Consumer {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          consumerGroup,
		"auto.offset.reset": "earliest",
	})

	utils.FailOnError(err, "Failed to create consumer")

	err = consumer.SubscribeTopics([]string{topic}, nil)

	return consumer
}

func readMessages(consumer *kafka.Consumer) {
	for {
		message, err := consumer.ReadMessage(100 * time.Millisecond)

		if err != nil {
			continue
		}

		key := string(message.Key)
		value := message.Value

		pluginEvent := PluginEvent{}
		err = json.Unmarshal(value, &pluginEvent)

		if err != nil {
			fmt.Printf("Failed to decode JSON at offset %d: %v", message.TopicPartition.Offset, err)

			continue
		}

		fmt.Printf("Consumed record with key \"%s\" and value %v\n", key, pluginEvent)
	}
}

func parseArgs() map[string]*string {
	topic := flag.String("t", "", "Topic name")
	consumerGroup := flag.String("g", "", "Consumer group identifier")

	flag.Parse()

	if *topic == "" {
		flag.Usage()

		os.Exit(2)
	}

	if *consumerGroup == "" {
		fmt.Printf("Using random consumer group: %s\n", uuid.NewV4().String())
	}

	return map[string]*string{
		"topic":         topic,
		"consumerGroup": consumerGroup,
	}
}
