package main

import (
	"fmt"
	"log"

	gamelogic "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	pubsub "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const rabbitMqURL = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril server...")
	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println(err)
	}

	rabbitConn, err := amqp.Dial(rabbitMqURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()

	fmt.Println("connection succesful")

	PauseQue := routing.PauseKey + "." + userName
	_, _, err = pubsub.DeclareAndBind(rabbitConn, routing.ExchangePerilDirect, PauseQue, routing.PauseKey, pubsub.Transient)
	if err != nil {
		fmt.Println(err)
	}
	gamestate := gamelogic.NewGameState(userName)
	pubsub.SubscribeJSON(rabbitConn, routing.ExchangePerilDirect, PauseQue, routing.PauseKey, pubsub.Transient, handlerPause(gamestate))
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		cmd := input[0]
		switch cmd {
		case "spawn":
			err = gamestate.CommandSpawn(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			_, err := gamestate.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "status":
			gamestate.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
