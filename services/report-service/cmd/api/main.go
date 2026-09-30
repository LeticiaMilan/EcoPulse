package main

import (
	"log"
	"os"
	"report-service/internal/infraestructure/messaging/amqp"

	_ "github.com/joho/godotenv/autoload"
)

func main() {

	log.Print(os.Getenv("AMQP_URL"))
	QueueManager := amqp.QueueManager{}
	err := QueueManager.NewEventManager((os.Getenv("AMQP_URL")))

	if err != nil {
		log.Fatal(err)
	}

	consumer1 := amqp.Consumer{}
	consumer1.NewConsumer(&QueueManager, "teste")

	consumer1.Start()
}
