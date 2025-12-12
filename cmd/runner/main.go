package main

import (
	"fmt"
	"log"

	"github.com/delta/code-runner/internal/cgroup"
	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/manager"
	"github.com/delta/code-runner/internal/nsjail"
)

func run() error {
	cfg := config.New()

	cg, err := cgroup.UnshareAndMount()
	if err != nil {
		return err
	}

	_, err = nsjail.WriteConfig(cfg, cg)
	if err != nil {
		return err
	}

	manager := manager.NewGameManager(cfg)

	// TODO: Create RabbitMQ Consumer (max 20 or so concurrency)
	// TODO: Pass incoming matches to game manager
	// Make sure new match is called in a goroutine
	err = manager.NewMatch("12", "player1", "player2", egCode, egCode)

	if err != nil {
		return fmt.Errorf("Failed to simulate match %v", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
