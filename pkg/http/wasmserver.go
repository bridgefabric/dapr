package http

import (
	"context"
	"net"
	"net/http"

	"github.com/dapr/dapr/pkg/channel"
	"github.com/dapr/dapr/pkg/wasmapi"
)

// todo ioc

type wasmServer struct {
	c channel.AppChannel
}

func (m *wasmServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	invokeReq, err := wasmapi.DeconstructRequest(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	resp, err := m.c.InvokeMethod(context.Background(), invokeReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.WriteHeader(http.StatusOK)
	_, body := resp.RawData()
	w.Write(body)
}

func StartWasmInternalServer(c channel.AppChannel, l net.Listener) {
	go func() {
		err := http.Serve(l, &wasmServer{c})
		l.Close() // todo this should call outside
		if err != nil {
			log.Info(err)
		}
	}()

}
