package usecases

type ConnectionSocketOptions struct {
	URL     *string               `json:"url,omitempty"`
	Options *ConnectionSocketOpts `json:"options,omitempty"`
}

type ConnectionSocketOpts struct {
	Host    *string           `json:"host,omitempty"`
	Origin  *string           `json:"origin,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type ConnectionSocket interface {
	Connect(options ConnectionSocketOptions) error
	On(event string, callback func(data interface{}))
	Emit(event string, data interface{})
	Send(data interface{}) error
	Disconnect() error
}
