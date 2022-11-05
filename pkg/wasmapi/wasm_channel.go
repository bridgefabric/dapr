package wasmapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/kit/logger"
	"github.com/pkg/errors"
	"github.com/wapc/wapc-go/engines/wazero"
	"google.golang.org/grpc/metadata"

	"github.com/dapr/dapr/pkg/actors"
	"github.com/dapr/dapr/pkg/apphealth"
	"github.com/dapr/dapr/pkg/channel"
	"github.com/dapr/dapr/pkg/config"
	invokev1 "github.com/dapr/dapr/pkg/messaging/v1"
	"github.com/dapr/dapr/pkg/p2p"
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
	log.Infof("actor %s is called for method %s", actorType, method)

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
	eth.CS.ReduceActor <- actorType
	rsp = invokev1.NewInvokeMethodResponse(int32(200), "", nil)
	rsp.WithRawData(res, "")
	return rsp, err
}

func (h *Channel) callRemoteActor(ctx context.Context, hostIds []string, req *invokev1.InvokeMethodRequest) (*invokev1.InvokeMethodResponse, error) {
	if len(hostIds) == 0 {
		return nil, errors.New("no host id")
	}
	return Call(req, hostIds[0])
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

func Call(req *invokev1.InvokeMethodRequest, id string) (*invokev1.InvokeMethodResponse, error) {
	client, err := p2p.AcquireP2PClient()
	if err != nil {
		return nil, err
	}
	// Check if HTTP Extension is given. Otherwise, it will return error.
	//httpExt := req.Message().GetHttpExtension()
	//if httpExt == nil {
	//	return nil, status.Error(codes.InvalidArgument, "missing HTTP extension field")
	//}
	//if httpExt.GetVerb() == commonv1pb.HTTPExtension_NONE { //nolint:nosnakecase
	//	return nil, status.Error(codes.InvalidArgument, "invalid HTTP verb")
	//}

	var rsp *invokev1.InvokeMethodResponse
	channelReq, err := constructRequest(context.Background(), req, id)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(channelReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	contentType := (string)(res.Header.Get("Content-Type"))
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// Convert status code
	rsp = invokev1.NewInvokeMethodResponse(int32(res.StatusCode), "", nil)
	rsp.WithHeaders(metadata.MD(res.Header)).WithRawData(body, contentType)

	return rsp, err
}

func constructRequest(ctx context.Context, req *invokev1.InvokeMethodRequest, id string) (*http.Request, error) {
	var err error
	//channelReq := http.Request{}
	httpMethod := http.MethodPost
	if req.Message().GetHttpExtension() != nil && len(req.Message().GetHttpExtension().GetVerb().String()) > 0 {
		httpMethod = req.Message().GetHttpExtension().GetVerb().String()
	}

	// Construct app channel URI: VERB http://localhost:3000/method?query1=value1
	var uri string
	actorType := req.Actor().ActorType
	_ = req.Actor().ActorId
	contentType, body := req.RawData()
	method := req.Message().GetMethod()
	//_ = m
	//methods := strings.Split(req.Message().GetMethod(), "/")
	//method := methods[len(methods)-1]
	if method == "" {
		method = "default"
	}
	if strings.HasPrefix(method, "/") {
		uri = fmt.Sprintf("%s://%s/%s%s", p2p.P2PProtocol, id, actorType, method)
	} else {
		uri = fmt.Sprintf("%s://%s/%s/%s", p2p.P2PProtocol, id, actorType, method)
	}

	channelReq, err := http.NewRequest(httpMethod, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	channelReq.Header.Set("Content-Type", contentType)

	channelReq.Header = make(http.Header)

	// Recover headers
	invokev1.InternalMetadataToHTTPHeader(ctx, req.Metadata(), channelReq.Header.Set)

	// HTTP client needs to inject traceparent header for proper tracing stack.
	//span := diagUtils.SpanFromContext(ctx)
	//tp := diag.SpanContextToW3CString(span.SpanContext())
	//ts := diag.TraceStateToW3CString(span.SpanContext())
	//channelReq.Header.Set("traceparent", tp)
	//if ts != "" {
	//	channelReq.Header.Set("tracestate", ts)
	//}

	return channelReq, nil
}

// DeconstructRequest convert http to invoke format
func DeconstructRequest(req *http.Request) (*invokev1.InvokeMethodRequest, error) {
	// /actorType/method
	reqMethod := req.URL.Path
	parts := strings.Split(reqMethod, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("error method format: %s", reqMethod)
	}
	actorType, method := parts[1], strings.Join(parts[2:], "/")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	return invokev1.NewInvokeMethodRequest(method).WithRawData(body, "").WithActor(actorType, "default"), nil
}
