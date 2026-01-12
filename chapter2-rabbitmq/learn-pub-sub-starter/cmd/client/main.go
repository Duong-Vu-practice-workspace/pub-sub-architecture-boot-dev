package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	connStr := "amqp://guest:guest@localhost:5672/"
	con, err := amqp.Dial(connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to RabbitMQ: %v\n", err)
		os.Exit(1)
	}
	defer con.Close()
	fmt.Println("connected to RabbitMQ")
	username, err := gamelogic.ClientWelcome()
	queueName := fmt.Sprintf("pause.%s", username)
	newChannel, queue, err := pubsub.DeclareAndBind(con, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.SimpleQueueTransient)
	defer newChannel.Close()
	fmt.Printf("declared transient queue %s\n", queue.Name)
	// wait for ctrl+c
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	fmt.Println("shutdown signal received, exiting")

}
