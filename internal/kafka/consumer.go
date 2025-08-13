package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"go-postgres-docker/internal/cache"
	"go-postgres-docker/internal/config"
	"go-postgres-docker/internal/model"
	"go-postgres-docker/internal/repository"
	"log"
	"time"
	kgo "github.com/segmentio/kafka-go"
)


func StartConsumer(ctx context.Context, cfg *config.Config, repo *repository.OrderRepository, cache *cache.OrderCache) {
	if cfg == nil {
		log.Println("kafka consumer: config is nil, skipping consumer start")
		return
	}

	if len(cfg.KafkaBrokers) == 0 {
		log.Println("kafka brokers not configured, skipping consumer start")
		return
	}

	if cfg.KafkaTopic == "" {
		log.Println("kafka topic not configured, skipping consumer start")
		return
	}

	if cfg.KafkaGroupID == "" {
		log.Println("kafka group id not configured, skipping consumer start")
		return
	}

	r := kgo.NewReader(kgo.ReaderConfig{
		Brokers:        cfg.KafkaBrokers,
		GroupID:        cfg.KafkaGroupID,
		Topic:          cfg.KafkaTopic,
		MinBytes:       1e3, 
		MaxBytes:       10e6,
		CommitInterval: 0,
	})

	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("kafka reader close error: %v", err)
		}
	}()

	log.Printf("Kafka consumer started (brokers=%v, topic=%s, group=%s)", cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)

	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			log.Println("kafka consumer: context canceled, stopping")
			return
		default:
		}

		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Println("kafka consumer: fetch stopped by context")
				return
			}
			log.Printf("kafka fetch error: %v - retrying after %s", err, backoff)
			time.Sleep(backoff)
			if backoff < 30 * time.Second {
				backoff *= 2
			}
			continue
		}
		backoff = time.Second

		// Декодируем payload
		var order model.Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			// Некорректный JSON - логируем и подтверждаем сообщение, чтобы не застревало (poison message)
			log.Printf("kafka: invalid message json at partition=%d offset=%d: %v; message=%s", msg.Partition, msg.Offset, err, string(msg.Value))
			if err := r.CommitMessages(ctx, msg); err != nil {
				log.Printf("failed to commit offset for invalid message: %v", err)
			}
			continue
		}

		// Если нет времени, то проставляем текущее время
		if order.DateCreated.IsZero() {
			order.DateCreated = time.Now().UTC()
		}
		if order.Payment.PaymentDt == 0 {
			order.Payment.PaymentDt = int(time.Now().Unix())
		}

		// Сохраняем в БД 
		if err := repo.SaveOrder(&order); err != nil {
			log.Printf("failed to save order from kafka (partition=%d offset=%d): %v", msg.Partition, msg.Offset, err)
			// небольшая пауза, чтобы не спамить на ошибке
			time.Sleep(2 * time.Second)
			continue
		}

		// Обновляем кэш
		cache.Set(&order)

		if err := r.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit kafka message (partition=%d offset=%d): %v", msg.Partition, msg.Offset, err)
		} else {
			log.Printf("processed and committed message partition=%d offset=%d", msg.Partition, msg.Offset)
		}
	}
}
