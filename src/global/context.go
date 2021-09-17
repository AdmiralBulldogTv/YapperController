package global

import (
	"context"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/instances"
)

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
	Config() ServerCfg
	SetMongoInstance(instance instances.MongoInstance)
	GetMongoInstance() instances.MongoInstance
	SetRedisInstance(instance instances.RedisInstance)
	GetRedisInstance() instances.RedisInstance
	SetTtsInstance(instance instances.TtsInstance)
	GetTtsInstance() instances.TtsInstance
}

type gCtx struct {
	ctx context.Context
	cfg ServerCfg
}

func (c *gCtx) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

func (c *gCtx) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *gCtx) Err() error {
	return c.ctx.Err()
}

func (c *gCtx) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c *gCtx) Config() ServerCfg {
	return c.cfg
}

func (c *gCtx) SetMongoInstance(instance instances.MongoInstance) {
	c.cfg.mongo = instance
}

func (c *gCtx) GetMongoInstance() instances.MongoInstance {
	return c.cfg.mongo
}

func (c *gCtx) SetRedisInstance(instance instances.RedisInstance) {
	c.cfg.redis = instance
}

func (c *gCtx) GetRedisInstance() instances.RedisInstance {
	return c.cfg.redis
}

func (c *gCtx) SetTtsInstance(instance instances.TtsInstance) {
	c.cfg.tts = instance
}

func (c *gCtx) GetTtsInstance() instances.TtsInstance {
	return c.cfg.tts
}

func NewCtx(ctx context.Context, config ServerCfg) Context {
	return &gCtx{ctx: ctx, cfg: config}
}

func WithValue(ctx Context, key interface{}, value interface{}) Context {
	cfg := ctx.Config()
	return NewCtx(context.WithValue(ctx, key, value), cfg)
}

func WithCancel(ctx Context) (Context, context.CancelFunc) {
	cfg := ctx.Config()
	nCtx, cancel := context.WithCancel(ctx)
	return NewCtx(nCtx, cfg), cancel
}

func WithDeadline(ctx Context, deadline time.Time) (Context, context.CancelFunc) {
	cfg := ctx.Config()
	nCtx, cancel := context.WithDeadline(ctx, deadline)
	return NewCtx(nCtx, cfg), cancel
}

func WithTimeout(ctx Context, timeout time.Duration) (Context, context.CancelFunc) {
	cfg := ctx.Config()
	nCtx, cancel := context.WithTimeout(ctx, timeout)
	return NewCtx(nCtx, cfg), cancel
}
