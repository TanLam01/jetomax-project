package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	"github.com/redis/go-redis/v9"
)

const messageChannel = "chat:messages"
const realtimeChannel = "chat:realtime"

type MessageBus struct {
	client *redis.Client
	cancel context.CancelFunc
	pubsub *redis.PubSub
	wg     sync.WaitGroup
}

func NewMessageBus(client *redis.Client) *MessageBus { return &MessageBus{client: client} }

func (b *MessageBus) PublishMessage(ctx context.Context, event entity.MessageEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode message event: %w", err)
	}
	if err := b.client.Publish(ctx, messageChannel, payload).Err(); err != nil {
		return fmt.Errorf("publish Redis message event: %w", err)
	}
	return nil
}

func (b *MessageBus) PublishRealtime(ctx context.Context, event entity.RealtimeEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode realtime event: %w", err)
	}
	if err := b.client.Publish(ctx, realtimeChannel, payload).Err(); err != nil {
		return fmt.Errorf("publish Redis realtime event: %w", err)
	}
	return nil
}

func (b *MessageBus) Start(startupCtx context.Context, messageHandler func(entity.MessageEvent), realtimeHandler func(entity.RealtimeEvent)) error {
	ctx, cancel := context.WithCancel(context.Background())
	pubsub := b.client.Subscribe(ctx, messageChannel, realtimeChannel)
	if _, err := pubsub.Receive(startupCtx); err != nil {
		cancel()
		_ = pubsub.Close()
		return fmt.Errorf("subscribe Redis message channel: %w", err)
	}
	b.cancel = cancel
	b.pubsub = pubsub
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for message := range b.pubsub.Channel() {
			switch message.Channel {
			case messageChannel:
				var event entity.MessageEvent
				if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
					slog.Error("decode Redis message event", "error", err)
					continue
				}
				messageHandler(event)
			case realtimeChannel:
				var event entity.RealtimeEvent
				if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
					slog.Error("decode Redis realtime event", "error", err)
					continue
				}
				realtimeHandler(event)
			}
		}
	}()
	return nil
}

func (b *MessageBus) ConnectPresence(ctx context.Context, userID, connectionID string) (bool, error) {
	key := "chat:presence:" + userID
	now := time.Now().UTC()
	pipe := b.client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(now.Unix(), 10))
	card := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.Add(60 * time.Second).Unix()), Member: connectionID})
	pipe.Expire(ctx, key, 2*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("connect Redis presence: %w", err)
	}
	return card.Val() == 0, nil
}

func (b *MessageBus) TouchPresence(ctx context.Context, userID, connectionID string) error {
	key := "chat:presence:" + userID
	pipe := b.client.TxPipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(time.Now().UTC().Add(60 * time.Second).Unix()), Member: connectionID})
	pipe.Expire(ctx, key, 2*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("refresh Redis presence: %w", err)
	}
	return nil
}

func (b *MessageBus) DisconnectPresence(ctx context.Context, userID, connectionID string) (bool, error) {
	key := "chat:presence:" + userID
	now := time.Now().UTC()
	pipe := b.client.TxPipeline()
	pipe.ZRem(ctx, key, connectionID)
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(now.Unix(), 10))
	card := pipe.ZCard(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("disconnect Redis presence: %w", err)
	}
	return card.Val() == 0, nil
}

func (b *MessageBus) OnlinePresence(ctx context.Context, userIDs []string) (map[string]bool, error) {
	result := make(map[string]bool, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	now := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	pipe := b.client.TxPipeline()
	cards := make(map[string]*redis.IntCmd, len(userIDs))
	for _, userID := range userIDs {
		key := "chat:presence:" + userID
		pipe.ZRemRangeByScore(ctx, key, "-inf", now)
		cards[userID] = pipe.ZCard(ctx, key)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("read Redis presence: %w", err)
	}
	for userID, card := range cards {
		result[userID] = card.Val() > 0
	}
	return result, nil
}

func (b *MessageBus) Close() error {
	if b.cancel == nil {
		return nil
	}
	b.cancel()
	err := b.pubsub.Close()
	b.wg.Wait()
	return err
}
