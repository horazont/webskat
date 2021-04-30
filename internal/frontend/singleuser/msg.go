package singleuser

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/horazont/webskat/internal/replay"
	"github.com/horazont/webskat/internal/skat"
)

const (
	MsgPing      MessageType = 0x0000
	MsgPong      MessageType = 0x0001
	MsgAction    MessageType = 0x0002
	MsgPollState MessageType = 0x0003
	MsgState     MessageType = 0x0004
	MsgError     MessageType = 0x0005
	MsgLoginReq  MessageType = 0x0006
	MsgLoginOk   MessageType = 0x0007
	MsgAck       MessageType = 0x0008
)

type PingPongMessage struct {
	isPong bool
}

func NewPing() Message {
	return &PingPongMessage{isPong: false}
}

func NewPong() Message {
	return &PingPongMessage{isPong: true}
}

func (m *PingPongMessage) Type() MessageType {
	if m.isPong {
		return MsgPong
	} else {
		return MsgPing
	}
}

type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewErrorMessage(code int, message string) *ErrorMessage {
	return &ErrorMessage{Code: code, Message: message}
}

func (m *ErrorMessage) Type() MessageType {
	return MsgError
}

func (m *ErrorMessage) Error() string {
	return fmt.Sprintf("[%d] %s", m.Code, m.Message)
}

type LoginRequestMessage struct {
	ServerPassword string `json:"serverPassword"`
	ClientID       string `json:"clientId"`
	ClientSecret   string `json:"clientSecret"`
}

func (m *LoginRequestMessage) Type() MessageType {
	return MsgLoginReq
}

type LoginOkMessage struct {
}

func (m *LoginOkMessage) Type() MessageType {
	return MsgLoginOk
}

type ActionMessage struct {
	Action        json.RawMessage `json:"action"`
	decodedAction replay.Action
	decodeError   error
}

func NewActionMessage(action replay.Action) (*ActionMessage, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := replay.ActionToJSON(action, enc)
	if err != nil {
		return nil, err
	}

	return &ActionMessage{
		Action: buf.Bytes(),
	}, nil
}

func (m *ActionMessage) Type() MessageType {
	return MsgAction
}

func (m *ActionMessage) DecodePayload() error {
	if m.decodeError != nil || m.decodedAction != nil {
		return m.decodeError
	}

	reader := bytes.NewReader(m.Action)
	dec := json.NewDecoder(reader)
	m.decodedAction, m.decodeError = replay.ActionFromJSON(dec)
	return m.decodeError
}

func (m *ActionMessage) Payload() (replay.Action, error) {
	if m.decodedAction == nil && m.decodeError == nil {
		m.DecodePayload()
	}
	return m.decodedAction, m.decodeError
}

type AckMessage struct {
}

func (m *AckMessage) Type() MessageType {
	return MsgAck
}

type StateMessage struct {
	YourPlayerIndex int                    `json:"playerIndex"`
	GameState       *skat.BlindedGameState `json:"gameState"`
}

func NewStateMessage(playerIndex int, gameState *skat.BlindedGameState) *StateMessage {
	return &StateMessage{
		YourPlayerIndex: playerIndex,
		GameState:       gameState,
	}
}

func (m *StateMessage) Type() MessageType {
	return MsgState
}
