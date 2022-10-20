package wasmapi

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dapr/kit/logger"
	"github.com/pkg/errors"
	"github.com/wapc/wapc-go/engines/wazero"

	"github.com/dapr/dapr/pkg/apphealth"
	"github.com/dapr/dapr/pkg/channel"
	"github.com/dapr/dapr/pkg/config"
	invokev1 "github.com/dapr/dapr/pkg/messaging/v1"
	"github.com/dapr/dapr/pkg/wasmapi/abi"
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
	pool *Pool
	r    *abi.ComponentRegistry
}

// CreateWASMChannel creates an WASM AppChannel
// nolint:gosec
func CreateWASMChannel(port, maxConcurrency int, spec config.TracingSpec, sslEnabled bool, maxRequestBodySize int, readBufferSize int) (channel.AppChannel, error) {
	ctx := context.Background()
	code, err := os.ReadFile("./examples/uppercase/example.wasm")
	if err != nil {
		return nil, err
	}

	mod, err := wazero.Engine().New(ctx, code, hostCall)

	// At the moment, we use a global logger for any wasm output.
	mod.SetLogger(logfn)
	mod.SetWriter(logfn)

	c := &Channel{}
	c.pool, err = NewPool(ctx, mod, uint64(maxConcurrency))
	if err != nil {
		return nil, fmt.Errorf("error creating module pool from wasm at")
	}

	if maxConcurrency > 0 {
	}

	return c, nil
}

// GetBaseAddress returns the application base address.
func (h *Channel) GetBaseAddress() string {
	return ""
}

// GetAppConfig gets application config from user application
func (h *Channel) GetAppConfig() (*config.ApplicationConfig, error) {
	// todo
	var config config.ApplicationConfig

	return &config, nil
}

// InvokeMethod invokes user code via HTTP.
func (h *Channel) InvokeMethod(ctx context.Context, req *invokev1.InvokeMethodRequest) (*invokev1.InvokeMethodResponse, error) {
	var rsp *invokev1.InvokeMethodResponse
	var err error
	contentType, body := req.RawData()
	_ = contentType
	method := req.Message().GetMethod()
	if strings.HasPrefix(method, "/") {
		method = method[1:]
	}
	if method == "" {
		method = "default"
	}

	ins, err := h.pool.Get(2 * time.Second)
	if err != nil {
		return nil, err
	}
	defer h.pool.Return(ins)
	// todo use method
	res, err := ins.Invoke(ctx, method, body)
	if err != nil {
		return nil, err
	}

	rsp = invokev1.NewInvokeMethodResponse(int32(200), "", nil)
	rsp.WithRawData(res, "")
	return rsp, err
}

func hostCall(ctx context.Context, component, comtype, operation string, payload []byte) ([]byte, error) {
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
			name := string(payload)
			name = strings.Title(name)
			return []byte(name), nil
		}
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
