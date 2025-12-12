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

	err := cgroup.Unshare()
	if err != nil {
		return err
	}

	cg, err := cgroup.ReadCgroup()
	if err != nil {
		return err
	}

	if err := cg.Mount(); err != nil {
		return err
	}

	msg, err := nsjail.GetMsg(cfg)
	if err != nil {
		return err
	}

	if err = cg.SetConfig(msg); err != nil {
		return err
	}

	if err = nsjail.Write(cfg.NsjailCfgPath, msg); err != nil {
		return err
	}

	manager := manager.NewGameManager(cfg)

	// TODO: Create RabbitMQ Consumer (max 20 or so concurrency)
	// TODO: Pass incoming matches to game manager
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
