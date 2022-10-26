package wasmapi

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/kit/logger"
	"github.com/pkg/errors"
	"github.com/wapc/wapc-go/engines/wazero"

	"github.com/dapr/dapr/pkg/actors"
	"github.com/dapr/dapr/pkg/apphealth"
	"github.com/dapr/dapr/pkg/channel"
	"github.com/dapr/dapr/pkg/config"
	invokev1 "github.com/dapr/dapr/pkg/messaging/v1"
	"github.com/dapr/dapr/pkg/wasmapi/abi"
	"github.com/dapr/dapr/pkg/wasmapi/eth"
	"github.com/dapr/dapr/pkg/wasmapi/w3s"
)

const (
	// HTTPStatusCode is an dapr http channel status code.
	HTTPStatusCode    = "http.status_code"
	httpScheme        = "http"
	httpsScheme       = "https"
	appConfigEndpoint = "dapr/config"

	poolSize = 20 // todo make it configurable
)

var log = logger.NewLogger("dapr.runtime.wasm")

// Channel is an WASM implementation of an AppChannel.
type Channel struct {
	lock                  sync.RWMutex
	pools                 map[string]*Pool
	r                     *abi.ComponentRegistry
	maxConcurrency        int
	actor                 actors.Actors
	nodeID                string
	sendToOutputBindingFn func(name string, req *bindings.InvokeRequest) (*bindings.InvokeResponse, error)
}

