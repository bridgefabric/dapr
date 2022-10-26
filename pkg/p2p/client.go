package p2p

import (
	"fmt"
	"net/http"
	"os"

	"github.com/libp2p/go-libp2p/core/host"
)

const p2pProtocol = "libp2p"

func CallRemote(clientHost host.Host, id string) {
	tr := &http.Transport{}
	tr.RegisterProtocol(p2pProtocol, NewTransport(clientHost))
	client := &http.Client{Transport: tr}
	res, err := client.Get(fmt.Sprintf("%s://%s/hello", p2pProtocol, id))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	res.Write(os.Stdout)
}
