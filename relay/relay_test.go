package relay

import (
	"reflect"
	"testing"
	"time"
)

func TestRelayImpl_SubscribeAndUnsubscribe(t *testing.T) {
	relay := NewRelay()
	data := []byte{1, 2, 3, 4}

	pub, err := relay.Publish("test")
	if err != nil {
		t.Fatal(err)
	}

	sub, unsub, err := relay.Subscribe("test")
	if err != nil {
		t.Fatal(err)
	}

	// send
	pub <- data

	// receive
	got, ok := <-sub
	if !ok {
		t.Fatal("Subscriber channel should not be closed")
	}
	if !reflect.DeepEqual(got, data) {
		t.Errorf("Sub ret = %x, want %x", got, data)
	}

	// unsubscribe
	unsub()

	// 2nd send
	pub <- data
	got, ok = <-sub

	if got != nil || ok {
		t.Errorf("Read after unsub ret %x, want nil", got)
	}
}

func TestRelayImpl_PublisherClose(t *testing.T) {
	relay := NewRelay()

	ch, _ := relay.Publish("test")
	sub, unsub, _ := relay.Subscribe("test")
	close(ch)

	// Wait for async teardown in goroutine
	time.Sleep(100 * time.Millisecond)

	if _, ok := <-sub; ok {
		t.Error("Subscriber channel should be closed")
	}

	// unsub after close shouldn't break
	unsub()

	_, err := relay.Publish("test")
	if err != nil {
		t.Error("Publish should be possible again after close")
	}
}

func TestRelayImpl_DoublePublish(t *testing.T) {
	relay := NewRelay()
	relay.Publish("foo")
	_, err := relay.Publish("foo")

	if err != StreamAlreadyExists {
		t.Errorf("Publish to existing stream should return '%s', got '%s'", StreamAlreadyExists, err)
	}
}

func TestRelayImpl_SubscribeNonExisting(t *testing.T) {
	relay := NewRelay()

	_, _, err := relay.Subscribe("foobar")
	if err != StreamNotExisting {
		t.Errorf("Subscribe to non-existing stream should return '%s', got '%s'", StreamNotExisting, err)
	}
}
