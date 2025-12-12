package events

import (
	"time"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/nats-io/nats.go"
)

type StreamConfig struct {
	Name       string
	Subjects   []string
	MaxAge     time.Duration
	MaxBytes   int64
	Replicas   int
}

func NewStreamConfig(name string, subjects []string) *StreamConfig {
	return &StreamConfig{
		Name:       name,
		Subjects:   subjects,
		MaxAge:     7 * 24 * time.Hour,
		MaxBytes:   1024 * 1024 * 1024,
		Replicas:   1,
	}
}

func (sc *StreamConfig) ToNATSConfig() *nats.StreamConfig {
	return &nats.StreamConfig{
		Name:       sc.Name,
		Subjects:   sc.Subjects,
		Storage:    nats.FileStorage,
		Retention:  nats.LimitsPolicy,
		MaxAge:     sc.MaxAge,
		MaxBytes:   sc.MaxBytes,
		Replicas:   sc.Replicas,
		Discard:    nats.DiscardOld,
		Duplicates: 2 * time.Minute,
	}
}

func CreateStream(js nats.JetStreamContext, config *StreamConfig) error {
	if js == nil {
		return errors.NewBadRequest("JetStream context is required")
	}

	if config == nil {
		return errors.NewBadRequest("stream config is required")
	}

	if config.Name == "" {
		return errors.NewBadRequest("stream name is required")
	}

	if len(config.Subjects) == 0 {
		return errors.NewBadRequest("stream subjects are required")
	}

	_, err := js.AddStream(config.ToNATSConfig())
	if err != nil {
		return errors.NewInternalServerError("failed to create stream")
	}

	return nil
}

func UpdateStream(js nats.JetStreamContext, config *StreamConfig) error {
	if js == nil {
		return errors.NewBadRequest("JetStream context is required")
	}

	if config == nil {
		return errors.NewBadRequest("stream config is required")
	}

	_, err := js.UpdateStream(config.ToNATSConfig())
	if err != nil {
		return errors.NewInternalServerError("failed to update stream")
	}

	return nil
}

func DeleteStream(js nats.JetStreamContext, streamName string) error {
	if js == nil {
		return errors.NewBadRequest("JetStream context is required")
	}

	if streamName == "" {
		return errors.NewBadRequest("stream name is required")
	}

	err := js.DeleteStream(streamName)
	if err != nil {
		return errors.NewInternalServerError("failed to delete stream")
	}

	return nil
}

func GetStreamInfo(js nats.JetStreamContext, streamName string) (*nats.StreamInfo, error) {
	if js == nil {
		return nil, errors.NewBadRequest("JetStream context is required")
	}

	if streamName == "" {
		return nil, errors.NewBadRequest("stream name is required")
	}

	info, err := js.StreamInfo(streamName)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to get stream info")
	}

	return info, nil
}

func GetDefaultStreams() []*StreamConfig {
	return []*StreamConfig{
		NewStreamConfig("AUTH_EVENTS", []string{"auth.>"}),
		NewStreamConfig("CATALOG_EVENTS", []string{"catalog.>"}),
		NewStreamConfig("DDMRP_EVENTS", []string{"ddmrp.>"}),
		NewStreamConfig("EXECUTION_EVENTS", []string{"execution.>"}),
		NewStreamConfig("ANALYTICS_EVENTS", []string{"analytics.>"}),
		NewStreamConfig("AI_AGENT_EVENTS", []string{"ai_agent.>"}),
		NewStreamConfig("DLQ_EVENTS", []string{"dlq.>"}),
	}
}

func CreateDefaultStreams(js nats.JetStreamContext) error {
	if js == nil {
		return errors.NewBadRequest("JetStream context is required")
	}

	streams := GetDefaultStreams()
	for _, stream := range streams {
		if err := CreateStream(js, stream); err != nil {
			return err
		}
	}

	return nil
}
