package wasmapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	invokev1 "github.com/dapr/dapr/pkg/messaging/v1"

	"github.com/dapr/dapr/pkg/config"
)

func TestChannel(t *testing.T) {
	c, e := CreateWASMChannel(0, 10, config.TracingSpec{}, false, 0, 0)
	if e != nil {
		t.Error(e)
	}
	resp, err := c.InvokeMethod(context.Background(),
		invokev1.NewInvokeMethodRequest("upperCase").WithRawData([]byte("sss"), ""))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "SSS", string(resp.Message().GetData().GetValue()))
}

func TestSave(t *testing.T) {
	c, e := CreateWASMChannel(0, 10, config.TracingSpec{}, false, 0, 0)
	if e != nil {
		t.Error(e)
	}
	resp, err := c.InvokeMethod(context.Background(),
		invokev1.NewInvokeMethodRequest("upperCase").WithRawData([]byte("sss"), ""))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "SSS", string(resp.Message().GetData().GetValue()))
}
