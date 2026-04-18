package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rogue0026/kafka-contracts/contracts"
	"github.com/segmentio/kafka-go"
)

type Relay struct {
	pool    *pgxpool.Pool
	writers map[contracts.Topic]*kafka.Writer
}

func NewRelay(pool *pgxpool.Pool, brokers []string, topics []contracts.Topic) *Relay {
	writers := make(map[contracts.Topic]*kafka.Writer)
	for _, topicName := range topics {
		writers[topicName] = &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  string(topicName),
			MaxAttempts:            1,
			Compression:            kafka.Lz4,
			AllowAutoTopicCreation: true,
		}
	}

	return &Relay{
		pool:    pool,
		writers: writers,
	}

}

type key struct {
	rowID uint64
	topic string
}

func (r *Relay) Run(ctx context.Context) {
	const getOutboxRecordsSQL = `
	SELECT 
	    id,
	    topic_name,
	    partition_key,
	    payload 
	FROM outbox 
	WHERE status = 'PENDING' 
	LIMIT 100
	FOR UPDATE SKIP LOCKED`

	t := time.NewTicker(time.Second * 2)

LOOP:
	for {
		select {
		case <-t.C:
			records, err := r.pool.Query(ctx, getOutboxRecordsSQL)
			if err != nil {
				fmt.Printf("outbox relay: %s\n", err.Error())
				continue
			}

			messages := make(map[key]*kafka.Message)

			for records.Next() {
				var rowID uint64
				var topic string
				var partitionKey string
				var payload json.RawMessage
				err = records.Scan(&rowID, &topic, &partitionKey, &payload)
				if err != nil {
					fmt.Printf("outbox relay: %s\n", err.Error())
					continue
				}
				k := key{
					rowID: rowID,
					topic: topic,
				}
				messages[k] = &kafka.Message{
					Offset: kafka.LastOffset,
					Key:    []byte(partitionKey),
					Value:  payload,
				}
			}

			err = records.Err()
			if err != nil {
				fmt.Printf("outbox relay: %s\n", err.Error())
			}
			records.Close()

			sentSuccessfully := make([]uint64, 0)
			sentWithErrors := make([]struct {
				id  uint64
				err error
			}, 0)

			for k, msg := range messages {
				wr, ok := r.writers[contracts.Topic(k.topic)]
				if ok {
					err = wr.WriteMessages(ctx, *msg)
					if err != nil {
						fmt.Printf("outbox relay, topic %s: %w\n", k.topic, err.Error())
						sentWithErrors = append(sentWithErrors, struct {
							id  uint64
							err error
						}{
							id:  k.rowID,
							err: err,
						})
						continue
					}
					fmt.Printf("message %v sent successfully to topic %s\n", string(msg.Value), k.topic)
					sentSuccessfully = append(sentSuccessfully, k.rowID)
				}
			}

			_, err = r.pool.Exec(
				ctx,
				`UPDATE outbox SET status = 'PROCESSED' WHERE id = ANY($1)`,
				sentSuccessfully,
			)
			if err != nil {
				fmt.Printf("outbox relay: %s\n", err.Error())
			}

			for _, fail := range sentWithErrors {
				_, err = r.pool.Exec(
					ctx,
					`UPDATE outbox SET last_error = $2 WHERE id = $1`,
					fail.id,
					fail.err.Error(),
				)
				if err != nil {
					fmt.Printf("outbox relay: %s\n", err.Error())
				}
			}

		case <-ctx.Done():
			break LOOP
		}
	}
}
