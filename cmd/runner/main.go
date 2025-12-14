package main

import (
	"fmt"
	"log"
	"sync"

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

	var (
		mCnt int = 5
		wg   sync.WaitGroup
	)

	wg.Add(mCnt)

	for i := range mCnt {
		go func() {
			defer wg.Done()

			if err := manager.NewMatch(
				fmt.Sprintf("%d", 2*i),
				fmt.Sprintf("player%d", 2*i),
				fmt.Sprintf("player%d", 2*i+1),
				egCode,
				egCode,
			); err != nil {
				fmt.Printf("Failed to simulate match %d: %v\n", 2*i, err)
			}
		}()
	}

	wg.Wait()

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
