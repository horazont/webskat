package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"time"

	"go.uber.org/zap"

	"github.com/horazont/webskat/internal/frontend/singleuser"
)

var (
	serverListenAddress = flag.String("server.listen-address", "127.0.0.1:5023", "")
	serverPassword      = flag.String("server.password", "foobar2342", "")
)

func generateSelfSigned() tls.Certificate {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return tlsCert
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	zap.ReplaceGlobals(logger)

	sl := zap.S()

	flag.Parse()

	ns, err := singleuser.Listen(
		*serverListenAddress,
		&tls.Config{
			InsecureSkipVerify: true,
			NextProtos: []string{
				singleuser.ProtocolName,
			},
			Certificates: []tls.Certificate{generateSelfSigned()},
		},
		sl.With(
			"component", "net_server",
		),
	)
	if err != nil {
		sl.Fatalw("failed to start server",
			err, "err",
		)
	}
	sl.Infow("listener set up",
		"listenAddress", *serverListenAddress,
	)

	gs, err := singleuser.NewGameServer(singleuser.GameServerConfig{
		ServerPassword: *serverPassword,
	}, sl.With("component", "game_server"))
	if err != nil {
		sl.Fatalw("failed to initialize game",
			"err", err,
		)
	}

	clients := ns.ClientsChannel()

	go gs.Run()

	go func() {
		for {
			client := <-clients
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := gs.ConnectClient(client, ctx)
			if err != nil {
				sl.Debugw("failed to fully accept client connection",
					"err", err,
				)
			} else {
				sl.Infow("new client connected")
			}
			cancel()
		}
	}()

	ctx := context.Background()
	ns.Serve(ctx)
}
