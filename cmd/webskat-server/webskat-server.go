package main

import (
	"flag"
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/horazont/webskat/internal/frontend/singleuser"
)

var (
	webListenAddress = flag.String("web.listen-address", "127.0.0.1:5023", "")
	stateDirectory   = flag.String("data.state-directory", "./data", "")
	serverPassword   = flag.String("web.server-password", "foobar2342", "")
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	zap.ReplaceGlobals(logger)

	sl := zap.S()

	handler, err := singleuser.NewV1Handler(
		*stateDirectory,
		*serverPassword,
	)
	if err != nil {
		sl.Panicw("failed to create initial handler",
			"err", err,
		)
	}

	sl.Infow("starting listener",
		"address", *webListenAddress,
	)
	err = http.ListenAndServe(*webListenAddress, handler)
	sl.Fatalw("listener stopped",
		"err", err,
	)
}
