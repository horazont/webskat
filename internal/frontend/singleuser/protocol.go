package singleuser

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

type MessageType uint16
type MessageID uint32
type RequestReplyContextKey int

const (
	MsgIDNone = 0x00000000
)

const (
	ProtocolName = "net.zombofant.webskat"
)

const (
	contextKeyMessageID RequestReplyContextKey = 1
)

var (
	ErrInternalCommunicationError = errors.New("internal communication error")
	ErrUnknownMessageType         = errors.New("unknown message type")
	ErrMessageTooLong             = errors.New("message too long")
	ErrProtocolViolation          = errors.New("protocol violation")
	ErrWrongVersion               = errors.New("wrong version")
)

const (
	MaxMessageSize int = 65535

	protocolFrame     uint16 = 0x2342
	protocolVersionID uint8  = 0
)

type Message interface {
	Type() MessageType
}

type OptionalMessage struct {
	Msg   Message
	Error error
}

type MessageEndpoint interface {
	Request(ctx context.Context, msg Message) (<-chan OptionalMessage, error)
	Reply(ctx context.Context, msg Message) error
	OneShot(ctx context.Context, msg Message) error
	RecvChannel() <-chan MessageHandle
	Close() error
}

type MessageHandle interface {
	Message() Message
	ExpectsReply() bool
	Context() context.Context
	Close()
}

type messageHandle struct {
	msg          Message
	ctx          context.Context
	cancel       context.CancelFunc
	expectsReply bool
}

func NewMessageHandle(msg Message, id MessageID) MessageHandle {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, contextKeyMessageID, id)
	return &messageHandle{
		msg:          msg,
		ctx:          ctx,
		cancel:       cancel,
		expectsReply: id != MsgIDNone,
	}
}

func (r *messageHandle) Message() Message {
	return r.msg
}

func (r *messageHandle) Context() context.Context {
	return r.ctx
}

func (r *messageHandle) ExpectsReply() bool {
	return r.expectsReply
}

func (r *messageHandle) Close() {
	r.cancel()
	r.msg = nil
	r.ctx = nil
}

func RequestResponse(ep MessageEndpoint, ctx context.Context, msg Message) (Message, error) {
	replyChan, err := ep.Request(ctx, msg)
	if err != nil {
		return nil, err
	}

	select {
	case reply, ok := <-replyChan:
		{
			if !ok {
				return nil, ErrInternalCommunicationError
			}
			if reply.Error != nil {
				return nil, reply.Error
			}
			return reply.Msg, nil
		}
	case <-ctx.Done():
		{
			return nil, ctx.Err()
		}
	}
}

func ParseMessage(msgType MessageType, dec *json.Decoder) (Message, error) {
	switch msgType {
	case MsgPing, MsgPong:
		msg := &PingPongMessage{
			isPong: msgType == MsgPong,
		}
		if err := dec.Decode(msg); err != nil {
			return nil, err
		}
		return msg, nil
	case MsgLoginReq:
		msg := &LoginRequestMessage{}
		if err := dec.Decode(msg); err != nil {
			return nil, err
		}
		return msg, nil
	case MsgLoginOk:
		msg := &LoginOkMessage{}
		if err := dec.Decode(msg); err != nil {
			return nil, err
		}
		return msg, nil
	case MsgError:
		msg := &ErrorMessage{}
		if err := dec.Decode(msg); err != nil {
			return nil, err
		}
		return msg, nil
	case MsgAction:
		msg := &ActionMessage{}
		if err := dec.Decode(msg); err != nil {
			return nil, err
		}
		if err := msg.DecodePayload(); err != nil {
			return nil, err
		}
		return msg, nil
	case MsgAck:
		msg := &AckMessage{}
		if err := dec.Decode(msg); err != nil {
			return nil, err
		}
		return msg, nil
	case MsgState:
		msg := &StateMessage{}
		if err := dec.Decode(msg); err != nil {
			return nil, err
		}
		return msg, nil
	default:
		{
			return nil, ErrUnknownMessageType
		}
	}
}

func SendMessage(msg Message, id MessageID, w io.Writer) (err error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err = enc.Encode(msg)
	if err != nil {
		return err
	}

	if buf.Len() > MaxMessageSize {
		return ErrMessageTooLong
	}

	var u8 uint8
	var u16 uint16
	var u32 uint32

	u16 = protocolFrame
	if err = binary.Write(w, binary.LittleEndian, &u16); err != nil {
		return err
	}

	u8 = protocolVersionID
	if err = binary.Write(w, binary.LittleEndian, &u8); err != nil {
		return err
	}

	u16 = uint16(msg.Type())
	if err = binary.Write(w, binary.LittleEndian, &u16); err != nil {
		return err
	}

	u32 = uint32(id)
	if err = binary.Write(w, binary.LittleEndian, &u32); err != nil {
		return err
	}

	u16 = uint16(buf.Len())
	if err = binary.Write(w, binary.LittleEndian, &u16); err != nil {
		return err
	}

	_, err = buf.WriteTo(w)
	return err
}

func RecvMessage(r io.Reader, id *MessageID) (msg OptionalMessage, err error) {
	var u8 uint8
	var u16 uint16
	var u32 uint32
	var msgType MessageType

	if err = binary.Read(r, binary.LittleEndian, &u16); err != nil {
		return msg, err
	}
	if u16 != protocolFrame {
		return msg, ErrProtocolViolation
	}

	if err = binary.Read(r, binary.LittleEndian, &u8); err != nil {
		return msg, err
	}
	if u8 != protocolVersionID {
		return msg, ErrWrongVersion
	}

	if err = binary.Read(r, binary.LittleEndian, &u16); err != nil {
		return msg, err
	}
	msgType = MessageType(u16)

	if err = binary.Read(r, binary.LittleEndian, &u32); err != nil {
		return msg, err
	}
	*id = MessageID(u32)

	if err = binary.Read(r, binary.LittleEndian, &u16); err != nil {
		return msg, err
	}

	if int(u16) > MaxMessageSize {
		return msg, ErrMessageTooLong
	}

	limitedReader := &io.LimitedReader{
		R: r,
		N: int64(u16),
	}

	dec := json.NewDecoder(limitedReader)
	parsedMsg, msgErr := ParseMessage(msgType, dec)
	return OptionalMessage{
		Msg:   parsedMsg,
		Error: msgErr,
	}, nil
}
