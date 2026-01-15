package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel, username string) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			// Publish war recognition
			war := gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.GetPlayerSnap(),
			}
			routingKey := fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, username)
			err := pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, routingKey, war)
			if err != nil {
				fmt.Printf("error publishing war: %v\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState, publishCh *amqp.Channel, username string) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon, gamelogic.WarOutcomeDraw:
			var msg string
			if outcome == gamelogic.WarOutcomeDraw {
				msg = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			} else {
				msg = fmt.Sprintf("%s won a war against %s", winner, loser)
			}
			err := publishGameLog(publishCh, rw.Attacker.Username, msg, username)
			if err != nil {
				fmt.Printf("error publishing game log: %v\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("error: unknown war outcome")
			return pubsub.NackDiscard
		}
	}
}

func publishGameLog(publishCh *amqp.Channel, attackerUsername, msg, clientUsername string) error {
	return pubsub.PublishGob(
		publishCh,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.GameLogSlug, attackerUsername),
		routing.GameLog{
			CurrentTime: time.Now(),
			Message:     msg,
			Username:    clientUsername,
		},
	)
}
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
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get username: %v\n", err)
		os.Exit(1)
	}

	// Declare the exchanges
	ch, err := con.Channel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open channel: %v\n", err)
		os.Exit(1)
	}
	err = ch.ExchangeDeclare(routing.ExchangePerilDirect, "direct", true, false, false, false, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to declare exchange: %v\n", err)
		os.Exit(1)
	}
	err = ch.ExchangeDeclare(routing.ExchangePerilTopic, "topic", true, false, false, false, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to declare exchange: %v\n", err)
		os.Exit(1)
	}
	ch.Close()

	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		con,
		routing.ExchangePerilDirect,
		fmt.Sprintf("pause.%s", username),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)
	if err != nil {
		fmt.Printf("failed to subscribe to pause messages: %v\n", err)
	}

	// Create a channel for publishing moves
	publishCh, err := con.Channel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open publish channel: %v\n", err)
		os.Exit(1)
	}
	defer publishCh.Close()

	// Subscribe to army moves
	err = pubsub.SubscribeJSON(
		con,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.SimpleQueueTransient,
		handlerMove(gameState, publishCh, username),
	)
	if err != nil {
		fmt.Printf("failed to subscribe to army moves: %v\n", err)
	}

	// Subscribe to war declarations
	err = pubsub.SubscribeJSON(
		con,
		routing.ExchangePerilTopic,
		"war",
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.SimpleQueueDurable,
		handlerWar(gameState, publishCh, username),
	)
	if err != nil {
		fmt.Printf("failed to subscribe to war: %v\n", err)
	}
	// wait for ctrl+c
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	for {
		select {
		case <-sigCh:
			fmt.Println("shutdown signal received, exiting")
			return
		default:
		}
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err := gameState.CommandSpawn(words)
			if err != nil {
				fmt.Println(err.Error())
			}
		case "move":
			move, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				// Publish the move
				routingKey := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username)
				err = pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, routingKey, move)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to publish move: %v\n", err)
				} else {
					fmt.Println("move published successfully")
				}
			}
		case "status":
			gameState.CommandStatus()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("invalid command")
		}
	}

}
