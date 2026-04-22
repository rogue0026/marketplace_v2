package messaging

import (
	"context"
	"fmt"
	"notification_service/internal/domain"
	"time"

	"github.com/rogue0026/kafka-contracts/contracts"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
)

type NotificationsRepository interface {
	CreateNotification(ctx context.Context, item *domain.Notification) error
}

type Consumer struct {
	readers []*topicReader
}

type topicReader struct {
	topic  contracts.Topic
	reader *kafka.Reader
	repo   NotificationsRepository
}

func NewConsumer(brokers []string, groupID string, repo NotificationsRepository) *Consumer {
	topics := []contracts.Topic{
		contracts.UserEvents,
		contracts.WalletEvents,
		contracts.OrderEvents,
	}

	readers := make([]*topicReader, 0, len(topics))
	for _, topic := range topics {
		readers = append(readers, &topicReader{
			topic: topic,
			reader: kafka.NewReader(kafka.ReaderConfig{
				Brokers:  brokers,
				GroupID:  groupID,
				Topic:    string(topic),
				MinBytes: 1,
				MaxBytes: 10e6,
				MaxWait:  time.Second,
			}),
			repo: repo,
		})
	}

	return &Consumer{
		readers: readers,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, topicReader := range c.readers {
		reader := topicReader
		g.Go(func() error {
			return reader.Run(ctx)
		})
	}

	return g.Wait()
}

func (c *Consumer) Close() error {
	var firstErr error

	for _, reader := range c.readers {
		if err := reader.reader.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (r *topicReader) Run(ctx context.Context) error {
	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			fmt.Printf("notifications consumer, topic=%s, fetch message: %s\n", r.topic, err.Error())
			continue
		}

		err = r.handleMessage(ctx, msg)
		shouldCommit := err == nil || isNonRetriable(err)
		if err != nil {
			fmt.Printf("notifications consumer, topic=%s, offset=%d: %s\n", r.topic, msg.Offset, err.Error())
		}

		if !shouldCommit {
			continue
		}

		if err = r.reader.CommitMessages(ctx, msg); err != nil {
			if ctx.Err() != nil {
				return nil
			}

			fmt.Printf("notifications consumer, topic=%s, offset=%d, commit: %s\n", r.topic, msg.Offset, err.Error())
		}
	}
}

func (r *topicReader) handleMessage(ctx context.Context, msg kafka.Message) error {
	notification, err := notificationFromKafkaMessage(msg.Value)
	if err != nil {
		return err
	}

	if err = r.repo.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("save notification: %w", err)
	}

	return nil
}
