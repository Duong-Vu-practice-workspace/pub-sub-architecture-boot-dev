package pubsub

import amqp "github.com/rabbitmq/amqp091-go"

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	newChannel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	isDurable := false
	isAutoDelete := true
	isExclusive := true
	if queueType == SimpleQueueDurable {
		isDurable = true
		isAutoDelete = false
		isExclusive = false
	}
	defer newChannel.Close()
	queue, err := newChannel.QueueDeclare(queueName, isDurable, isAutoDelete, isExclusive, false, nil)
	if err != nil {
		return newChannel, amqp.Queue{}, err
	}
	err = newChannel.QueueBind(queueName, exchange, key, false, nil)
	return newChannel, queue, nil
}
