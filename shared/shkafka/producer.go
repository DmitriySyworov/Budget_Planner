package shkafka

import (
	"context"
	"net"
	"shared/loggers"
	"time"

	_ "crypto/sha256"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type KafkaProducer struct {
	Producer *kafka.Writer
	Logger   *loggers.Logger
}
type ConfigProducer struct {
	Brokers       []string
	KafkaUser     string
	KafkaPassword string
	Topic         string
}

func NewProducer(conf *ConfigProducer, logger *loggers.Logger) (*KafkaProducer, error) {
	mechanism, errScramMechanism := scram.Mechanism(scram.SHA256, conf.KafkaUser, conf.KafkaPassword)
	if errScramMechanism != nil {
		return nil, errScramMechanism
	}

	netDialer := &net.Dialer{
		Timeout:   10 * time.Second,
		DualStack: false,
		Resolver: &net.Resolver{
			PreferGo: true,
		},
	}
	customTransport := &kafka.Transport{
		SASL: mechanism,
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return netDialer.DialContext(ctx, network, addr)
		},
	}

	producer := &kafka.Writer{
		Addr:        kafka.TCP(conf.Brokers...),
		Topic:       conf.Topic,
		Balancer:    &kafka.Hash{},
		MaxAttempts: 5,
		Transport:   customTransport,
	}
	return &KafkaProducer{
		Producer: producer,
		Logger:   logger,
	}, nil
}

func (p *KafkaProducer) SendEvent(ctxTimeout context.Context, keyUUID string, event []byte) error {
	if errSendMessage := p.Producer.WriteMessages(ctxTimeout, kafka.Message{
		Key:   []byte(keyUUID),
		Value: event,
	}); errSendMessage != nil {
		return errSendMessage
	}
	return nil
}
func (p *KafkaProducer) CloseProducer() {
	if errClose := p.Producer.Close(); errClose != nil {
		p.Logger.Error("failed to close producer kafka connection: " + errClose.Error())
	}
}
