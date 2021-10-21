package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/diegohordi/go-kafka/internal/configs"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type ReadFunc func(key []byte, content []byte) error

type Reader interface {
	Read(ctx context.Context, readFunc ReadFunc) (err error)
}

type Writer interface {
	Write(ctx context.Context, msg interface{}) error
}

type Client interface {
	Close()
	Reader
	Writer
}

type defaultClient struct {
	reader *kafka.Reader
	writer *kafka.Writer
}

func NewClient(config configs.KafkaConfigurer, groupName string) Client {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{config.DSN()},
		Topic:     config.Topic(),
		GroupID:   groupName,
		Partition: 0,
	})
	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.DSN()),
		Topic:        config.Topic(),
		RequiredAcks: kafka.RequireAll,
	}
	return &defaultClient{reader: reader, writer: writer}
}

func (c *defaultClient) Close() {
	if c.reader == nil {
		return
	}
	if err := c.reader.Close(); err != nil {
		log.Printf("could not close Kafka connection %v\n", err)
	}
	log.Println("Kafka connection released successfully")
}

func (c *defaultClient) Read(ctx context.Context, readFunc ReadFunc) (err error) {
	if c.reader == nil {
		return fmt.Errorf("no reader was given")
	}
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return fmt.Errorf("an error occured while fetching the message: %w", err)
	}
	defer func(reader *kafka.Reader, ctx context.Context, msg kafka.Message) {
		if err != nil {
			return
		}
		err = reader.CommitMessages(ctx, msg)
	}(c.reader, ctx, msg)
	if readFunc == nil {
		return nil
	}
	err = readFunc(msg.Key, msg.Value)
	return
}

func (c *defaultClient) Write(ctx context.Context, msg interface{}) error {
	if c.writer == nil {
		return fmt.Errorf("no writer was given")
	}
	mb, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("an error occured while marshalling the message: %w", err)
	}
	return c.writer.WriteMessages(ctx, kafka.Message{
		Value: mb,
		Time:  time.Now(),
	})
}
