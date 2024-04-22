package runner

import (
	"context"
	"sync"
	"time"

	"github.com/go-gost/core/logger"
)

var (
	runner = NewRunner()
)

type TaskEvent struct {
	TaskID TaskID
	Err    error
}

type taskState struct {
	task   Task
	cancel context.CancelFunc
}

type Options struct {
	Async    bool
	Interval time.Duration
	Cancel   bool
}

type Option func(opts *Options)

func WithAync(aync bool) Option {
	return func(opts *Options) {
		opts.Async = aync
	}
}

func WithInterval(interval time.Duration) Option {
	return func(opts *Options) {
		opts.Interval = interval
	}
}

func WithCancel(cancel bool) Option {
	return func(opts *Options) {
		opts.Cancel = cancel
	}
}

func Event() <-chan *TaskEvent {
	return runner.Event()
}

func Exec(ctx context.Context, task Task, opts ...Option) error {
	return runner.Exec(ctx, task, opts...)
}

func Cancel(id TaskID) {
	runner.Cancel(id)
}

type Runner struct {
	events chan *TaskEvent
	states map[TaskID]taskState
	mu     sync.RWMutex
}

func NewRunner() *Runner {
	return &Runner{
		events: make(chan *TaskEvent, 16),
		states: make(map[TaskID]taskState),
	}
}

func (r *Runner) Event() <-chan *TaskEvent {
	return r.events
}

func (r *Runner) Exec(ctx context.Context, task Task, opts ...Option) error {
	if task == nil || task.ID() == "" {
		return nil
	}

	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	if options.Cancel {
		r.Cancel(task.ID())
	}

	ctx, cancel := context.WithCancel(ctx)
	r.setState(taskState{
		task:   task,
		cancel: cancel,
	})

	log := logger.Default().WithFields(map[string]any{
		"kind":  "runner",
		"async": options.Async,
	})
	log.Debugf("task %s started", task.ID())

	if !options.Async {
		t := time.Now()
		err := task.Run(ctx)

		log.WithFields(map[string]any{
			"duration": time.Since(t),
		}).Debugf("task %s done: %v", task.ID(), err)

		select {
		case r.events <- &TaskEvent{
			TaskID: task.ID(),
			Err:    err,
		}:
		default:
		}

		return err
	}

	go func() {
		defer cancel()

		t := time.Now()
		defer func() {
			log.WithFields(map[string]any{
				"duration": time.Since(t),
			}).Debugf("task %s done", task.ID())
		}()

		run := func() {
			select {
			case r.events <- &TaskEvent{
				TaskID: task.ID(),
				Err:    task.Run(ctx),
			}:
			default:
			}
		}

		run()

		interval := options.Interval
		if interval <= 0 {
			return
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				run()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (r *Runner) Cancel(id TaskID) {
	r.delState(id)
}

func (r *Runner) setState(state taskState) {
	if state.task == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.states[state.task.ID()] = state
}

func (r *Runner) delState(id TaskID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state := r.states[id]
	if state.cancel != nil {
		state.cancel()
	}

	delete(r.states, id)
}
