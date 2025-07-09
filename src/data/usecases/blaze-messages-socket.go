package usecases

import (
	"encoding/json"
	"regexp"
	"time"

	domainUsecases "github.com/viniciusgdr/blazego/src/domain/usecases"

	"github.com/viniciusgdr/blazego/src/data/interfaces"
)

type BlazeMessageSocket struct {
	socket    domainUsecases.ConnectionSocket
	callbacks map[string][]func(interface{})
	interval  *time.Ticker
}

func NewBlazeMessageSocket(socket domainUsecases.ConnectionSocket) *BlazeMessageSocket {
	return &BlazeMessageSocket{
		socket:    socket,
		callbacks: make(map[string][]func(interface{})),
	}
}

func (b *BlazeMessageSocket) Connect(options domainUsecases.SocketOptions) error {
	connectionOptions := domainUsecases.ConnectionSocketOptions{
		URL:     options.URL,
		Options: options.Options,
	}

	err := b.socket.Connect(connectionOptions)
	if err != nil {
		return err
	}

	timeoutPing := 10000
	if options.TimeoutPing != nil {
		timeoutPing = *options.TimeoutPing
	}

	b.initPing(timeoutPing)
	b.initOpen(options.Token)
	b.onMessage()
	b.initClose(options)

	return nil
}

func (b *BlazeMessageSocket) initPing(timeoutPing int) {
	b.interval = time.NewTicker(time.Duration(timeoutPing) * time.Millisecond)

	go func() {
		for range b.interval.C {
			b.socket.Send("2")
		}
	}()
}

func (b *BlazeMessageSocket) onMessage() {
	b.socket.On("message", func(data interface{}) {
		var msg string
		switch v := data.(type) {
		case []byte:
			msg = string(v)
		case string:
			msg = v
		default:
			return
		}

		regex := regexp.MustCompile(`^\d+\["data",\s*({.*})]$`)
		matches := regex.FindStringSubmatch(msg)

		if len(matches) < 2 {
			return
		}

		var messageData struct {
			Payload interface{} `json:"payload"`
			ID      string      `json:"id"`
		}

		err := json.Unmarshal([]byte(matches[1]), &messageData)
		if err != nil {
			return
		}

		if messageData.Payload == nil || messageData.ID == "" {
			return
		}

		b.emit(messageData.ID, messageData.Payload)
	})
}

func (b *BlazeMessageSocket) initClose(options domainUsecases.SocketOptions) {
	b.socket.On("close", func(data interface{}) {
		if b.interval != nil {
			b.interval.Stop()
		}

		b.socket.Disconnect()

		code, ok := data.(int)
		if !ok {
			code = 1000
		}

		reconnect := false
		if options.Reconnect != nil {
			reconnect = *options.Reconnect
		}

		if reconnect {
			go func() {
				time.Sleep(100 * time.Millisecond)
				b.Connect(options)
			}()
		}

		closeEvent := interfaces.CloseEvent{
			Code:      code,
			Reconnect: reconnect,
		}

		b.emit("close", closeEvent)
	})
}

func (b *BlazeMessageSocket) initOpen(token *string) {
	b.socket.On("open", func(data interface{}) {
		subscriptions := []string{}

		subscribeMsg := `420["cmd",{"id":"subscribe","payload":{"room":"chat_room_2"}}]`
		b.socket.Send(subscribeMsg)
		subscriptions = append(subscriptions, "chat_room_2")

		b.emit("subscriptions", subscriptions)
	})
}

func (b *BlazeMessageSocket) On(event string, callback func(data interface{})) {
	if b.callbacks[event] == nil {
		b.callbacks[event] = make([]func(interface{}), 0)
	}
	b.callbacks[event] = append(b.callbacks[event], callback)
}

func (b *BlazeMessageSocket) emit(event string, data interface{}) {
	if callbacks, exists := b.callbacks[event]; exists {
		for _, callback := range callbacks {
			go callback(data)
		}
	}
}

func (b *BlazeMessageSocket) Emit(event string, data interface{}) {
	b.emit(event, data)
}

func (b *BlazeMessageSocket) Send(data interface{}) error {
	return b.socket.Send(data)
}

func (b *BlazeMessageSocket) Disconnect() error {
	if b.interval != nil {
		b.interval.Stop()
	}
	return b.socket.Disconnect()
}
