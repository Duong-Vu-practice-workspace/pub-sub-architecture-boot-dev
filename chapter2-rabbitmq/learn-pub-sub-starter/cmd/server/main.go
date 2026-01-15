package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)
import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connStr := "amqp://guest:guest@localhost:5672/"
	con, err := amqp.Dial(connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to RabbitMQ: %v\n", err)
		os.Exit(1)
	}
	defer con.Close()
	fmt.Println("connected to RabbitMQ")

	publishCh, err := con.Channel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open channel: %v\n", err)
		os.Exit(1)
	}
	defer publishCh.Close()
	err = publishCh.ExchangeDeclare(routing.ExchangePerilDirect, "direct", true, false, false, false, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to declare exchange: %v\n", err)
		os.Exit(1)
	}
	err = publishCh.ExchangeDeclare(routing.ExchangePerilTopic, "topic", true, false, false, false, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to declare exchange: %v\n", err)
		os.Exit(1)
	}

	// Declare dead letter exchange
	err = publishCh.ExchangeDeclare(routing.ExchangePerilDLX, "fanout", true, false, false, false, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to declare dead letter exchange: %v\n", err)
		os.Exit(1)
	}

	// Declare dead letter queue
	_, err = publishCh.QueueDeclare("peril_dlq", true, false, false, false, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to declare dead letter queue: %v\n", err)
		os.Exit(1)
	}

	// Bind dead letter queue to dead letter exchange
	err = publishCh.QueueBind("peril_dlq", "", routing.ExchangePerilDLX, false, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind dead letter queue: %v\n", err)
		os.Exit(1)
	}

	queueName := "game_logs"
	routingKey := "game_logs.*"
	newChannel, _, err := pubsub.DeclareAndBind(con, routing.ExchangePerilTopic, queueName, routingKey, pubsub.SimpleQueueDurable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to declare and bind: %v\n", err)
		os.Exit(1)
	}
	defer newChannel.Close()
	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	gamelogic.PrintServerHelp()
	for {
		// Non-blocking check for Ctrl+C
		select {
		case <-signalChan:
			fmt.Println("exiting")
			return
		default:
		}

		s := gamelogic.GetInput()
		if len(s) == 0 {
			continue
		}
		firstWord := s[0]
		switch firstWord {
		case "pause":
			fmt.Println("sending pause message")
			data := routing.PlayingState{IsPaused: true}
			if err := pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, data); err != nil {
				fmt.Fprintf(os.Stderr, "failed to publish pause: %v\n", err)
			} else {
				fmt.Println("pause published")
			}
		case "resume":
			fmt.Println("sending resume message")
			data := routing.PlayingState{IsPaused: false}
			if err := pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, data); err != nil {
				fmt.Fprintf(os.Stderr, "failed to publish resume: %v\n", err)
			} else {
				fmt.Println("resume published")
			}
		case "quit":
			fmt.Println("exiting")
			return
		default:
			fmt.Printf("unknown command: %s\n", s)
		}
	}
}
