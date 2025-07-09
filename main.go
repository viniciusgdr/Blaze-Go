package blaze

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"viniciusgdr/blaze/src/data/interfaces"
	"viniciusgdr/blaze/src/data/usecases"
	domainUsecases "viniciusgdr/blaze/src/domain/usecases"
	"viniciusgdr/blaze/src/infra/app"
	"viniciusgdr/blaze/src/infra/blaze"
)

type ConnectionBlaze struct {
	Web  string // "blaze" or "blaze-chat"
	Type string // "crash", "doubles", "crash_2", "crash_neymarjr"
}

type Connection struct {
	URL                       *string
	Type                      *string
	Token                     *string
	Options                   *ConnectionOptions
	TimeoutPing               *int
	CacheIgnoreRepeatedEvents *bool
	Web                       string
	GameType                  string
}

type ConnectionOptions struct {
	Host    *string
	Origin  *string
	Headers map[string]string
}

type ConnectionSocketResponses interface {
	Connect(options domainUsecases.SocketOptions) error
	On(event string, callback func(data interface{}))
	Emit(event string, data interface{})
	Send(data interface{}) error
	Disconnect() error
}

func MakeConnection(conn Connection) (ConnectionSocketResponses, error) {
	switch conn.Web {
	case "blaze":
		var url string
		if conn.URL != nil {
			url = *conn.URL
		} else {
			url = blaze.GetBlazeURL("games")
		}

		headers := map[string]string{
			"Upgrade":                  "websocket",
			"Sec-Websocket-Extensions": "permessage-deflate; client_max_window_bits",
			"Pragma":                   "no-cache",
			"Connection":               "Upgrade",
			"Accept-Encoding":          "gzip, deflate, br",
			"User-Agent":               "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36",
		}

		if conn.Options != nil && conn.Options.Headers != nil {
			for k, v := range conn.Options.Headers {
				headers[k] = v
			}
		}

		host := "api-v2.blaze1.space"
		origin := "https://api-gaming.blaze.com"

		if conn.Options != nil {
			if conn.Options.Host != nil {
				host = *conn.Options.Host
			}
			if conn.Options.Origin != nil {
				origin = *conn.Options.Origin
			}
		}

		socketOptions := domainUsecases.SocketOptions{
			URL:   &url,
			Type:  &conn.GameType,
			Token: conn.Token,
			Options: &domainUsecases.ConnectionSocketOpts{
				Host:    &host,
				Origin:  &origin,
				Headers: headers,
			},
			TimeoutPing: conn.TimeoutPing,
		}

		socket := app.NewNodeConnectionSocket()

		cacheIgnoreRepeatedEvents := true
		if conn.CacheIgnoreRepeatedEvents != nil {
			cacheIgnoreRepeatedEvents = *conn.CacheIgnoreRepeatedEvents
		}

		blazeSocket := usecases.NewBlazeSocket(socket, cacheIgnoreRepeatedEvents)
		err := blazeSocket.Connect(socketOptions)
		if err != nil {
			return nil, err
		}

		return blazeSocket, nil

	case "blaze-chat":
		var url string
		if conn.URL != nil {
			url = *conn.URL
		} else {
			url = blaze.GetBlazeURL("general")
		}

		headers := map[string]string{
			"Upgrade":                  "websocket",
			"Sec-Websocket-Extensions": "permessage-deflate; client_max_window_bits",
			"Pragma":                   "no-cache",
			"Connection":               "Upgrade",
			"Accept-Encoding":          "gzip, deflate, br",
			"User-Agent":               "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36",
		}

		if conn.Options != nil && conn.Options.Headers != nil {
			maps.Copy(headers, conn.Options.Headers)
		}

		host := "api-v2.blaze1.space"
		origin := "https://api-gaming.blaze.com"

		if conn.Options != nil {
			if conn.Options.Host != nil {
				host = *conn.Options.Host
			}
			if conn.Options.Origin != nil {
				origin = *conn.Options.Origin
			}
		}

		socketOptions := domainUsecases.SocketOptions{
			URL:   &url,
			Type:  &conn.GameType,
			Token: conn.Token,
			Options: &domainUsecases.ConnectionSocketOpts{
				Host:    &host,
				Origin:  &origin,
				Headers: headers,
			},
			TimeoutPing: conn.TimeoutPing,
		}

		socketForMessages := app.NewNodeConnectionSocket()
		blazeSocketForMessages := usecases.NewBlazeMessageSocket(socketForMessages)
		err := blazeSocketForMessages.Connect(socketOptions)
		if err != nil {
			return nil, err
		}

		return blazeSocketForMessages, nil

	default:
		return nil, fmt.Errorf("missing web")
	}
}

type GameEventResult struct {
	Events []interfaces.CrashTickEvent `json:"events"`
	Error  error                       `json:"error,omitempty"`
}

// GetNextGameEventTick aguarda o próximo jogo completo e retorna todos os eventos
// 1. Conecta ao crash
// 2. Aguarda status "waiting" (início do próximo jogo)
// 3. Coleta todos os eventos crash.tick dessa rodada
// 4. Quando status for "complete", encerra a conexão e retorna os eventos
func GetNextGameEventTick(gameType string) (<-chan any, <-chan error) {
	return GetNextGameEventTickWithContext(context.Background(), gameType)
}

// GetNextGameEventTickWithContext agora retorna um canal de eventos em tempo real e um canal de erro.
// O canal de eventos envia cada CrashTickEvent assim que recebido.
// O canal de erro envia qualquer erro ocorrido durante o processo.
// O canal de eventos é fechado quando o jogo termina ou ocorre erro/cancelamento.
func GetNextGameEventTickWithContext(ctx context.Context, gameType string) (<-chan any, <-chan error) {
	eventChan := make(chan any)
	errorChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		defer close(errorChan)

		conn, err := MakeConnection(Connection{
			GameType: gameType,
			Web:      "blaze",
		})
		if err != nil {
			errorChan <- fmt.Errorf("erro ao conectar: %w", err)
			return
		}

		gameStarted := false

		conn.On("crash.tick", func(data interface{}) {
			dataBytes, err := json.Marshal(data)
			if err != nil {
				errorChan <- err
				return
			}

			var tickEvent interfaces.CrashTickEvent
			if err := json.Unmarshal(dataBytes, &tickEvent); err != nil {
				errorChan <- err
				return
			}

			if !gameStarted {
				if tickEvent.Status == "waiting" {
					gameStarted = true
					eventChan <- tickEvent
				}
				return
			}

			eventChan <- tickEvent

			if tickEvent.Status == "complete" {
				go func() {
					conn.Disconnect()
				}()
			}
		})

		conn.On("double.tick", func(data interface{}) {
			dataBytes, err := json.Marshal(data)
			if err != nil {
				errorChan <- err
				return
			}

			var tickEvent interfaces.DoubleTickEvent
			if err := json.Unmarshal(dataBytes, &tickEvent); err != nil {
				errorChan <- err
				return
			}

			if !gameStarted {
				if tickEvent.Status == "waiting" {
					gameStarted = true
					eventChan <- tickEvent
				}
				return
			}

			eventChan <- tickEvent

			if tickEvent.Status == "complete" {
				go func() {
					conn.Disconnect()
				}()
			}
		})

		conn.On("close", func(data interface{}) {
			if !gameStarted {
				errorChan <- fmt.Errorf("connection closed before game started")
			}
		})

		<-ctx.Done()
		conn.Disconnect()
		errorChan <- ctx.Err()
	}()

	return eventChan, errorChan
}
