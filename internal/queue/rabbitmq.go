package queue

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

func Connect(url string, exchangeName string) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	err = ch.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	return conn, ch
}

func Publish(ch *amqp.Channel, exchangeName string, routingKey string, msg any) {
	body, _ := json.Marshal(msg)

	err := ch.Publish(
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Println("publish failed:", err)
	}
}
