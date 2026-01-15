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

	// Configure dead letter exchange
	args := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}

	queue, err := newChannel.QueueDeclare(queueName, isDurable, isAutoDelete, isExclusive, false, args)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = newChannel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return newChannel, queue, nil
}
