package interfaces

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// BlazeEventMap representa os eventos disponíveis na Blaze
type BlazeEventMap struct {
	CrashTick     *CrashTickEvent     `json:"crash.tick,omitempty"`
	CrashTickBets *CrashTickBetsEvent `json:"crash.tick-bets,omitempty"`
	DoubleTick    *DoubleTickEvent    `json:"double.tick,omitempty"`
	ChatMessage   *ChatMessageEvent   `json:"chat.message,omitempty"`
	Close         *CloseEvent         `json:"close,omitempty"`
	Subscriptions []string            `json:"subscriptions,omitempty"`
}

// CrashTickEvent representa um evento de tick do crash
type CrashTickEvent struct {
	ID           string   `json:"id"`
	UpdatedAt    string   `json:"updated_at"`
	Status       string   `json:"status"`
	CrashPoint   *Float64String `json:"crash_point"`
	IsBonusRound bool     `json:"is_bonus_round"`
}

// CrashTickBetsEvent representa as apostas do crash (apenas crash_2)
type CrashTickBetsEvent struct {
	ID              string  `json:"id"`
	RoomID          int     `json:"roomId"`
	TotalEurBet     float64 `json:"total_eur_bet"`
	TotalBetsPlaced string  `json:"total_bets_placed"`
	TotalEurWon     float64 `json:"total_eur_won"`
	Bets            []Bet   `json:"bets"`
}

// DoubleTickEvent representa um evento de tick do double
type DoubleTickEvent struct {
	ID                   string  `json:"id"`
	Color                *StringOrNumber `json:"color"`
	Roll                 *StringOrNumber `json:"roll"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
	Status               string  `json:"status"` // "rolling", "waiting", "complete"
	TotalRedEurBet       float64 `json:"total_red_eur_bet"`
	TotalRedBetsPlaced   int     `json:"total_red_bets_placed"`
	TotalWhiteEurBet     float64 `json:"total_white_eur_bet"`
	TotalWhiteBetsPlaced int     `json:"total_white_bets_placed"`
	TotalBlackEurBet     float64 `json:"total_black_eur_bet"`
	TotalBlackBetsPlaced int     `json:"total_black_bets_placed"`
	Bets                 []Bet   `json:"bets"`
}

// ChatMessageEvent representa uma mensagem do chat
type ChatMessageEvent struct {
	ID        string   `json:"id"`
	Text      string   `json:"text"`
	Available bool     `json:"available"`
	CreatedAt string   `json:"created_at"`
	User      ChatUser `json:"user"`
}

// ChatUser representa um usuário do chat
type ChatUser struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	Rank     string  `json:"rank"`
	Label    *string `json:"label"`
	Level    int     `json:"level"`
}

// CloseEvent representa um evento de fechamento da conexão
type CloseEvent struct {
	Code      int  `json:"code"`
	Reconnect bool `json:"reconnect"`
}

// Bet representa uma aposta
type Bet struct {
	ID           string   `json:"id"`
	CashedOutAt  *float64 `json:"cashed_out_at"`
	Amount       float64  `json:"amount"`
	CurrencyType string   `json:"currency_type"`
	WinAmount    string   `json:"win_amount"`
	Status       string   `json:"status"` // "win", "created"
}

type Float64String float64

func (f *Float64String) UnmarshalJSON(data []byte) error {
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*f = Float64String(num)
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str == "" {
			*f = 0
			return nil
		}
		n, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		*f = Float64String(n)
		return nil
	}
	return fmt.Errorf("Float64String: cannot unmarshal %s", string(data))
}

type StringOrNumber string

func (s *StringOrNumber) UnmarshalJSON(data []byte) error {
	// Try as string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = StringOrNumber(str)
		return nil
	}
	// Try as number
	var num json.Number
	if err := json.Unmarshal(data, &num); err == nil {
		*s = StringOrNumber(num.String())
		return nil
	}
	return fmt.Errorf("StringOrNumber: cannot unmarshal %s", string(data))
}
