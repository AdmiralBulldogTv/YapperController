package global

import (
	"context"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/configure"
)

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
	Config() *configure.Config
	Inst() *Instance
}

type gCtx struct {
	ctx  context.Context
	cfg  *configure.Config
	inst *Instance
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

func (c *gCtx) Config() *configure.Config {
	return c.cfg
}

func (c *gCtx) Inst() *Instance {
	return c.inst
}

func NewCtx(ctx context.Context, config *configure.Config) Context {
	return &gCtx{ctx: ctx, cfg: config, inst: &Instance{}}
}

func WithValue(ctx Context, key interface{}, value interface{}) Context {
	return &gCtx{ctx: context.WithValue(ctx, key, value), cfg: ctx.Config(), inst: ctx.Inst()}
}

func WithCancel(ctx Context) (Context, context.CancelFunc) {
	nCtx, cancel := context.WithCancel(ctx)
	return &gCtx{ctx: nCtx, cfg: ctx.Config(), inst: ctx.Inst()}, cancel
}

func WithDeadline(ctx Context, deadline time.Time) (Context, context.CancelFunc) {
	nCtx, cancel := context.WithDeadline(ctx, deadline)
	return &gCtx{ctx: nCtx, cfg: ctx.Config(), inst: ctx.Inst()}, cancel
}

func WithTimeout(ctx Context, timeout time.Duration) (Context, context.CancelFunc) {
	nCtx, cancel := context.WithTimeout(ctx, timeout)
	return &gCtx{ctx: nCtx, cfg: ctx.Config(), inst: ctx.Inst()}, cancel
}
