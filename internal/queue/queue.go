package queue

import (
	"log"

	"github.com/delta/code-runner/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

type MatchJobQueue struct {
	cfg  config.MatchJobQueueConfig
	conn *amqp091.Connection
	ch   *amqp091.Channel
}

// broker side back pressure
func (q *MatchJobQueue) SetMaxConcurrentMatches(max int) error {
	return q.ch.Qos(max, 0, false)
}

func (q *MatchJobQueue) Consume() (<-chan amqp091.Delivery, error) {
	msgs, err := q.ch.Consume(
		q.cfg.QueueName,
		"",
		false, // no manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func NewMatchJobQueue(cfg config.MatchJobQueueConfig) (*MatchJobQueue, error) {
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	err = ch.ExchangeDeclare(
		cfg.ExchangeName,
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

	q, err := ch.QueueDeclare(
		cfg.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(
		q.Name,
		cfg.RoutingKey,
		cfg.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &MatchJobQueue{
		cfg:  cfg,
		conn: conn,
		ch:   ch,
	}, nil
}

func (q *MatchJobQueue) Close() error {
	err := q.ch.Close()
	if err != nil {
		return err
	}
	err = q.conn.Close()
	if err != nil {
		return err
	}
	return nil
}
