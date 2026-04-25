package messaging

import (
	"context"
	"fmt"
	"notification_service/internal/domain"
	"time"

	"github.com/rogue0026/kafka-contracts/contracts"
	"github.com/segmentio/kafka-go"
)

type NotificationsRepository interface {
	CreateNotification(ctx context.Context, item *domain.Notification) error
}

type Consumer struct {
	reader *kafka.Reader
	repo   NotificationsRepository
}

func NewConsumer(brokers []string, groupID string, repo NotificationsRepository) *Consumer {
	topics := []contracts.Topic{
		contracts.UserEvents,
		contracts.WalletEvents,
		contracts.OrderEvents,
	}
	topicsAsString := make([]string, 0, len(topics))
	for _, topic := range topics {
		topicsAsString = append(topicsAsString, string(topic))
	}

	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:                brokers,
			GroupID:                groupID,
			GroupTopics:            topicsAsString,
			MinBytes:               1,
			MaxBytes:               10e6,
			MaxWait:                time.Second,
			WatchPartitionChanges:  true,
			PartitionWatchInterval: 2 * time.Second,
		}),
		repo: repo,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			fmt.Printf("notifications consumer, fetch message: %s\n", err.Error())
			continue
		}

		err = c.handleMessage(ctx, msg)
		shouldCommit := err == nil || isNonRetriable(err)
		if err != nil {
			fmt.Printf("notifications consumer, topic=%s, offset=%d: %s\n", msg.Topic, msg.Offset, err.Error())
		}

		if !shouldCommit {
			continue
		}

		if err = c.reader.CommitMessages(ctx, msg); err != nil {
			if ctx.Err() != nil {
				return nil
			}

			fmt.Printf("notifications consumer, topic=%s, offset=%d, commit: %s\n", msg.Topic, msg.Offset, err.Error())
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	notification, err := notificationFromKafkaMessage(msg.Value)
	if err != nil {
		return err
	}

	if err = c.repo.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("save notification: %w", err)
	}

	return nil
}
