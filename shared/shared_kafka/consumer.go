package shared_kafka

import (
	"context"
	"net"
	"shared/loggers"
	"time"

	_ "crypto/sha512"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type KafkaConsumer struct {
	Consumer *kafka.Reader
	Logger   *loggers.Logger
}
type ConfigConsumer struct {
	Brokers       []string
	KafkaUser     string
	KafkaPassword string
	Topic         string
	GroupID       string
}

func NewConsumer(conf *ConfigConsumer, logger *loggers.Logger) (*KafkaConsumer, error) {
	mechanism, errScramMechanism := scram.Mechanism(scram.SHA512, conf.KafkaUser, conf.KafkaPassword)
	if errScramMechanism != nil {
		return nil, errScramMechanism
	}
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        conf.Brokers,
		GroupID:        conf.GroupID,
		Topic:          conf.Topic,
		CommitInterval: 0,
		Dialer: &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     false,
			SASLMechanism: mechanism,
			Resolver: &net.Resolver{
				PreferGo: true,
			},
		},
	})
	return &KafkaConsumer{
		Consumer: consumer,
		Logger:   logger,
	}, nil
}
func (c *KafkaConsumer) WaitEvent(ctxCancel context.Context, processorFunc func([]byte) error) {
	time.Sleep(5 * time.Second)
	for {
		msg, errReadMessage := c.Consumer.ReadMessage(ctxCancel)
		if errReadMessage != nil {
			c.Logger.Error("failed to read message consumer kafka: " + errReadMessage.Error())
			continue
		}
		if errProcess := processorFunc(msg.Value); errProcess != nil {
			c.Logger.Error("failed to process message consumer kafka: " + errProcess.Error())
			continue
		}
		ctxTimeout, cancel := context.WithTimeout(ctxCancel, time.Second*3)
		if errCommit := c.Consumer.CommitMessages(ctxTimeout, msg); errCommit != nil {
			c.Logger.Error("failed to commit message consumer kafka: " + errCommit.Error())
		}
		cancel()
	}
}
func (c *KafkaConsumer) CloseConsumer() {
	if errClose := c.Consumer.Close(); errClose != nil {
		c.Logger.Error("failed to close consumer kafka connection: " + errClose.Error())
	}
}
