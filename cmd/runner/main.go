package main

import (
	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/nsjail"
	"github.com/delta/code-runner/internal/simulator"
)

func main() {
	cfg := config.New()

	// Write nsjail configuration to file
	_ = nsjail.Setup(cfg)

	// TODO: Create RabbitMQ Consumer
	// TODO: Pass incoming matches to simulator

	// example test
	simulator.Simulate(cfg, simulator.NewMatch("12", "player1", "player2", "code1", "code2"))
}
