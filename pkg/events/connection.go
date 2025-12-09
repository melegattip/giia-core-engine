package events

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type ConnectionConfig struct {
	URL             string
	MaxReconnects   int
	ReconnectWait   time.Duration
	ConnectionName  string
}

func Connect(config *ConnectionConfig) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name(config.ConnectionName),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(config.ReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				fmt.Printf("NATS disconnected: %v\n", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("NATS reconnected to %s\n", nc.ConnectedUrl())
		}),
	}

	nc, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return nc, nil
}

func ConnectWithDefaults(url string) (*nats.Conn, error) {
	config := &ConnectionConfig{
		URL:             url,
		MaxReconnects:   10,
		ReconnectWait:   2 * time.Second,
		ConnectionName:  "giia-service",
	}

	return Connect(config)
}

func Disconnect(nc *nats.Conn) error {
	if nc == nil {
		return nil
	}

	nc.Drain()
	nc.Close()

	return nil
}
