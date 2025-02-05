package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable   SimpleQueueType = 0
	Transient SimpleQueueType = 1
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	vals, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        vals,
	})
	if err != nil {
		return err
	}
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	// simpleQuetype:
	// 0 = durable
	// 1 = transient
	durable := true
	autodel := false
	exclusive := false
	if simpleQueueType == Transient {
		durable = false
		autodel = true
		exclusive = true
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	q, err := channel.QueueDeclare(queueName, durable, autodel, exclusive, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return channel, q, nil

}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T),
) error {

	channel, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}
	deliveryChan, err := channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() error {
		for i := range deliveryChan {
			var msg T
			err = json.Unmarshal(i.Body, &msg)
			if err != nil {
				return err
			}
			handler(msg)
			i.Ack(false)
		}
		return nil
	}()
	return nil
}
