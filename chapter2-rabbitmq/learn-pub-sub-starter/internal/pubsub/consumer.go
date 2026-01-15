package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	deli, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for delivery := range deli {
			var message T
			err := json.Unmarshal(delivery.Body, &message)
			if err != nil {
				fmt.Printf("failed to unmarshal message: %v\n", err)
				continue
			}
			ackType := handler(message)
			switch ackType {
			case Ack:
				delivery.Ack(false)
				fmt.Println("Ack: message acknowledged")
			case NackRequeue:
				delivery.Nack(false, true)
				fmt.Println("NackRequeue: message requeued")
			case NackDiscard:
				delivery.Nack(false, false)
				fmt.Println("NackDiscard: message discarded")
			}
		}
	}()
	return nil
}
