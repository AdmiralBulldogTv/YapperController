package instance

import (
	"context"
	"time"
)

type Redis interface {
	Ping(ctx context.Context) error
	Subscribe(ctx context.Context, ch chan string, subscribeTo ...string)
	Publish(ctx context.Context, channel string, data string) error
	SAdd(ctx context.Context, set string, values ...interface{}) error
	Set(ctx context.Context, key string, value string, expiry time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}
