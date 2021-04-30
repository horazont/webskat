package singleuser

import (
	"context"
	"crypto/tls"
	"errors"

	"go.uber.org/zap"

	"github.com/lucas-clemente/quic-go"
)

var (
	ErrClosed         = errors.New("already closed")
	ErrPipelining     = errors.New("too many requests pipelined")
	ErrNoReplyContext = errors.New("context has no request information")
)

type NetServer struct {
	l             *zap.SugaredLogger
	listener      quic.Listener
	maxPipelining int
	clients       chan MessageEndpoint
}

func Listen(address string, tlsConfig *tls.Config, l *zap.SugaredLogger) (*NetServer, error) {
	listener, err := quic.ListenAddr(address, tlsConfig, &quic.Config{
		KeepAlive:          true,
		ConnectionIDLength: 8,
	})
	if err != nil {
		return nil, err
	}

	return &NetServer{
		l:             l,
		listener:      listener,
		maxPipelining: 4,
		clients:       make(chan MessageEndpoint, 4),
	}, nil
}

func (s *NetServer) Serve(ctx context.Context) error {
	for {
		session, err := s.listener.Accept(ctx)
		if err != nil {
			return err
		}
		s.l.Debugw("accepted connection",
			"remoteAddr", session.RemoteAddr(),
		)

		client, err := wrapSessionServer(session, s.maxPipelining, s.l.With(
			"remoteAddr", session.RemoteAddr(),
		))
		if err != nil {
			s.l.Warnw("failed to bootstrap session",
				"err", err,
			)
			session.CloseWithError(500, "failed to bootstrap session")
			continue
		}

		if len(s.clients) >= cap(s.clients) {
			s.l.Warnw("connection queue overflow")
			session.CloseWithError(503, "too many connections")
			continue
		}

		go client.loop()
		s.clients <- client
	}
}

func (s *NetServer) ClientsChannel() <-chan MessageEndpoint {
	return s.clients
}
