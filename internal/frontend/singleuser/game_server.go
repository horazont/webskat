package singleuser

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"go.uber.org/zap"

	"github.com/horazont/webskat/internal/replay"
	"github.com/horazont/webskat/internal/skat"
)

var (
	ErrPlayerNotFound = errors.New("player not found")
)

type gameClientConn struct {
	clientSecret string
	ep           MessageEndpoint
	playerIndex  int
}

type GameServer struct {
	l                *zap.SugaredLogger
	stateLock        sync.Mutex
	clients          map[string]*gameClientConn
	playerReverseMap []string
	wakeup           chan struct{}
	quit             chan struct{}

	serverPassword      string
	currentGame         *skat.GameState
	currentPlayerOffset int
}

type GameServerConfig struct {
	ServerPassword string
}

func NewGameServer(cfg GameServerConfig, l *zap.SugaredLogger) (*GameServer, error) {
	game, err := skat.NewGame(false, skat.StandardScoreDefinition())
	if err != nil {
		return nil, err
	}

	return &GameServer{
		l:                l,
		clients:          make(map[string]*gameClientConn),
		playerReverseMap: make([]string, 0),
		wakeup:           make(chan struct{}, 1),
		quit:             make(chan struct{}, 0),
		serverPassword:   cfg.ServerPassword,
		currentGame:      game,
	}, nil
}

func (s *GameServer) getValidEndpoints() ([]string, []MessageEndpoint) {
	ids := make([]string, len(s.clients))
	eps := make([]MessageEndpoint, len(s.clients))
	i := 0
	for clientID, clientInfo := range s.clients {
		ep := clientInfo.ep
		if ep == nil {
			continue
		}
		ids[i] = clientID
		eps[i] = ep
		i += 1
	}
	ids = ids[:i]
	eps = eps[:i]
	return ids, eps
}

func (s *GameServer) prepareSelectCases() ([]reflect.SelectCase, []string, []MessageEndpoint, int, int) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	clientIDs, messageEndpoints := s.getValidEndpoints()
	caseWakeup := len(clientIDs)
	caseQuit := caseWakeup + 1
	nCases := caseQuit + 1
	selectCases := make([]reflect.SelectCase, nCases)
	for i, clientID := range clientIDs {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(s.clients[clientID].ep.RecvChannel()),
		}
	}
	selectCases[caseWakeup] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(s.wakeup),
	}
	selectCases[caseQuit] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(s.quit),
	}

	return selectCases, clientIDs, messageEndpoints, caseWakeup, caseQuit
}

func (s *GameServer) terminateClientByChannel(clientID string, expectedEp MessageEndpoint) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	client, ok := s.clients[clientID]
	if !ok {
		return
	}
	if client.ep != expectedEp {
		return
	}
	s.l.Debugw("closing and clearing endpoint",
		"clientID", clientID,
	)
	client.ep.Close()
	client.ep = nil
}

func (s *GameServer) getNextMessage() (string, MessageHandle, MessageEndpoint) {
	for {
		selectCases, clientIDs, endpoints, caseWakeup, caseQuit := s.prepareSelectCases()
		s.l.Debugw(
			"configured listeners for next message",
			"nlisteners", len(selectCases),
		)
		caseIndex, recv, recvOk := reflect.Select(selectCases)
		switch caseIndex {
		case caseWakeup:
			s.l.Debugw("got wakeup signal, reconfiguring select")
		case caseQuit:
			s.l.Debug("returning nil message after quit signal")
			return "", nil, nil
		default:
			clientID := clientIDs[caseIndex]
			ep := endpoints[caseIndex]
			if !recvOk {
				s.l.Warnw(
					"reception from client channel unsuccessful; disconnecting",
					"clientID", clientID,
				)
				// We have to match on the channel, otherwise there is a race
				// condition if the client is reconnecting before we notice
				// the disconnect.
				// getNextMessage() would then be woken up by the .Close() on
				// the endpoint from the ConnectClient() method. However, at
				// that point the ConnectClient() method may have already
				// replaced the endpoint, so we’re terminating the wrong (new)
				// endpoint here.
				// To avoid that, we guard the termination by the given
				// endpoint.

				// We cannot compare the current client state against the
				// stored endpoint here, because we do not hold the state
				// lock.
				s.terminateClientByChannel(clientID, ep)
				continue
			}
			message, okMessage := recv.Interface().(MessageHandle)
			if !okMessage {
				s.l.Debugw(
					"discarding non-Request message",
					"clientID", clientID,
					"rawMessage", recv,
				)
				continue
			}
			s.l.Debugw("received message from client",
				"clientID", clientID,
				"message", message,
			)
			return clientID, message, ep
		}
	}
}

