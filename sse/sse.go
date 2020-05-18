package sse

import (
	"encoding/json"
	"fmt"
	"github.com/sony/sonyflake"
	"log"
	"net/http"
	"sync"
	"time"
)

type Event struct {
	ID   uint64
	Message []byte
	// Consumer client's distributed ID
	Consumer uint64
}

func NewEvent(msg []byte, c uint64) Event {
	// Generate unique id for each event
	s := sonyflake.Settings{
		StartTime:      time.Time{},
		MachineID:      nil,
		CheckMachineID: nil,
	}

	f := sonyflake.NewSonyflake(s)

	var id uint64 = 0
	idF, err := f.NextID()
	if err == nil {
		id = idF
	}

	return Event{
		ID:   id,
		Message: msg,
		Consumer: c,
	}
}

func (e Event) String() string {
	return fmt.Sprintf("%d: %s", e.ID, string(e.Message))
}

type Broker struct {
	// consumers subscriber pool using which assigns a Distributed ID for each client
	consumers map[chan Event]uint64
	logger *log.Logger
	mtx *sync.Mutex
}

func NewBroker(logger *log.Logger) *Broker {
	return &Broker{
		consumers: make(map[chan Event]uint64),
		mtx: new(sync.Mutex),
		logger: logger,
	}
}

func (b *Broker) Subscribe() chan Event {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// Generate unique id for each client
	s := sonyflake.Settings{
		StartTime:      time.Time{},
		MachineID:      nil,
		CheckMachineID: nil,
	}

	f := sonyflake.NewSonyflake(s)

	var id uint64 = 0
	idF, err := f.NextID()
	if err == nil {
		id = idF
	}

	c := make(chan Event)
	b.consumers[c] = id

	b.logger.Printf("client %d connected", id)
	return c
}

func (b *Broker) Unsubscribe(c chan Event) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	id := b.consumers[c]
	close(c)
	delete(b.consumers, c)
	b.logger.Printf("client %d killed, %d remaining", id, len(b.consumers))
}

func (b *Broker) Publish(e Event) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	pubMsg := 0
	for s, id := range b.consumers {
		if e.Consumer > 0 {
			// Push to specific consumer
			if id == e.Consumer {
				s <- e
				pubMsg++
				break
			}
		} else {
			// Push to every consumer
			e.Consumer = id
			s <- e
			// Reset unused consumer
			e.Consumer = 0
			pubMsg++
		}
	}

	b.logger.Printf("published message to %d subscribers", pubMsg)
}

// Close Remove channels leftovers
func (b *Broker) Close() {
	for k, _ := range b.consumers {
		close(k)
		delete(b.consumers, k)
	}
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	w.Header().Set("Access-Control-Allow-Origin", "*")

	b.logger.Printf("id: %s", r.URL.Query().Get("id"))

	// Create new client channel for stream events
	c := b.Subscribe()
	defer b.Unsubscribe(c)

	// Send client it's new ID
	go b.Publish(NewEvent(nil, b.consumers[c]))

	for {
		select {
		case msg := <-c:
			// MIME Type (text/event-stream) formatted, DO NOT MODIFY IT
			msgJSON, err := json.Marshal(struct {
				ID uint64 `json:"event_id"`
				Message string `json:"message"`
				Consumer uint64 `json:"consumer_id"`
			}{msg.ID, string(msg.Message), msg.Consumer})
			if err != nil {
				_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
			} else {
				_, _ = fmt.Fprintf(w, "data: %s\n\n", msgJSON)
			}
			f.Flush()
		case <-ctx.Done():
			return
		}
	}
}
