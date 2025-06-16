package log

import (
	"fmt"
	"load-balancer/conf"
	"log"
	"net"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func waitForPort(host string, port string, timeout time.Duration) error {
	address := net.JoinHostPort(host, port)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 2*time.Second)
		if err == nil {
			_ = conn.Close()
			return nil // port is open
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", address)
}

func TestKafkaLogger(t *testing.T) {
	conf := &conf.Conf{
		Kafka: conf.KafkaConf{
			Servers:  "localhost:9093",
			ClientId: "test-client",
			LogTopic: "test-logs",
		},
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal("Could not connect to Docker:", err)
	}
	network, err := pool.CreateNetwork("kafka_test")
	if err != nil {
		t.Fatalf("Could not create Docker network: %s", err)
	}

	zooResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "confluentinc/cp-zookeeper",
		Tag:        "7.8.0",
		Name:       "zoo1",
		Env: []string{
			"ZOOKEEPER_CLIENT_PORT=2182",
			"ZOOKEEPER_TICK_TIME=2000",
		},
		ExposedPorts: []string{"2182/tcp"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"2182/tcp": {{HostIP: "0.0.0.0", HostPort: "2182"}},
		},
		Networks: []*dockertest.Network{network},
	})
	if err != nil {
		t.Fatalf("Could not start Zookeeper container: %s", err)
	}

	fmt.Println("Zookeeper container started:", zooResource.Container.Name)

	// Wait for Zookeeper port 2182 to be ready
	if err := waitForPort("localhost", "2182", 30*time.Second); err != nil {
		t.Fatalf("Zookeeper port not ready: %s", err)
	}
	fmt.Println("Zookeeper port 2182 is ready")

	kafkaResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "confluentinc/cp-kafka",
		Tag:        "7.8.0",
		Name:       "kafka",
		Env: []string{
			"KAFKA_BROKER_ID=1",
			"KAFKA_ZOOKEEPER_CONNECT=zoo1:2182",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			"KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092,PLAINTEXT_HOST://0.0.0.0:9093",
			"KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092,PLAINTEXT_HOST://localhost:9093",
			"KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS=0",
			"KAFKA_LOG4J_LOGGERS=kafka.controller=INFO,kafka.producer.async.DefaultEventHandler=INFO,state.change.logger=INFO",
		},
		ExposedPorts: []string{"9092/tcp", "9093/tcp"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9093/tcp": {{HostIP: "0.0.0.0", HostPort: "9093"}},
			"9092/tcp": {{HostIP: "0.0.0.0", HostPort: "9092"}},
		},
		Networks: []*dockertest.Network{network},
	})
	if err != nil {
		t.Fatalf("Could not start Kafka container: %s", err)
	}
	fmt.Println("Kafka container started:", kafkaResource.Container.Name)

	// Wait for Kafka port 9093 to be ready
	if err := waitForPort("localhost", "9093", 60*time.Second); err != nil {
		t.Fatalf("Kafka port not ready: %s", err)
	}
	fmt.Println("Kafka port 9093 is ready")

	time.Sleep(5 * time.Second) // Give Kafka some time to initialize
	fmt.Println("Waiting for Kafka to be ready via logger...")
	logger, err := NewKafkaLogger(conf)
	if err != nil {
		t.Fatal("Failed to create Kafka logger:", err)
	}
	err = logger.Info("Test message 1", "This is a test log entry for Kafka logger")
	if err != nil {
		t.Fatal("Failed to log message:", err)
	}
	defer func() {
		err := pool.RemoveNetwork(network)
		if err != nil {
			log.Printf("Could not remove Docker network: %s", err)
		}

		if err := pool.Purge(zooResource); err != nil {
			log.Printf("Could not purge zookeeper container: %s", err)
		}

		if err := pool.Purge(kafkaResource); err != nil {
			log.Printf("Could not purge kafka container: %s", err)
		}
	}()

}
