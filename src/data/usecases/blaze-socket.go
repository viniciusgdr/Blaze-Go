package usecases

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"viniciusgdr/blaze/src/data/interfaces"
	domainUsecases "viniciusgdr/blaze/src/domain/usecases"
)

type BlazeSocket struct {
	socket                    domainUsecases.ConnectionSocket
	callbacks                 map[string][]func(interface{})
	cache                     map[string]interface{}
	interval                  *time.Ticker
	cacheIgnoreRepeatedEvents bool
}

func NewBlazeSocket(socket domainUsecases.ConnectionSocket, cacheIgnoreRepeatedEvents bool) *BlazeSocket {
	blazeSocket := &BlazeSocket{
		socket:                    socket,
		callbacks:                 make(map[string][]func(interface{})),
		cacheIgnoreRepeatedEvents: cacheIgnoreRepeatedEvents,
	}

	if cacheIgnoreRepeatedEvents {
		blazeSocket.cache = make(map[string]interface{})
	}

	return blazeSocket
}

func (b *BlazeSocket) Connect(options domainUsecases.SocketOptions) error {
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

	socketType := "crash"
	if options.Type != nil {
		socketType = *options.Type
	}

	b.initOpen(socketType, options.Token)
	b.onMessage()
	b.initClose(options)

	return nil
}

func (b *BlazeSocket) initPing(timeoutPing int) {
	b.interval = time.NewTicker(time.Duration(timeoutPing) * time.Millisecond)

	go func() {
		for range b.interval.C {
			b.socket.Send("2")
		}
	}()
}

func (b *BlazeSocket) onMessage() {
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

		payloadMap, ok := messageData.Payload.(map[string]interface{})
		if !ok {
			return
		}

		payloadID, hasID := payloadMap["id"]
		payloadStatus, hasStatus := payloadMap["status"]

		if !hasID || !hasStatus {
			return
		}

		if b.cache != nil {
			b.emit(fmt.Sprintf("CB:%s", messageData.ID), messageData.Payload)

			if cachedStatus, exists := b.cache[payloadID.(string)]; exists {
				if cachedStatus != payloadStatus {
					b.emit(messageData.ID, messageData.Payload)
					b.cache[payloadID.(string)] = payloadStatus
				}
			} else {
				b.emit(messageData.ID, messageData.Payload)
				b.cache[payloadID.(string)] = payloadStatus
			}
			return
		}

		b.emit(messageData.ID, messageData.Payload)
	})
}

func (b *BlazeSocket) initClose(options domainUsecases.SocketOptions) {
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

func (b *BlazeSocket) initOpen(socketType string, token *string) {
	subscriptions := []string{}

	roomMap := map[string]string{
		"crash":          "crash_room_4",
		"doubles":        "double_room_1",
		"crash_2":        "crash_room_1",
		"crash_neymarjr": "crash_room_3",
	}

	room, exists := roomMap[socketType]
	if !exists {
		b.emit("error", errors.New("missing type of socket"))
		return
	}

	subscribeMsg := fmt.Sprintf(`420["cmd",{"id":"subscribe","payload":{"room":"%s"}}]`, room)
	b.socket.Send(subscribeMsg)
	subscriptions = append(subscriptions, room)

	if token != nil {
		authMsg1 := fmt.Sprintf(`423["cmd",{"id":"authenticate","payload":{"token":"%s"}}]`, *token)
		authMsg2 := fmt.Sprintf(`422["cmd",{"id":"authenticate","payload":{"token":"%s"}}]`, *token)
		authMsg3 := fmt.Sprintf(`420["cmd",{"id":"authenticate","payload":{"token":"%s"}}]`, *token)

		b.socket.Send(authMsg1)
		b.socket.Send(authMsg2)
		b.socket.Send(authMsg3)
	}

	b.emit("subscriptions", subscriptions)
}

func (b *BlazeSocket) On(event string, callback func(data interface{})) {
	if b.callbacks[event] == nil {
		b.callbacks[event] = make([]func(interface{}), 0)
	}
	b.callbacks[event] = append(b.callbacks[event], callback)
}

func (b *BlazeSocket) emit(event string, data interface{}) {
	if callbacks, exists := b.callbacks[event]; exists {
		for _, callback := range callbacks {
			go callback(data)
		}
	}
}

func (b *BlazeSocket) Emit(event string, data interface{}) {
	b.emit(event, data)
}

func (b *BlazeSocket) Send(data interface{}) error {
	return b.socket.Send(data)
}

func (b *BlazeSocket) Disconnect() error {
	if b.interval != nil {
		b.interval.Stop()
	}
	return b.socket.Disconnect()
}
