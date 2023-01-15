package main

import (
	"go.dedis.ch/cs438/cmd"
	"log"
	"os"
	"strconv"
)

func main() {
	// Enters the command line interface
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) == 0 {
		// Normal node, just initiate one node
		cmd.UserInterface()
		return
	}

	if argsWithoutProg[0] == "simu" {
		// Run in simulation mode
		if len(argsWithoutProg) > 1 {
			nbNodes, err := strconv.Atoi(argsWithoutProg[1])
			if err != nil {
				log.Fatalf("Run the program as `go run .` for normal mode or `go run . simu " +
					"$num_of_nodes` for simulation mode")
			}
			cmd.SimuUserInterface(nbNodes)
		} else {
			defaultNbNodes := 6
			cmd.SimuUserInterface(defaultNbNodes)
		}
		return
	}

	log.Fatalf("Run the program as `go run .` for normal mode or `go run . simu " +
		"$num_of_nodes` for simulation mode")
}
