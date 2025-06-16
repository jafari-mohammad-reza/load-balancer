package log

import (
	"encoding/json"
	"fmt"
	"load-balancer/conf"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaLogger struct {
	conf     *conf.Conf
	mu       sync.Mutex
	producer *kafka.Producer
}

func NewKafkaLogger(conf *conf.Conf) (*KafkaLogger, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": conf.Kafka.Servers,
		"client.id":         conf.Kafka.ClientId,
		"acks":              "all"})

	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %s", err)
	}
	return &KafkaLogger{
		conf:     conf,
		mu:       sync.Mutex{},
		producer: p,
	}, nil
}
func (k *KafkaLogger) write(level LogLevel, args ...any) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	entry := Log{
		DateTime: time.Now().Format(time.RFC3339Nano),
		Level:    level,
		Args:     toStringSlice(args),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	topic := k.conf.Kafka.LogTopic
	deliveryChan := make(chan kafka.Event, 1)

	err = k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          data,
	}, deliveryChan)

	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		return fmt.Errorf("failed to deliver message: %v", m.TopicPartition.Error)
	}

	return nil
}
func (k *KafkaLogger) Info(args ...any) error {
	return k.write(Info, args...)
}
func (k *KafkaLogger) Warn(args ...any) error {
	return k.write(Warn, args...)
}
func (k *KafkaLogger) Error(args ...any) error {
	return k.write(Error, args...)
}
func (k *KafkaLogger) Debug(args ...any) error {
	return k.write(Debug, args...)
}
