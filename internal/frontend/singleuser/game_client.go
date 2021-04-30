package singleuser

import (
	"context"

	"go.uber.org/zap"

	"github.com/horazont/webskat/internal/replay"
	"github.com/horazont/webskat/internal/skat"
)

type GameClient struct {
	l        *zap.SugaredLogger
	conn     MessageEndpoint
	clientID string
	quit     chan struct{}

	states       chan ClientState
	memoizedSeed []byte
}

type ClientState struct {
	PlayerIndex int
	GameState   *skat.BlindedGameState
}

func NewGameClient(l *zap.SugaredLogger, ctx context.Context, conn MessageEndpoint) (*GameClient, error) {
	if err := exchangeInitialPingPongClient(ctx, conn); err != nil {
		l.Debugw("initial ping/pong failed",
			"err", err,
		)
		return nil, err
	}

	result := &GameClient{
		l:      l,
		conn:   conn,
		quit:   make(chan struct{}, 0),
		states: make(chan ClientState, 1),
	}
	go result.loop()
	return result, nil
}

func (c *GameClient) handleMessage(msgH MessageHandle) {
	var reply Message
	defer msgH.Close()

	msg := msgH.Message()
	type_ := msg.Type()
	switch type_ {
	case MsgState:
		{
			stateMsg := msg.(*StateMessage)
			c.l.Debugw("received state update",
				"state", stateMsg.GameState,
			)
			if len(c.states) >= cap(c.states) {
				// drop one state from the queue in case the recipient is
				// overloaded
				select {
				case _, ok := <-c.states:
					if ok {
						c.l.Warnw("state update not consumed")
					}
				default:
				}
			}
			c.states <- ClientState{
				PlayerIndex: stateMsg.YourPlayerIndex,
				GameState:   stateMsg.GameState,
			}
		}
	case MsgPing:
		reply = NewPong()
	case MsgPong, MsgAck:
		c.l.Debugw("stray message?!",
			"messageType", type_,
		)
		reply = NewErrorMessage(400, "stray message")
	default:
		c.l.Debugw("not implemented",
			"messageType", type_,
		)
		reply = NewErrorMessage(500, "not implemented")
	}

	if reply != nil {
		if !msgH.ExpectsReply() {
			c.l.Debugw("discarding reply to oneshot message")
		} else {
			err := c.conn.Reply(msgH.Context(), reply)
			if err != nil {
				c.l.Warnw("failed to send reply",
					"err", err,
				)
			}
		}
	} else if msgH.ExpectsReply() {
		c.l.Fatalw("no reply?!",
			"messageType", type_,
			"message", msg,
		)
	}
}

func (c *GameClient) loop() error {
	for {
		select {
		case msgH, ok := <-c.conn.RecvChannel():
			if !ok {
				return ErrClosed
			}
			c.handleMessage(msgH)
		case <-c.quit:
			return nil
		}
	}
}

func (c *GameClient) Close() {
	close(c.quit)
	c.clientID = ""
}

func (c *GameClient) Login(ctx context.Context, clientID string, clientSecret string, serverPassword string) error {
	login := &LoginRequestMessage{
		ServerPassword: serverPassword,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
	}
	replyChan, err := c.conn.Request(ctx, login)
	if err != nil {
		return err
	}

	msgWrapper, ok := <-replyChan
	if !ok {
		return ErrClosed
	}

	if msgWrapper.Error != nil {
		return msgWrapper.Error
	}

	msg := msgWrapper.Msg
	switch msg.Type() {
	case MsgError:
		return msg.(*ErrorMessage)
	case MsgLoginOk:
	default:
		return ErrProtocolViolation
	}

	c.clientID = clientID
	return nil
}

func (c *GameClient) NetPing(ctx context.Context) error {
	ping := NewPing()
	replyChan, err := c.conn.Request(ctx, ping)

	if err != nil {
		return err
	}

	var msgW OptionalMessage
	var ok bool

	select {
	case <-ctx.Done():
		return ctx.Err()
	case msgW, ok = <-replyChan:
		if !ok {
			return ErrClosed
		}
	}

	if msgW.Error != nil {
		return msgW.Error
	}

	reply := msgW.Msg
	switch reply.Type() {
	case MsgPong:
		return nil
	case MsgError:
		return reply.(*ErrorMessage)
	default:
		return ErrProtocolViolation
	}
}

func (c *GameClient) sendAction(ctx context.Context, action replay.Action) error {
	req, err := NewActionMessage(action)
	if err != nil {
		return err
	}
	replyChan, err := c.conn.Request(ctx, req)
	if err != nil {
		return err
	}

	var msgW OptionalMessage
	var ok bool

	select {
	case <-ctx.Done():
		return ctx.Err()
	case msgW, ok = <-replyChan:
		if !ok {
			return ErrClosed
		}
	}

	if msgW.Error != nil {
		return msgW.Error
	}

	reply := msgW.Msg
	switch reply.Type() {
	case MsgAck:
		return nil
	case MsgError:
		return reply.(*ErrorMessage)
	default:
		return ErrProtocolViolation
	}
}

func (c *GameClient) SetSeed(ctx context.Context, seed []byte) error {
	return c.sendAction(
		ctx,
		&replay.ActionSetSeed{
			Seed: seed,
		},
	)
}

func (c *GameClient) CallBid(ctx context.Context, value int) error {
	return c.sendAction(
		ctx,
		&replay.ActionCallBid{
			Value: value,
		},
	)
}

func (c *GameClient) ReplyToBid(ctx context.Context, hold bool) error {
	return c.sendAction(
		ctx,
		&replay.ActionReplyToBid{
			Hold: hold,
		},
	)
}

func (c *GameClient) TakeSkat(ctx context.Context) error {
	return c.sendAction(
		ctx,
		&replay.ActionTakeSkat{},
	)
}

func (c *GameClient) Declare(ctx context.Context, gameType skat.GameType, announcedModifiers skat.GameModifier, cardsToPush skat.CardSet) error {
	return c.sendAction(
		ctx,
		&replay.ActionDeclare{
			GameType:          gameType,
			AnnounceModifiers: announcedModifiers,
			CardsToPush:       cardsToPush,
		},
	)
}

func (c *GameClient) PlayCard(ctx context.Context, card skat.Card) error {
	return c.sendAction(
		ctx,
		&replay.ActionPlayCard{
			Card: card,
		},
	)
}

func (c *GameClient) StateChannel() <-chan ClientState {
	return c.states
}

func exchangeInitialPingPongClient(ctx context.Context, ep MessageEndpoint) error {
	var msgH MessageHandle
	var ok bool

	select {
	case <-ctx.Done():
		return ctx.Err()
	case msgH, ok = <-ep.RecvChannel():
		if !ok {
			return ErrClosed
		}
	}

	if msgH.Message().Type() != MsgPing {
		return ErrProtocolViolation
	}

	return ep.Reply(msgH.Context(), NewPong())
}
