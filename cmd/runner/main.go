package main

import (
	"fmt"
	"log"

	"github.com/delta/code-runner/internal/cgroup"
	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/nsjail"
	"github.com/delta/code-runner/internal/simulator"
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

	if err = nsjail.Write(cfg, msg); err != nil {
		return err
	}

	// TODO: Create RabbitMQ Consumer
	// TODO: Pass incoming matches to simulator

	// example test
	// this wud be started in a goroutine
	code := `
import numpy
import random

def x():
	print("Hello, World!")
	print(numpy.random.rand(3, 3))
	print(random.random())
	`
	err = simulator.Simulate(cfg, simulator.NewMatch("12", "player1", "player2", code, code))
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
