package singleuser

import (
	"context"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/lucas-clemente/quic-go"
)

type NetClientConn struct {
	msgIdCtr     uint32
	l            *zap.SugaredLogger
	stateLock    sync.Mutex
	closed       bool
	session      quic.Session
	stream       quic.Stream
	writerLock   sync.Mutex
	msgListeners map[MessageID]chan<- OptionalMessage
	requests     chan MessageHandle
}

func wrapSessionServer(session quic.Session, maxPipelining int, l *zap.SugaredLogger) (*NetClientConn, error) {
	stream, err := session.OpenStream()
	if err != nil {
		return nil, err
	}

	return &NetClientConn{
		l:            l,
		msgIdCtr:     1,
		closed:       false,
		session:      session,
		stream:       stream,
		msgListeners: make(map[MessageID]chan<- OptionalMessage),
		requests:     make(chan MessageHandle, maxPipelining),
	}, nil
}

func wrapSessionClient(ctx context.Context, session quic.Session, maxPipelining int, l *zap.SugaredLogger) (*NetClientConn, error) {
	stream, err := session.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}

	return &NetClientConn{
		l:            l,
		msgIdCtr:     1,
		closed:       false,
		session:      session,
		stream:       stream,
		msgListeners: make(map[MessageID]chan<- OptionalMessage),
		requests:     make(chan MessageHandle, maxPipelining),
	}, nil
}

func (c *NetClientConn) processReceived(msg Message, id MessageID) error {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	listener, ok := c.msgListeners[id]
	if ok {
		c.l.Debugw("delivering message to listener",
			"messageID", uint64(id),
		)
		listener <- OptionalMessage{Msg: msg}
		close(listener)
		delete(c.msgListeners, id)
		return nil
	}

	req := NewMessageHandle(msg, id)
	if len(c.requests) >= cap(c.requests) {
		c.l.Debugw("pipeline capacity exceeded, returning 503",
			"capacity", cap(c.requests),
		)
		if req.ExpectsReply() {
			c.reply(req.Context(), &ErrorMessage{Message: ErrPipelining.Error(), Code: 503})
		}
		req.Close()
		return nil
	}

	c.l.Debugw("delivering request to channel",
		"messageID", uint64(id),
	)
	c.requests <- req
	return nil
}

func (c *NetClientConn) RecvChannel() <-chan MessageHandle {
	return c.requests
}

func (c *NetClientConn) loop() error {
	for {
		var id MessageID
		msg, err := RecvMessage(c.stream, &id)
		if err == nil {
			// TODO: treat msg.Error as recoverable?
			err = msg.Error
		}
		if err != nil {
			c.l.Debugw("failed to receive message; terminating client",
				"err", err,
			)
			c.send(&ErrorMessage{Message: err.Error(), Code: 400}, MsgIDNone)
			c.terminate(err)
			return err
		}

		if msg.Msg == nil {
			c.l.Warnw("decoded message without error, but no message?!",
				"id", id,
			)
			continue
		}

		err = c.processReceived(msg.Msg, id)
		if err != nil && id != MsgIDNone {
			c.l.Debugw("failed to process message; returning error 500",
				"err", err,
			)
			c.send(&ErrorMessage{Message: err.Error(), Code: 500}, id)
		}
	}
}

func (c *NetClientConn) send(msg Message, id MessageID) error {
	c.writerLock.Lock()
	defer c.writerLock.Unlock()
	return SendMessage(msg, id, c.stream)
}

func (c *NetClientConn) reply(ctx context.Context, msg Message) error {
	id, ok := ctx.Value(contextKeyMessageID).(MessageID)
	if !ok {
		return ErrNoReplyContext
	}

	err := c.send(msg, id)
	c.l.Debugw("reply message sent",
		"messageID", id,
		"err", err,
	)
	return err
}

func (c *NetClientConn) Request(ctx context.Context, msg Message) (<-chan OptionalMessage, error) {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()
	if c.closed {
		return nil, ErrClosed
	}

	replyChan := make(chan OptionalMessage, 1)
	id := MessageID(atomic.AddUint32(&c.msgIdCtr, 1))
	if err := c.send(msg, id); err != nil {
		close(replyChan)
		c.l.Debugw("request message sent",
			"messageID", id,
			"err", err,
		)
		return nil, err
	}
	c.msgListeners[id] = replyChan
	c.l.Debugw("request message sent",
		"messageID", id,
		"err", nil,
	)

	return replyChan, nil
}

func (c *NetClientConn) Reply(ctx context.Context, msg Message) error {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()
	if c.closed {
		return ErrClosed
	}
	return c.reply(ctx, msg)
}

func (c *NetClientConn) OneShot(ctx context.Context, msg Message) error {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()
	if c.closed {
		return ErrClosed
	}
	err := c.send(msg, MsgIDNone)
	c.l.Debugw("oneshot message sent",
		"err", nil,
	)
	return err
}

func (c *NetClientConn) Close() error {
	c.terminate(ErrClosed)
	return nil
}

func (c *NetClientConn) terminate(err error) {
	c.l.Debugw("terminating",
		"err", err,
	)
	c.stateLock.Lock()
	defer c.stateLock.Unlock()
	wasClosed := c.closed
	c.closed = true
	c.writerLock.Lock()
	defer c.writerLock.Unlock()
	for _, listener := range c.msgListeners {
		listener <- OptionalMessage{Msg: nil, Error: ErrClosed}
		close(listener)
	}
	c.msgListeners = nil
	if !wasClosed {
		close(c.requests)
	}
	c.stream.Close()
	c.session.CloseWithError(400, err.Error())
}
