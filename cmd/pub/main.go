package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"test/internal/entities"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func OpenJsonFile(name string) (entities.Order, error) {
	// создаем массив с заказами
	orders := entities.Order{}
	// открываем models.json
	file, err := os.Open(name)
	if err != nil {
		return entities.Order{}, err
	}
	// приводим к байтовому представлению
	r, err := io.ReadAll(file)
	if err != nil {
		return entities.Order{}, err
	}
	err = json.Unmarshal(r, &orders)
	if err != nil {
		log.Printf("Error unmarshaling %v\n", err)
	}
	return orders, nil
}

func main() {
	log.Println("Publisher service starting...")
	// подключаемся к натсу
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Error connecting to nats server: %v\n", err)
	}

	// создаем новый поток
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("Error creating jetstream manager interface: %v\n", err)
	}

	// подключаемся к потоку
	_, er := js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name: "TEST_STREAM",
		Subjects: []string{
			"TEST.*"},
	})
	if er != nil {
		log.Fatal(er)
	}

	// открываем model.json
	orders, err := OpenJsonFile("model.json")
	if err != nil {
		log.Fatal(err)
	}
	// постим в поток
	d, err := json.Marshal(orders)
	if err != nil {
		log.Fatalf("Error marhsalling data: %v\n", err)
	}
	js.Publish(context.Background(), "TEST.HELLO", d)
	log.Println("Published a message!")
}
