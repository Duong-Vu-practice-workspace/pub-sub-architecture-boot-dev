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
	newChannel, err := con.Channel()
	defer newChannel.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create channel: %v\n", err)
		os.Exit(1)
	}

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
			if err := pubsub.PublishJSON(newChannel, routing.ExchangePerilDirect, routing.PauseKey, data); err != nil {
				fmt.Fprintf(os.Stderr, "failed to publish pause: %v\n", err)
			} else {
				fmt.Println("pause published")
			}
		case "resume":
			fmt.Println("sending resume message")
			data := routing.PlayingState{IsPaused: false}
			if err := pubsub.PublishJSON(newChannel, routing.ExchangePerilDirect, routing.PauseKey, data); err != nil {
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