func (s *GameServer) clientToPlayer(clientID string) int {
	clientInfo, ok := s.clients[clientID]
	if !ok {
		return skat.PlayerNone
	}

	absoluteIndex := clientInfo.playerIndex
	// TODO: support dealer round by adjusting 3 to 4 if there is a dealer
	// round and then aliasing the dealer somehow. *shrug*
	relativeIndex := (s.currentPlayerOffset + absoluteIndex) % 3

	return relativeIndex
}

func (s *GameServer) playerToClient(relativeIndex int) string {
	// TODO: support dealer round by adjusting 3 to 4 if there is a dealer
	// round and then aliasing the dealer somehow. *shrug*
	absoluteIndex := (relativeIndex - s.currentPlayerOffset + 3) % 3

	clientID := s.playerReverseMap[absoluteIndex]
	return clientID
}

func (s *GameServer) processAction(clientID string, action replay.Action) error {
	// TODO: record action
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	playerIndex := s.clientToPlayer(clientID)
	if playerIndex == skat.PlayerNone {
		return ErrPlayerNotFound
	}

	err := action.Apply(s.currentGame, playerIndex)
	s.l.Debugw("applied action",
		"player", playerIndex,
		"action", action.Kind(),
	)
	if err == nil {
		s.pushState()
	}
	return err
}

func (s *GameServer) pushSingleState(ctx context.Context, clientID string) {
	clientInfo, ok := s.clients[clientID]
	if !ok {
		s.l.Warnw("player not mapped to client",
			"clientID", clientID,
		)
		return
	}
	playerIndex := clientInfo.playerIndex

	ep := clientInfo.ep
	if ep == nil {
		s.l.Warnw("client has no associated endpoint",
			"clientID", clientID,
			"playerIndex", playerIndex,
		)
		return
	}

	state := s.currentGame.BlindedForPlayer(playerIndex)
	msg := NewStateMessage(playerIndex, state)
	if err := ep.OneShot(ctx, msg); err != nil {
		s.l.Warnw("failed to push state to client",
			"clientID", clientID,
			"playerIndex", playerIndex,
		)
		return
	}

	s.l.Debugw("pushed state",
		"clientID", clientID,
		"playerIndex", playerIndex,
	)
}

func (s *GameServer) pushState() {
	// TODO: we probably want to .. I don’t know, somehow deadline this, but
	// not sure how to best deadline it.
	ctx := context.Background()
	for _, clientID := range s.playerReverseMap {
		s.pushSingleState(ctx, clientID)
	}
	s.l.Debugw("state pushed to players")
}

func (s *GameServer) handleActionMessage(clientID string, msg *ActionMessage) (Message, error) {
	action, err := msg.Payload()
	if err != nil {
		return NewErrorMessage(400, err.Error()), nil
	}

	if err := s.processAction(clientID, action); err != nil {
		// when in doubt, it’s the clients fault
		return NewErrorMessage(400, err.Error()), nil
	}

	return &AckMessage{}, nil
}

func (s *GameServer) handleMessage(clientID string, msgH MessageHandle, endpoint MessageEndpoint) {
	var reply Message
	var err error
	defer msgH.Close()

	msg := msgH.Message()
	type_ := msg.Type()
	switch type_ {
	case MsgAction:
		actionMessage, ok := msg.(*ActionMessage)
		if !ok {
			reply = NewErrorMessage(400, "malformed action")
		} else {
			reply, err = s.handleActionMessage(clientID, actionMessage)
			if err != nil {
				reply = NewErrorMessage(500, "failed to process action")
			}
		}
	case MsgPing:
		reply = NewPong()
	case MsgPong, MsgAck:
		s.l.Debugw("stray message?!",
			"clientID", clientID,
			"messageType", type_,
		)
		reply = NewErrorMessage(400, "stray message")
	default:
		s.l.Debugw("not implemented",
			"clientID", clientID,
			"messageType", type_,
		)
		reply = NewErrorMessage(500, "not implemented")
	}

	if reply != nil {
		if !msgH.ExpectsReply() {
			s.l.Debugw("discarding reply to oneshot message")
		} else {
			err := endpoint.Reply(msgH.Context(), reply)
			if err != nil {
				s.l.Warnw("failed to send reply",
					"clientID", clientID,
					"err", err,
				)
			}
		}
	} else if msgH.ExpectsReply() {
		s.l.Fatalw("no reply?!",
			"clientID", clientID,
			"messageType", type_,
			"message", msg,
		)
	}
}

