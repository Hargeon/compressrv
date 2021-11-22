// Package broker uses for implementation message broker
package broker

import (
	"fmt"

	"github.com/streadway/amqp"
)

// Rabbit represent rabbitmq client
type Rabbit struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue

	user     string
	password string
	host     string
	port     string
}

// NewRabbit initialize Rabbit
func NewRabbit(user, password, host, port string) *Rabbit {
	return &Rabbit{
		user:     user,
		password: password,
		host:     host,
		port:     port,
	}
}

// Connect to rabbit, create chan, declare queue
func (r *Rabbit) Connect(queueName string) (*amqp.Connection, error) {
	rabbitDsn := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		r.user, r.password, r.host, r.port)
	conn, err := amqp.Dial(rabbitDsn)

	if err != nil {
		return nil, err
	}

	r.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	r.ch = ch

	q, err := ch.QueueDeclare(
		queueName,
		true, // message will not lose if rabbit crashed
		false,
		false,
		false,
		nil)
	if err != nil {
		return nil, err
	}

	r.q = q

	return conn, err
}

// Publish body to rabbit
func (r *Rabbit) Publish(body []byte) error {
	return r.ch.Publish(
		"",
		r.q.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // message will not lose if rabbit crashed
			ContentType:  "application/json",
			Body:         body,
		})
}

func (r *Rabbit) Consume() (<-chan amqp.Delivery, error) {
	msgs, err := r.ch.Consume(
		r.q.Name,
		"",
		false, // needs to mark a message was processed
		false,
		false,
		false,
		nil)

	return msgs, err
}
