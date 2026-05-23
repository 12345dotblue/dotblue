package dataplane

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Address   string `json:"address" yaml:"address"`
	Password  string `json:"password" yaml:"password"`
	DB        int    `json:"db" yaml:"db"`
	KeyPrefix string `json:"keyPrefix" yaml:"keyPrefix"`
}

type Config struct {
	RequestStateRunningTTL time.Duration
	RequestStateFinalTTL   time.Duration
	StreamMaxLen           int64
}

type WorkerConfig struct {
	ID        string
	InboxTTL  time.Duration
	ClaimBlock time.Duration
}

type TurnTask struct {
	RequestID        string    `json:"requestId"`
	SessionKey       string    `json:"sessionKey"`
	FenceToken       int64     `json:"fenceToken"`
	EnterpriseID     string    `json:"enterpriseId"`
	UserID           string    `json:"userId"`
	AgentID          string    `json:"agentId"`
	ConversationID   string    `json:"conversationId"`
	ConnectionID     string    `json:"connectionId"`
	IngressType      string    `json:"ingressType"`
	InboundMessageID string    `json:"inboundMessageId"`
	ExternalChatID   string    `json:"externalChatId,omitempty"`
	ExternalThreadID string    `json:"externalThreadId,omitempty"`
	ReplyHandle      map[string]any `json:"replyHandle,omitempty"`
	Content          string    `json:"content"`
	CreatedAt        time.Time `json:"createdAt"`
}

type TurnEvent struct {
	RequestID      string    `json:"requestId"`
	SessionKey     string    `json:"sessionKey"`
	ConversationID string    `json:"conversationId"`
	Type           string    `json:"type"`
	Status         string    `json:"status"`
	Content        string    `json:"content,omitempty"`
	Thinking       string    `json:"thinking,omitempty"`
	Error          string    `json:"error,omitempty"`
	Seq            int64     `json:"seq"`
	At             time.Time `json:"at"`
}

type QueuedTask struct {
	MessageID string
	Task      TurnTask
}

type EventSubscription interface {
	Next(ctx context.Context) (*TurnEvent, error)
	Close() error
}

type TaskQueue interface {
	Enqueue(ctx context.Context, workerID string, task TurnTask) error
	Claim(ctx context.Context, workerID string, block time.Duration) (*QueuedTask, error)
	Ack(ctx context.Context, workerID, messageID string) error
}

type RequestEventBus interface {
	Publish(ctx context.Context, requestID string, event TurnEvent) error
	Subscribe(ctx context.Context, requestID string) (EventSubscription, error)
}

type RequestStateStore interface {
	SetRunning(ctx context.Context, requestID string, fields map[string]any) error
	SetFinal(ctx context.Context, requestID string, fields map[string]any) error
	Patch(ctx context.Context, requestID string, fields map[string]any, ttl time.Duration) error
}

type RequestRouteStore interface {
	Set(ctx context.Context, requestID string, fields map[string]any, ttl time.Duration) error
}

type QueuedOutbound struct {
	MessageID string
	Payload   string
}

type IMOutbox interface {
	EnqueueJSON(ctx context.Context, connectionID string, payload string) error
	Claim(ctx context.Context, connectionID string, block time.Duration) (*QueuedOutbound, error)
	Ack(ctx context.Context, connectionID, messageID string) error
}

type Redis struct {
	cfg        RedisConfig
	keyspace   Keyspace
	client     *redis.Client
	dpCfg      Config
	workerCfg  WorkerConfig
	streamGroup string
}

var (
	defaultOnce sync.Once
	defaultDP   *Redis
)

func Default(ctx context.Context) (*Redis, error) {
	var err error
	defaultOnce.Do(func() {
		var rc RedisConfig
		if e := g.Cfg().MustGet(ctx, "redis").Struct(&rc); e != nil {
			err = e
			return
		}
		if rc.KeyPrefix == "" {
			err = errors.New("redis.keyPrefix is required")
			return
		}
		var dp Config
		dp.RequestStateRunningTTL = g.Cfg().MustGet(ctx, "dataplane.requestStateRunningTTL").Duration()
		dp.RequestStateFinalTTL = g.Cfg().MustGet(ctx, "dataplane.requestStateFinalTTL").Duration()
		dp.StreamMaxLen = g.Cfg().MustGet(ctx, "dataplane.streamMaxLen").Int64()
		client := redis.NewClient(&redis.Options{
			Addr:     rc.Address,
			Password: rc.Password,
			DB:       rc.DB,
		})
		defaultDP = &Redis{
			cfg:         rc,
			keyspace:    NewKeyspace(rc.KeyPrefix),
			client:      client,
			dpCfg:       dp,
			streamGroup: "cg",
		}
	})
	return defaultDP, err
}

func (r *Redis) Keyspace() Keyspace {
	return r.keyspace
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func (r *Redis) RunningTTL() time.Duration {
	if r == nil {
		return 0
	}
	return r.dpCfg.RequestStateRunningTTL
}

func (r *Redis) FinalTTL() time.Duration {
	if r == nil {
		return 0
	}
	return r.dpCfg.RequestStateFinalTTL
}

func (r *Redis) StreamMaxLen() int64 {
	if r == nil {
		return 0
	}
	return r.dpCfg.StreamMaxLen
}

func marshalJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
