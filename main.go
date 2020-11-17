package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/magiconair/properties"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

const (
	configPathArg     = "config"
	topicArg          = "topic"
	messageArg        = "message" // json format
	messageKeyArg     = "key"
	messageHeadersArg = "headers"
	versionArg        = "version"
)

var (
	Tag       string
	Commit    string
	SourceURL string
	GoVersion string

	shouldShowVersion   *bool
	configPathValue     string
	topicValue          string
	messageValue        string
	messageKeyValue     string
	messageHeadersValue string
)

func init() {
	flag.StringVar(
		&configPathValue,
		configPathArg,
		"",
		"Config file path. Required parameter.",
	)

	flag.StringVar(
		&topicValue,
		topicArg,
		"",
		"Kafka topic. Required parameter.",
	)

	flag.StringVar(
		&messageValue,
		messageArg,
		"",
		"Kafka message. Required parameter.",
	)

	flag.StringVar(
		&messageKeyValue,
		messageKeyArg,
		"",
		"Message key.",
	)

	flag.StringVar(
		&messageHeadersValue,
		messageHeadersArg,
		"",
		"Message headers.",
	)

	if Tag != "" {
		shouldShowVersion = flag.Bool(
			versionArg,
			false,
			"Show version.",
		)
	}
}

func printUsage() {
	if _, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0]); err != nil {
		log.Fatalf("failed to print usage: %s", err)
	}

	flag.PrintDefaults()
}

func printVersion() {
	fmt.Printf(
		"version: %s\nbuild with: %s\ntag: %s\ncommit: %s\nsource: %s\n",
		strings.TrimPrefix(Tag, "v"),
		GoVersion,
		Tag,
		Commit,
		SourceURL,
	)
}

func main() {
	flag.Parse()

	if shouldShowVersion != nil && *shouldShowVersion {
		printVersion()
		return
	}

	if configPathValue == "" || topicValue == "" || messageValue == "" {
		printUsage()
		os.Exit(1)
	}

	cfgProps, err := readConfig(configPathValue)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	kafkaMessage, err := json.Marshal(messageValue)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	configMap := &kafka.ConfigMap{}
	for k, v := range cfgProps {
		if err := configMap.SetKey(k, v); err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
	}

	producer, err := kafka.NewProducer(configMap)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	defer producer.Close()

	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
					os.Exit(1)
				} else {
					fmt.Printf("Delivered succesfully to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicValue, Partition: kafka.PartitionAny},
		Value:          kafkaMessage,
	}

	if messageKeyValue != "" {
		msg.Key = []byte(messageKeyValue)
	}

	if messageHeadersValue != "" {
		msg.Headers = toKafkaHeaders(messageHeadersValue)
	}

	err = producer.Produce(msg, nil)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	producer.Flush(15 * 1000)
}

func readConfig(configPath string) (map[string]string, error) {
	props, err := properties.LoadFile(configPath, properties.UTF8)
	if err != nil {
		return nil, err
	}

	return props.Map(), nil
}

func toKafkaHeaders(s string) []kafka.Header {
	keyWithValues := strings.Split(s, ",")

	headers := make([]kafka.Header, 0, len(keyWithValues))

	for _, keyWithValue := range keyWithValues {
		kv := strings.Split(keyWithValue, "=")
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])

			if key == "" || value == "" {
				continue
			}

			headers = append(headers, kafka.Header{
				Key:   key,
				Value: []byte(value),
			})
		}
	}

	return headers
}