func (s *GameServer) Run() {
	for {
		clientID, msgH, ep := s.getNextMessage()
		if msgH == nil {
			s.l.Infow("got nil message, shutting down")
			return
		}

		s.handleMessage(clientID, msgH, ep)
	}
}

func (s *GameServer) ConnectClient(ep MessageEndpoint, ctx context.Context) error {
	msg, err := RequestResponse(ep, ctx, NewPing())
	if err != nil {
		s.l.Debugw("initial ping/pong failed",
			"err", err,
		)
		return err
	}

	if msg.Type() != MsgPong {
		s.l.Debugw("invalid reply to initial ping",
			"type", msg.Type(),
		)
		ep.Close()
		return ErrProtocolViolation
	}

	s.l.Debugw("initial ping/pong exchanged")

	var loginMessage *LoginRequestMessage
	var loginCtx context.Context
	select {
	case req, ok := <-ep.RecvChannel():
		{
			if !ok {
				s.l.Debugw("prematurely closed connection during initial handshake")
				return ErrClosed
			}
			loginMessage, ok = req.Message().(*LoginRequestMessage)
			if !ok || loginMessage.Type() != MsgLoginReq {
				s.l.Debugw("initial message is not a login request",
					"type", msg.Type(),
				)
				ep.Close()
				return ErrProtocolViolation
			}
			loginCtx = req.Context()
		}
	case <-ctx.Done():
		{
			s.l.Debugw("initial handshake failed",
				"err", ctx.Err(),
			)
			return ctx.Err()
		}
	}
	s.l.Debugw("login request received")

	if s.serverPassword != "" {
		if loginMessage.ServerPassword != s.serverPassword {
			s.l.Debugw("invalid server password, returning 401")
			err = ep.Reply(
				loginCtx,
				&ErrorMessage{
					Code:    401,
					Message: "unauthorized",
				},
			)
			ep.Close()
			return err
		}
	}

	clientID := loginMessage.ClientID
	clientSecret := loginMessage.ClientSecret

	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	existing, ok := s.clients[clientID]
	var playerIndex int
	if ok {
		if existing.clientSecret != clientSecret {
			s.l.Debugw("secret does not match existing secret")
			err = ep.Reply(
				loginCtx,
				&ErrorMessage{
					Code:    401,
					Message: "unauthorized",
				},
			)
			ep.Close()
			return err
		}
		s.l.Debugw("previous client returned with new connection",
			"clientID", clientID,
		)
		// previous client returning
		if existing.ep != nil {
			existing.ep.Close()
		}
		existing.ep = ep
		playerIndex = existing.playerIndex
	} else {
		playerIndex = len(s.playerReverseMap)
		if len(s.playerReverseMap) >= 3 {
			s.l.Debugw("too many clients, rejecting new client",
				"clientID", clientID,
			)
			err = ep.Reply(
				loginCtx,
				&ErrorMessage{
					Code:    403,
					Message: "too many users",
				},
			)
			ep.Close()
			return err
		}
	}

	l := s.l.With("clientID", clientID)

	l.Debugw("client authenticated, replying OK")

	err = ep.Reply(
		loginCtx,
		&LoginOkMessage{},
	)
	if err != nil {
		l.Errorw("failed to send ok reply to client",
			"err", err,
		)
		ep.Close()
		return err
	}

	s.clients[clientID] = &gameClientConn{
		clientSecret: clientSecret,
		ep:           ep,
		playerIndex:  playerIndex,
	}
	if len(s.playerReverseMap) == playerIndex {
		s.playerReverseMap = append(s.playerReverseMap, clientID)
	}

	l.Infow("client connected successfully, refreshing worker")

	select {
	case s.wakeup <- struct{}{}:
		l.Debugw("worker refreshed successfully")
	default:
		l.Warnw("worker was not ready for refreshments")
	}

	s.pushSingleState(ctx, clientID)

	return nil
}
