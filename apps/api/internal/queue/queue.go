package queue

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Job struct {
	Type string         `json:"type"`
	ID   string         `json:"id"`
	Data map[string]any `json:"data,omitempty"`
}

type Queue interface {
	Enqueue(ctx context.Context, job Job) error
}

type RedisQueue struct {
	client *redis.Client
	name   string
	log    *slog.Logger
}

func NewRedisQueue(client *redis.Client, name string, log *slog.Logger) *RedisQueue {
	return &RedisQueue{client: client, name: name, log: log}
}

func (q *RedisQueue) Enqueue(ctx context.Context, job Job) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	q.log.Info("enqueue job", "type", job.Type, "id", job.ID)
	return q.client.LPush(ctx, q.name, payload).Err()
}
