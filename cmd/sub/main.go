package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"test/internal/broker"
	"test/internal/entities"
	"test/internal/handler"
	"test/internal/storage/cache"
	"test/internal/storage/postgres"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	log.Println("Consumer service starting...")

	// создаем кэш в памяти
	var inMemoryCache = cache.New()

	// подключаемся к натсу
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Error connecting to	nats server: %v\n", err)
	}
	log.Println("Connection to nats server successful")

	// создаем нового консьюмера
	consumer, err := broker.CreateNewConsumer(nc)
	if err != nil {
		log.Fatalf("Error while creating consumer: %v\n", err)
	}
	log.Println("New consumer created successfully")

	// подключаемся к постгресу
	pgdb, err := postgres.CreateDB()
	if err != nil {
		log.Fatalf("Error while connecting to postgres: %v\n", err)
	}

	// проверяем соединение с постгресом
	var status string
	pgdb.Db.QueryRow(context.Background(), "select 'Postgres connection established'").Scan(&status)
	log.Println(status)

	// восстанавливаем кэш из постгреса
	go func() {
		status := inMemoryCache.RecoverFromPostgres(pgdb)
		if !status {
			log.Fatal("Cache recovery failed")
		}
	}()

	// обрабатываем сообщения из потока
	cc, err := consumer.Consume(func(msg jetstream.Msg) {
		// подтверждаем сообщение
		msg.Ack()
		// создаем структуру заказ
		order := entities.Order{}
		// обрабатываем json
		err := json.Unmarshal(msg.Data(), &order)
		if err != nil {
			log.Printf("Error while parsing JSON. Might be unsupported type of information : %v\n", err)
			return
		}
		// записываем данные в кэш
		ok := inMemoryCache.Add(order)
		if !ok {
			return
		}

		err = pgdb.WriteData(order)
		if err != nil {
			log.Printf("Error writing data to postgres: %v\n", err)
			return
		}
		log.Printf("New order %s added.\n", order.OrderUID)
	})
	if err != nil {
		log.Fatalf("Error while consuming: %v\n", err)
	}
	defer cc.Stop()

	generalHandler := handler.NewHandler(inMemoryCache)

	// обработчик для получения заказа по id
	http.HandleFunc("/id/{id}", generalHandler.GetByID)

	// запускаем сервис
	serviceError := http.ListenAndServe(":8080", nil)
	if serviceError != nil {
		log.Fatalf("%v\n", serviceError)
	}
}
