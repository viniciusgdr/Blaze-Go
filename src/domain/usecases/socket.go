package usecases

type SocketOptions struct {
	URL         *string               `json:"url,omitempty"`
	Type        *string               `json:"type,omitempty"`
	Token       *string               `json:"token,omitempty"`
	Reconnect   *bool                 `json:"reconnect,omitempty"`
	Options     *ConnectionSocketOpts `json:"options,omitempty"`
	TimeoutPing *int                  `json:"timeoutPing,omitempty"`
}

type GenericSocket[T any] interface {
	On(event string, callback func(data T))
	Emit(event string, data interface{})
}

type Socket[T any] interface {
	GenericSocket[T]
	Connect(options SocketOptions) error
	Send(data interface{}) error
	Disconnect() error
}

type SocketEvents struct {
	Subscriptions []string    `json:"subscriptions,omitempty"`
	Close         *CloseEvent `json:"close,omitempty"`
}

type CloseEvent struct {
	Code int `json:"code"`
}
