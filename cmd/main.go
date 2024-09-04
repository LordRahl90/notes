package main

import (
	"log"

	"notes/server"
)

func main() {
	svr := server.New()
	if err := svr.Start(":8181"); err != nil {
		log.Fatal(err)
	}
}
