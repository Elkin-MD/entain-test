package main

import (
	"log"
	"os"

	"entaintest/cmd/migrations"
	"entaintest/cmd/server"
)

func main() {
	command := "server"
	args := os.Args[1:]

	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	var err error

	switch command {
	case "server":
		err = server.Run()
	case "migrate":
		err = migrations.Run(args)
	default:
		log.Fatalf("unknown command %q (want server or migrate)", command)
	}

	if err != nil {
		log.Fatal(err)
	}
}
