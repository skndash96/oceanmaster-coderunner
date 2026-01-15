package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/delta/code-runner/internal/cgroup"
	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/manager"
	"github.com/delta/code-runner/internal/nsjail"
	"github.com/delta/code-runner/internal/queue"
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

	_, ch := queue.Connect(cfg.RabbitMQURL, cfg.ExchangeName)

	q, err := ch.QueueDeclare(
		"match_jobs",
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(q.Name, "match.jobs", cfg.ExchangeName, false, nil)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	resultsQueue := make(chan queue.MatchResult, 100)

	go func() {
		for result := range resultsQueue {
			log.Println("ENGINE SENDING RESULT:", result)
			queue.Publish(ch, cfg.ExchangeName, "match.results", result)
		}
	}()

	sem := make(chan struct{}, cfg.MaxConcurrentMatches)

	log.Printf("Runner started. Waiting for match jobs (max concurrent: %d)...\n", cfg.MaxConcurrentMatches)

	for d := range msgs {
		var job queue.MatchJob
		if err := json.Unmarshal(d.Body, &job); err != nil {
			log.Printf("Error unmarshaling job: %v\n", err)
			continue
		}

		if job.Priority == 0 {
			job.Priority = 100
		}

		log.Printf("ENGINE RECEIVED JOB: ID=%s P1=%s P2=%s Priority=%d\n", job.ID, job.P1, job.P2, job.Priority)

		sem <- struct{}{}

		go func(job queue.MatchJob) {
			defer func() { <-sem }()

			if err := manager.NewMatch(
				job.ID,
				job.P1,
				job.P2,
				job.P1Code,
				job.P2Code,
			); err != nil {
				// TODO: Add status to MatchResult incase game ends with error
				log.Printf("Failed to process match %s: %v\n", job.ID, err)
				resultsQueue <- queue.MatchResult{
					ID:     job.ID,
					Winner: -1,
				}
			} else {
				// TODO: Get actual winner from match result
				resultsQueue <- queue.MatchResult{
					ID:     job.ID,
					Winner: 0,
				}
			}
		}(job)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
