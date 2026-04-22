package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
	"notification_service/internal/domain"

	"github.com/rogue0026/kafka-contracts/contracts"
	"github.com/rogue0026/kafka-contracts/events"
)

type nonRetriableError struct {
	err error
}

func (e *nonRetriableError) Error() string {
	return e.err.Error()
}

func (e *nonRetriableError) Unwrap() error {
	return e.err
}

func newNonRetriableError(format string, args ...any) error {
	return &nonRetriableError{
		err: fmt.Errorf(format, args...),
	}
}

func isNonRetriable(err error) bool {
	if err == nil {
		return false
	}

	var target *nonRetriableError
	return errors.As(err, &target)
}

func notificationFromKafkaMessage(raw []byte) (*domain.Notification, error) {
	var message events.Message
	if err := json.Unmarshal(raw, &message); err != nil {
		return nil, newNonRetriableError("decode kafka envelope: %w", err)
	}

	switch message.EventType {
	case contracts.UserCreated:
		var event events.UserCreated
		if err := json.Unmarshal(message.Payload, &event); err != nil {
			return nil, newNonRetriableError("decode %s payload: %w", message.EventType, err)
		}

		return &domain.Notification{
			UserID:    event.UserID,
			Title:     "Account created",
			Body:      "Your account has been created successfully.",
			CreatedAt: message.OccurredAt,
		}, nil
	case contracts.FundsAdded:
		var event events.FundsAdded
		if err := json.Unmarshal(message.Payload, &event); err != nil {
			return nil, newNonRetriableError("decode %s payload: %w", message.EventType, err)
		}

		return &domain.Notification{
			UserID:    event.UserID,
			Title:     "Balance topped up",
			Body:      "Funds were added to your balance.",
			CreatedAt: message.OccurredAt,
		}, nil
	case contracts.FundsDebitted:
		var event events.FundsDebitted
		if err := json.Unmarshal(message.Payload, &event); err != nil {
			return nil, newNonRetriableError("decode %s payload: %w", message.EventType, err)
		}

		return &domain.Notification{
			UserID:    event.UserID,
			Title:     "Balance charged",
			Body:      "Funds were debited from your balance.",
			CreatedAt: message.OccurredAt,
		}, nil
	case contracts.OrderCreated:
		var event events.OrderCreated
		if err := json.Unmarshal(message.Payload, &event); err != nil {
			return nil, newNonRetriableError("decode %s payload: %w", message.EventType, err)
		}
		if event.UserID == 0 {
			return nil, newNonRetriableError("decode %s payload: user_id is empty", message.EventType)
		}
		if event.OrderID == 0 {
			return nil, newNonRetriableError("decode %s payload: order_id is empty", message.EventType)
		}

		return &domain.Notification{
			UserID:    event.UserID,
			Title:     "Order created",
			Body:      fmt.Sprintf("Your order #%d has been created.", event.OrderID),
			CreatedAt: message.OccurredAt,
		}, nil
	case contracts.OrderPayedFor:
		return nil, newNonRetriableError(
			"event %s cannot be converted to a user notification: payload does not contain user_id",
			message.EventType,
		)
	default:
		return nil, newNonRetriableError("unsupported event type: %s", message.EventType)
	}
}
