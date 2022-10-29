package p2p

import (
	"fmt"
	"net/http"
	"os"

	"github.com/libp2p/go-libp2p/core/host"
)

const P2PProtocol = "libp2p"

func CallRemote(clientHost host.Host, id string) {
	tr := &http.Transport{}
	tr.RegisterProtocol(P2PProtocol, NewTransport(clientHost))
	client := &http.Client{Transport: tr}
	res, err := client.Get(fmt.Sprintf("%s://%s/hello", P2PProtocol, id))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	res.Write(os.Stdout)
}

func AcquireP2PClient() (*http.Client, error) {
	if P2PHost == nil {
		return nil, ErrP2PNotInitialized
	}
	tr := &http.Transport{}
	tr.RegisterProtocol(P2PProtocol, NewTransport(P2PHost))
	client := &http.Client{Transport: tr}
	return client, nil
}
