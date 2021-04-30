package singleuser

import (
	"context"
	"crypto/tls"

	"go.uber.org/zap"

	"github.com/lucas-clemente/quic-go"
)

func DialAddr(l *zap.SugaredLogger, ctx context.Context, address string, tlsConfig *tls.Config) (MessageEndpoint, error) {
	maxPipelining := 4

	session, err := quic.DialAddr(
		address,
		tlsConfig,
		&quic.Config{
			KeepAlive:          true,
			ConnectionIDLength: 8,
		},
	)
	if err != nil {
		l.Errorw(
			"failed to connect to server",
			"remoteAddr", address,
			"err", err,
		)
		return nil, err
	}

	l.Debugw(
		"connected",
		"remoteAddr", session.RemoteAddr(),
	)

	result, err := wrapSessionClient(
		ctx,
		session,
		maxPipelining,
		l,
	)
	if err != nil {
		return nil, err
	}

	go result.loop()

	return result, nil
}
