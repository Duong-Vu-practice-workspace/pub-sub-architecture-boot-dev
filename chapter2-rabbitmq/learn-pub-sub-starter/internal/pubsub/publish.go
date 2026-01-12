package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var b []byte
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	publishing := amqp.Publishing{
		ContentType: "application/json",
		Body:        b,
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, publishing)

}
