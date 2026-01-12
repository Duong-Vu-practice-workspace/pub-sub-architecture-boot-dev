package main

import (
	"fmt"
	"os"
	"os/signal"

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
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create channel: %v\n", err)
	}
	data := routing.PlayingState{IsPaused: true}
	err = pubsub.PublishJSON(newChannel, routing.ExchangePerilDirect, routing.PauseKey, data)
	if err != nil {
		fmt.Errorf(err.Error())
	}
	fmt.Println("sending success")
	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
