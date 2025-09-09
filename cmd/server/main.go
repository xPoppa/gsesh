package main

import (
	"log"
	"os"

	"github.com/xPoppa/gsesh/server"
)

func main() {
	err := server.Run()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
