package helpers

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func ConnectToRabbitMQ() *amqp.Connection {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to the RabbitMQ")
	return conn
}
