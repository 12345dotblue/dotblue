package im

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type RuntimePayloadHandler func(ctx context.Context, payload []byte) error

type RuntimeStarter func(ctx context.Context, enqueue RuntimePayloadHandler) error

type RuntimeHooks struct {
	Start              RuntimeStarter
	ProcessPayload     RuntimePayloadHandler
	ProcessOutbound    func(ctx context.Context) error
	IsExpectedStop     func(error) bool
	ReconnectBaseDelay time.Duration
	ReconnectMaxDelay  time.Duration
}

type managedRuntime struct {
	Connection Connection
	StartedAt  time.Time
	cancel     context.CancelFunc
	events     chan []byte
	hooks      RuntimeHooks
}

type RuntimeManager struct {
	mu       sync.RWMutex
	runtimes map[string]managedRuntime
}

var defaultRuntimeManager = NewRuntimeManager()

func NewRuntimeManager() *RuntimeManager {
	return &RuntimeManager{
		runtimes: map[string]managedRuntime{},
	}
}

func (m *RuntimeManager) Start(ctx context.Context, conn Connection, hooks RuntimeHooks) error {
	if hooks.Start == nil {
		return fmt.Errorf("runtime start hook is not configured")
	}
	if hooks.ProcessPayload == nil {
		return fmt.Errorf("runtime payload handler is not configured")
	}

	runtimeCtx, cancel := context.WithCancel(context.Background())
	runtime := managedRuntime{
		Connection: conn,
		StartedAt:  time.Now(),
		cancel:     cancel,
		events:     make(chan []byte, 64),
		hooks:      hooks,
	}

	var existingCancel context.CancelFunc
	m.mu.Lock()
	if existing, ok := m.runtimes[conn.ID]; ok {
		existingCancel = existing.cancel
	}
	m.runtimes[conn.ID] = runtime
	m.mu.Unlock()
	if existingCancel != nil {
		existingCancel()
	}

	go m.runWorker(runtimeCtx, runtime)
	if runtime.hooks.ProcessOutbound != nil {
		go m.runOutboundWorker(runtimeCtx, runtime)
	}
	go m.runLoop(runtimeCtx, runtime)
	return nil
}

func (m *RuntimeManager) Stop(ctx context.Context, connectionID string) error {
	m.mu.Lock()
	runtime, ok := m.runtimes[connectionID]
	if ok {
		delete(m.runtimes, connectionID)
	}
	m.mu.Unlock()
	if ok && runtime.cancel != nil {
		runtime.cancel()
	}
	return nil
}

func (m *RuntimeManager) runWorker(ctx context.Context, runtime managedRuntime) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-runtime.events:
			if err := runtime.hooks.ProcessPayload(ctx, payload); err != nil {
				m.recordRuntimeError(runtime.Connection.ID, runtime.StartedAt, runtime.hooks, err)
			}
		}
	}
}

func (m *RuntimeManager) runOutboundWorker(ctx context.Context, runtime managedRuntime) {
	if runtime.hooks.ProcessOutbound == nil {
		return
	}
	if err := runtime.hooks.ProcessOutbound(ctx); err != nil && ctx.Err() == nil {
		m.recordRuntimeError(runtime.Connection.ID, runtime.StartedAt, runtime.hooks, err)
	}
}

func (m *RuntimeManager) runLoop(ctx context.Context, runtime managedRuntime) {
	attempt := 0
	for {
		if ctx.Err() != nil {
			return
		}

		err := runtime.hooks.Start(ctx, func(handlerCtx context.Context, payload []byte) error {
			return m.enqueuePayload(runtime.Connection.ID, payload)
		})
		if err == nil && ctx.Err() == nil {
			err = fmt.Errorf("runtime exited unexpectedly")
		}
		if err == nil || m.isExpectedStop(runtime.hooks, err) || ctx.Err() != nil {
			return
		}

		attempt++
		m.recordRuntimeError(runtime.Connection.ID, runtime.StartedAt, runtime.hooks, err)
		delay := nextRuntimeReconnectDelay(attempt, runtime.hooks.ReconnectBaseDelay, runtime.hooks.ReconnectMaxDelay)
		g.Log().Warningf(context.Background(), "runtime reconnect scheduled, platform=%s, connection=%s, attempt=%d, delay=%s, err=%v", runtime.Connection.Platform, runtime.Connection.ID, attempt, delay, err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

func (m *RuntimeManager) enqueuePayload(connectionID string, payload []byte) error {
	m.mu.RLock()
	runtime, ok := m.runtimes[connectionID]
	m.mu.RUnlock()
	if !ok {
		return nil
	}

	cloned := append([]byte(nil), payload...)
	select {
	case runtime.events <- cloned:
		return nil
	default:
		return fmt.Errorf("runtime inbound queue is full")
	}
}

func (m *RuntimeManager) recordRuntimeError(connectionID string, startedAt time.Time, hooks RuntimeHooks, err error) {
	if err == nil || strings.TrimSpace(connectionID) == "" {
		return
	}
	if m.isExpectedStop(hooks, err) {
		return
	}
	if !m.matchesRuntime(connectionID, startedAt) {
		return
	}
	g.Log().Errorf(context.Background(), "runtime error, connection=%s, err=%v", connectionID, err)
	_ = updateConnectionErrorState(context.Background(), connectionID, err.Error())
}

func (m *RuntimeManager) matchesRuntime(connectionID string, startedAt time.Time) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	runtime, ok := m.runtimes[connectionID]
	if !ok {
		return false
	}
	return runtime.StartedAt.Equal(startedAt)
}

func (m *RuntimeManager) isExpectedStop(hooks RuntimeHooks, err error) bool {
	if err == nil {
		return true
	}
	if hooks.IsExpectedStop != nil {
		return hooks.IsExpectedStop(err)
	}
	return false
}

func nextRuntimeReconnectDelay(attempt int, base, max time.Duration) time.Duration {
	if base <= 0 {
		base = time.Second
	}
	if max <= 0 {
		max = 30 * time.Second
	}
	if attempt <= 1 {
		if base > max {
			return max
		}
		return base
	}

	delay := base
	for i := 1; i < attempt; i++ {
		if delay >= max/2 {
			return max
		}
		delay *= 2
	}
	if delay > max {
		return max
	}
	return delay
}
