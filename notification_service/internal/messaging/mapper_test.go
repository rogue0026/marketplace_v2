package messaging

import (
	"errors"
	"testing"
	"time"

	"github.com/rogue0026/kafka-contracts/contracts"
	"github.com/rogue0026/kafka-contracts/events"
)

func TestNotificationFromKafkaMessage(t *testing.T) {
	t.Parallel()

	occurredAt := time.Unix(1_700_000_000, 0).UTC()

	tests := []struct {
		name      string
		event     events.Event
		wantUser  uint64
		wantTitle string
		wantBody  string
		wantError bool
	}{
		{
			name:      "user created",
			event:     events.UserCreated{UserID: 42},
			wantUser:  42,
			wantTitle: "Account created",
			wantBody:  "Your account has been created successfully.",
		},
		{
			name:      "funds added",
			event:     events.FundsAdded{UserID: 7},
			wantUser:  7,
			wantTitle: "Balance topped up",
			wantBody:  "Funds were added to your balance.",
		},
		{
			name:      "funds debitted",
			event:     events.FundsDebitted{UserID: 9},
			wantUser:  9,
			wantTitle: "Balance charged",
			wantBody:  "Funds were debited from your balance.",
		},
		{
			name:      "order created",
			event:     events.OrderCreated{OrderID: 101, UserID: 11},
			wantUser:  11,
			wantTitle: "Order created",
			wantBody:  "Your order #101 has been created.",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			raw, err := rawMessage(tt.event, occurredAt)
			if err != nil {
				t.Fatalf("rawMessage() error = %v", err)
			}

			got, err := notificationFromKafkaMessage(raw)
			if tt.wantError {
				if err == nil {
					t.Fatal("notificationFromKafkaMessage() error = nil, want non-nil")
				}
				if !isNonRetriable(err) {
					t.Fatalf("notificationFromKafkaMessage() error retriable = true, want false; err=%v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("notificationFromKafkaMessage() error = %v", err)
			}

			if got.UserID != tt.wantUser {
				t.Fatalf("notificationFromKafkaMessage() user_id = %d, want %d", got.UserID, tt.wantUser)
			}
			if got.Title != tt.wantTitle {
				t.Fatalf("notificationFromKafkaMessage() title = %q, want %q", got.Title, tt.wantTitle)
			}
			if got.Body != tt.wantBody {
				t.Fatalf("notificationFromKafkaMessage() body = %q, want %q", got.Body, tt.wantBody)
			}
			if !got.CreatedAt.Equal(occurredAt) {
				t.Fatalf("notificationFromKafkaMessage() created_at = %v, want %v", got.CreatedAt, occurredAt)
			}
		})
	}
}

func rawMessage(event events.Event, occurredAt time.Time) ([]byte, error) {
	message, err := events.NewMessage(event, contracts.UserService, occurredAt)
	if err != nil {
		return nil, err
	}

	return message.Raw()
}

func TestNotificationFromKafkaMessage_InvalidJSONIsNonRetriable(t *testing.T) {
	t.Parallel()

	_, err := notificationFromKafkaMessage([]byte("not-json"))
	if err == nil {
		t.Fatal("notificationFromKafkaMessage() error = nil, want non-nil")
	}

	if !isNonRetriable(err) {
		t.Fatalf("notificationFromKafkaMessage() error retriable = true, want false; err=%v", err)
	}

	var target *nonRetriableError
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(nonRetriableError) = false, want true; err=%v", err)
	}
}
