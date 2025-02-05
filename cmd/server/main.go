package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	gamelogic "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	pubsub "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const rabbitMqURL = "amqp://guest:guest@localhost:5672/"

func main() {
	gamelogic.PrintServerHelp()

	rabbitConn, err := amqp.Dial(rabbitMqURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()

	fmt.Println("connection succesful")

	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = pubsub.DeclareAndBind(rabbitConn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.Durable)
	if err != nil {
		fmt.Println(err)
	}
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		if input[0] == "pause" {
			fmt.Println("sending pause message")
			pubsub.PublishJSON(rabbitChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			continue
		} else if input[0] == "resume" {
			fmt.Println("sending resume message")
			pubsub.PublishJSON(rabbitChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			continue
		} else if input[0] == "quit" {
			fmt.Println("exiting..")
			break
		} else {
			fmt.Println("not valid command")
			continue
		}
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	log.Println("Program shutting down, closing connection")

	os.Exit(0)

}
