package p2p

import (
	"net/http"

	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/host"
)

func GetListener(host host.Host) {
	listener, _ := gostream.Listen(host, DefaultP2PProtocol)
	go func() {
		http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hi!"))
		})
		server := &http.Server{}
		server.Serve(listener)
		listener.Close()
	}()
}
