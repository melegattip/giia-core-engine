package testutil

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func SetupTestNATS(t *testing.T, url string) (*nats.Conn, func()) {
	nc, err := nats.Connect(url, nats.Timeout(10*time.Second))
	if err != nil {
		t.Fatalf("Failed to connect to test NATS: %v", err)
	}

	cleanup := func() {
		nc.Close()
	}

	return nc, cleanup
}

func CreateTestStream(t *testing.T, js nats.JetStreamContext, streamName string, subjects []string) {
	_, err := js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: subjects,
	})
	if err != nil {
		t.Fatalf("Failed to create test stream: %v", err)
	}
}

func PurgeStream(t *testing.T, js nats.JetStreamContext, streamName string) {
	stream, err := js.StreamInfo(streamName)
	if err != nil {
		return
	}

	if stream != nil {
		err = js.PurgeStream(streamName)
		if err != nil {
			t.Logf("Failed to purge stream: %v", err)
		}
	}
}

func DeleteStream(t *testing.T, js nats.JetStreamContext, streamName string) {
	err := js.DeleteStream(streamName)
	if err != nil {
		t.Logf("Failed to delete stream: %v", err)
	}
}

func GetStreamInfo(t *testing.T, js nats.JetStreamContext, streamName string) *nats.StreamInfo {
	info, err := js.StreamInfo(streamName)
	if err != nil {
		t.Fatalf("Failed to get stream info: %v", err)
	}
	return info
}

func WaitForMessages(t *testing.T, js nats.JetStreamContext, streamName string, expectedCount uint64, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		info, err := js.StreamInfo(streamName)
		if err == nil && info.State.Msgs >= expectedCount {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}
