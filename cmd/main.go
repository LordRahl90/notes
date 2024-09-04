package main

import (
	"log"
	"log/slog"
	"os"

	"notes/server"
)

func main() {
	slog.Info("Target is", "Target", os.Getenv("TARGET"))
	svr := server.New()
	if err := svr.Start(":8181"); err != nil {
		log.Fatal(err)
	}
}
