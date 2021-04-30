package singleuser

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"go.uber.org/zap"

	"github.com/horazont/webskat/internal/api"
)

var (
	ErrWrongContentType = errors.New("wrong content type")
	ErrTooBig           = errors.New("request too big")
)

func getContentTypeFromString(header string) string {
	mediatype, _, err := mime.ParseMediaType(header)
	if err != nil {
		return ""
	}
	return mediatype
}

func getContentType(r *http.Request) string {
	return getContentTypeFromString(r.Header.Get("Content-Type"))
}

func readJSON(r *http.Request, v interface{}) error {
	contentType := getContentType(r)
	if contentType != "application/json" {
		return ErrWrongContentType
	}

	if r.ContentLength > 65535 {
		return ErrTooBig
	}

	bodyReader := &io.LimitedReader{
		R: r.Body,
		N: 65535,
	}
	defer r.Body.Close()

	dec := json.NewDecoder(bodyReader)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := enc.Encode(v)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	buf.WriteTo(w)
}

func NewV1Handler(dataDirectory string, serverPassword string) (http.Handler, error) {
	tally, err := NewTally(dataDirectory)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		sl := zap.S()

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		req := &api.RegisterV1Request{}
		err := readJSON(r, req)

		if err != nil {
			sl.Debugw("failed to parse request json",
				"err", err,
				"endpoint", "/register",
			)
			w.WriteHeader(400)
			return
		}

		sl.Debugw("received registration request",
			"request", req,
			"endpoint", "/register",
		)

		if req.ServerPassword != serverPassword {
			sl.Debugw("not authorized: incorrect server password",
				"endpoint", "/register",
			)
			w.WriteHeader(401)
			return
		}

		clientID, err := tally.Register(req.ClientSecret, req.DisplayName)
		if err != nil {
			sl.Errorw("failed to add user",
				"err", err,
				"endpoint", "/register",
			)
			w.WriteHeader(500)
			return
		}

		resp := &api.RegisterV1Response{
			ClientID: clientID,
		}
		writeJSON(w, resp)
	})

	return mux, nil
}
