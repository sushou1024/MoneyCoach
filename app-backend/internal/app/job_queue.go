package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const jobQueueKey = "jobs:queue"

type jobQueue struct {
	redis *redisStore
}

type jobPayload struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func newJobQueue(redis *redisStore) *jobQueue {
	return &jobQueue{redis: redis}
}

func (q *jobQueue) enqueue(ctx context.Context, jobType, id string) error {
	payload := jobPayload{Type: jobType, ID: id}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	return q.redis.client.RPush(ctx, jobQueueKey, encoded).Err()
}

func (q *jobQueue) startWorkers(ctx context.Context, server *Server) {
	workerCount := 2
	for i := 0; i < workerCount; i++ {
		go q.workerLoop(ctx, server)
	}
}

func (q *jobQueue) workerLoop(ctx context.Context, server *Server) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result, err := q.redis.client.BLPop(ctx, 5*time.Second, jobQueueKey).Result()
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				continue
			}
			if err.Error() == "redis: nil" {
				continue
			}
			server.logger.Printf("job queue error: %v", err)
			continue
		}
		if len(result) < 2 {
			continue
		}
		payload := result[1]
		var job jobPayload
		if err := json.Unmarshal([]byte(payload), &job); err != nil {
			server.logger.Printf("job decode error: %v", err)
			continue
		}
		start := time.Now()
		server.logger.Printf("job start type=%s id=%s", job.Type, job.ID)
		if err := server.processJob(ctx, job); err != nil {
			server.logger.Printf("job done type=%s id=%s status=error duration=%s err=%v", job.Type, job.ID, time.Since(start), err)
			continue
		}
		server.logger.Printf("job done type=%s id=%s status=ok duration=%s", job.Type, job.ID, time.Since(start))
	}
}
