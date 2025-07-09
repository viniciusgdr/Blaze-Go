package app

import (
	"errors"
	"net/http"
	"net/url"

	domainUsecases "viniciusgdr/blaze/src/domain/usecases"

	"github.com/gorilla/websocket"
)

type NodeConnectionSocket struct {
	conn      *websocket.Conn
	callbacks map[string][]func(interface{})
}

func NewNodeConnectionSocket() *NodeConnectionSocket {
	return &NodeConnectionSocket{
		callbacks: make(map[string][]func(interface{})),
	}
}

func (n *NodeConnectionSocket) Connect(options domainUsecases.ConnectionSocketOptions) error {
	if options.URL == nil {
		return errors.New("missing url")
	}

	u, err := url.Parse(*options.URL)
	if err != nil {
		return err
	}

	headers := http.Header{}
	if options.Options != nil && options.Options.Headers != nil {
		for key, value := range options.Options.Headers {
			if key != "Upgrade" && key != "Connection" && key != "Sec-WebSocket-Key" &&
				key != "Sec-WebSocket-Version" && key != "Sec-Websocket-Extensions" &&
				key != "Sec-WebSocket-Extensions" {
				headers.Set(key, value)
			}
		}
	}

	if headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36")
	}

	dialer := websocket.Dialer{}

	conn, _, err := dialer.Dial(u.String(), headers)
	if err != nil {
		return err
	}

	n.conn = conn

	go n.listen()

	n.emit("open", nil)

	return nil
}

func (n *NodeConnectionSocket) listen() {
	defer func() {
		if n.conn != nil {
			n.conn.Close()
		}
	}()

	for {
		if n.conn == nil {
			break
		}

		_, message, err := n.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				n.emit("error", err)
			}
			n.emit("close", websocket.CloseGoingAway)
			break
		}

		n.emit("message", message)
	}
}

func (n *NodeConnectionSocket) On(event string, callback func(data interface{})) {
	if n.callbacks[event] == nil {
		n.callbacks[event] = make([]func(interface{}), 0)
	}
	n.callbacks[event] = append(n.callbacks[event], callback)
}

func (n *NodeConnectionSocket) emit(event string, data interface{}) {
	if callbacks, exists := n.callbacks[event]; exists {
		for _, callback := range callbacks {
			go callback(data)
		}
	}
}

func (n *NodeConnectionSocket) Emit(event string, data interface{}) {
	n.emit(event, data)
}

func (n *NodeConnectionSocket) Send(data interface{}) error {
	if n.conn == nil {
		return errors.New("missing socket")
	}

	var message []byte
	switch v := data.(type) {
	case string:
		message = []byte(v)
	case []byte:
		message = v
	default:
		return errors.New("unsupported data type")
	}

	return n.conn.WriteMessage(websocket.TextMessage, message)
}

func (n *NodeConnectionSocket) Disconnect() error {
	if n.conn == nil {
		return errors.New("missing socket")
	}

	err := n.conn.Close()
	n.conn = nil
	return err
}
