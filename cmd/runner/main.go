package main

import (
	"encoding/json"
	"log"

	"github.com/delta/code-runner/internal/cgroup"
	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/manager"
	"github.com/delta/code-runner/internal/nsjail"
	"github.com/delta/code-runner/internal/queue"
	"github.com/rabbitmq/amqp091-go"
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

	gameManager := manager.NewGameManager(cfg)

	matchJobQ, err := queue.NewMatchJobQueue(cfg.MatchJobQueueConfig)
	if err != nil {
		return err
	}
	defer matchJobQ.Close()

	err = matchJobQ.SetMaxConcurrentMatches(cfg.MaxConcurrentMatches)
	if err != nil {
		return err
	}

	msgs, err := matchJobQ.Consume()
	if err != nil {
		return err
	}

	sem := make(chan struct{}, cfg.MaxConcurrentMatches)

	log.Printf("consumer started with %d max_concurrency)\n", cfg.MaxConcurrentMatches)

	for d := range msgs {
		sem <- struct{}{} // blocks if max concurrency reached

		go func(delivery amqp091.Delivery) {
			defer func() { <-sem }()

			var job manager.MatchJob
			if err := json.Unmarshal(delivery.Body, &job); err != nil {
				log.Println("INVALID JOB:", err)
				delivery.Nack(false, false) // discard
				return
			}

			log.Println("RUNNING MATCH:", job.ID)

			if err := gameManager.NewMatch(job); err != nil {
				log.Println("MATCH FAILED:", err)
				delivery.Nack(false, false) // discard
				// TODO: Try atleast twice and then discard
				return
			}

			log.Println("MATCH FINISHED:", job.ID)
			delivery.Ack(false)
		}(d)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