// CreateWASMChannel creates an WASM AppChannel
// nolint:gosec
func CreateWASMChannel(port, maxConcurrency int, spec config.TracingSpec, sslEnabled bool, maxRequestBodySize int, readBufferSize int, id string, sendToOutputBindingFn func(name string, req *bindings.InvokeRequest) (*bindings.InvokeResponse, error)) (channel.AppChannel, error) {
	//ctx := context.Background()
	//code, err := os.ReadFile("./examples/uppercase/example.wasm")
	//if err != nil {
	//	return nil, err
	//}
	//
	//mod, err := wazero.Engine().New(ctx, code, hostCall)
	//
	//// At the moment, we use a global logger for any wasm output.
	//mod.SetLogger(logfn)
	//mod.SetWriter(logfn)

	c := &Channel{maxConcurrency: 10, nodeID: id, sendToOutputBindingFn: sendToOutputBindingFn}
	c.pools = make(map[string]*Pool)
	//c.pool, err = NewPool(ctx, mod, uint64(maxConcurrency))
	//if err != nil {
	//	return nil, fmt.Errorf("error creating module pool from wasm at")
	//}

	if maxConcurrency > 0 {
		c.maxConcurrency = maxConcurrency
	}
	err := eth.InitETH(eth.DefaultConfig)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// GetBaseAddress returns the application base address.
func (h *Channel) GetBaseAddress() string {
	return ""
}

func (h *Channel) SetActor(actor actors.Actors) {
	h.actor = actor
}

// GetAppConfig gets application config from user application
func (h *Channel) GetAppConfig() (*config.ApplicationConfig, error) {
	// todo
	var config config.ApplicationConfig

	return &config, nil
}

func (h *Channel) isActorLocal(hosts []string) bool {
	for _, host := range hosts {
		if host == h.nodeID {
			return true
		}
	}
	return false
}

func (h *Channel) callLocalActor(ctx context.Context, req *invokev1.InvokeMethodRequest) (*invokev1.InvokeMethodResponse, error) {
	var rsp *invokev1.InvokeMethodResponse
	var err error
	actorType := req.Actor().ActorType
	_ = req.Actor().ActorId
	contentType, body := req.RawData()
	_ = contentType
	methods := strings.Split(req.Message().GetMethod(), "/")
	method := methods[len(methods)-1]
	if method == "" {
		method = "default"
	}

	pool, err := h.getActorPool(actorType)
	if err != nil {
		return nil, err
	}
	ins, err := pool.Get(2 * time.Second)
	if err != nil {
		return nil, err
	}
	defer pool.Return(ins)
	// todo use method
	res, err := ins.Invoke(ctx, method, body)
	if err != nil {
		return nil, err
	}

	rsp = invokev1.NewInvokeMethodResponse(int32(200), "", nil)
	rsp.WithRawData(res, "")
	return rsp, err
}

func (h *Channel) callRemoteActor(ctx context.Context, hostIds []string, req *invokev1.InvokeMethodRequest) (*invokev1.InvokeMethodResponse, error) {
	panic("implement me")
	//tr := &http.Transport{}
	//tr.RegisterProtocol("libp2p", p2p.NewTransport(clientHost))
	//client := &http.Client{Transport: tr}
	//res, err := client.Get("libp2p://Qmaoi4isbcTbFfohQyn28EiYM5CDWQx9QRCjDh3CTeiY7P/hello")
}

// InvokeMethod invokes user code via HTTP.
func (h *Channel) InvokeMethod(ctx context.Context, req *invokev1.InvokeMethodRequest) (*invokev1.InvokeMethodResponse, error) {
	var resp *invokev1.InvokeMethodResponse
	var err error
	actorType := req.Actor().ActorType
	_ = req.Actor().ActorId

	hosts, err := eth.CS.GetActorHolders(ctx, actorType)
	if err != nil {
		return nil, err
	}
	if h.isActorLocal(hosts) {
		resp, err = h.callLocalActor(ctx, req)
	} else {
		resp, err = h.callRemoteActor(ctx, hosts, req)
	}
	return resp, err
}

func (h *Channel) hostCall(ctx context.Context, component, comtype, operation string, payload []byte) ([]byte, error) {
	// Route the payload to any custom functionality accordingly.
	// You can even route to other waPC modules!!!
	switch comtype {
	case "state":
		switch operation {
		case "save":
			name := string(payload)
			log.Info(name)
			return []byte("OK"), nil
		}
	case "binding":
		switch operation {
		case "add":
			res, err := h.sendToOutputBindingFn(component, &bindings.InvokeRequest{Data: payload, Operation: bindings.OperationKind(operation)})
			if err != nil {
				return nil, err
			}
			return res.Data, nil
		}
	case "actor":
		// todo use actor runtime
		res, err := h.InvokeMethod(context.TODO(),
			invokev1.NewInvokeMethodRequest(operation).WithRawData(payload, "").WithActor(component, "default")) // todo fix id
		if err != nil {
			return nil, err
		}
		_, data := res.RawData()
		return data, nil
	case "pubsub":
		fallthrough
	default:
		return nil, errors.New("component type not found")
	}
	return nil, errors.New("operation not found")
}

// log implements wapc.Logger.
func logfn(msg string) {
	log.Info(msg)
}

// HealthProbe performs a health probe.
func (h *Channel) HealthProbe(ctx context.Context) (bool, error) {
	return true, nil
}

// SetAppHealth sets the apphealth.AppHealth object.
func (h *Channel) SetAppHealth(ah *apphealth.AppHealth) {
}

// SetAppHealth sets the apphealth.AppHealth object.
func (h *Channel) getActorPool(cid string) (*Pool, error) {
	if h.pools[cid] != nil {
		return h.pools[cid], nil
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.pools[cid] != nil {
		return h.pools[cid], nil
	}
	ctx := context.Background()
	code, err := w3s.Get(context.Background(), cid)
	if err != nil {
		return nil, err
	}
	mod, err := wazero.Engine().New(ctx, code, h.hostCall)
	if err != nil {
		return nil, err
	}
	// At the moment, we use a global logger for any wasm output.
	mod.SetLogger(logfn)
	mod.SetWriter(logfn)

	pool, err := NewPool(ctx, mod, uint64(h.maxConcurrency))
	if err != nil {
		return nil, fmt.Errorf("error creating module pool from wasm at")
	}
	h.pools[cid] = pool
	return pool, nil
}
