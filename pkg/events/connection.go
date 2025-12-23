package events

import (
	"time"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/nats-io/nats.go"
)

type ConnectionConfig struct {
	URL               string
	MaxReconnects     int
	ReconnectWait     time.Duration
	ConnectionName    string
	DisconnectHandler func(*nats.Conn, error)
	ReconnectHandler  func(*nats.Conn)
}

func Connect(config *ConnectionConfig) (*nats.Conn, error) {
	if config == nil {
		return nil, errors.NewBadRequest("connection config is required")
	}

	if config.URL == "" {
		return nil, errors.NewBadRequest("NATS URL is required")
	}

	opts := []nats.Option{
		nats.Name(config.ConnectionName),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(config.ReconnectWait),
	}

	if config.DisconnectHandler != nil {
		opts = append(opts, nats.DisconnectErrHandler(config.DisconnectHandler))
	}

	if config.ReconnectHandler != nil {
		opts = append(opts, nats.ReconnectHandler(config.ReconnectHandler))
	}

	nc, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to connect to NATS")
	}

	return nc, nil
}

func ConnectWithDefaults(url string) (*nats.Conn, error) {
	config := &ConnectionConfig{
		URL:            url,
		MaxReconnects:  10,
		ReconnectWait:  2 * time.Second,
		ConnectionName: "giia-service",
	}

	return Connect(config)
}

func Disconnect(nc *nats.Conn) error {
	if nc == nil {
		return nil
	}

	if err := nc.Drain(); err != nil {
		return errors.NewInternalServerError("failed to drain connection")
	}

	nc.Close()
	return nil
}
