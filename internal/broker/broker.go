package broker

import (
	"context"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func CreateNewConsumer(connect *nats.Conn) (jetstream.Consumer, error) {
	// создаем поток
	js, err := jetstream.New(connect)
	if err != nil {
		log.Printf("Error creating jetstream manager interface: %v\n", err)
		return nil, err
	}
	stream, err := js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name: "TEST_STREAM",
		Subjects: []string{
			"TEST.*"}})
	if err != nil {
		log.Printf("Error creating stream: %v\n", err)
		return nil, err
	}
	// создаем потребителя
	consumer, err := stream.CreateOrUpdateConsumer(context.Background(), jetstream.ConsumerConfig{
		Durable:   "TestConsumerConsume",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		log.Printf("Error creating consumer: %v\n", err)
		return nil, err
	}
	return consumer, nil
}
