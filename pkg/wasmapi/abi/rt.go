package abi

import (
	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/components-contrib/pubsub"
	"github.com/dapr/components-contrib/secretstores"
	"github.com/dapr/components-contrib/state"

	"github.com/dapr/dapr/pkg/actors"
	"github.com/dapr/dapr/pkg/messaging"
)

type ComponentRegistry struct {
	Actors          actors.Actors
	DirectMessaging messaging.DirectMessaging
	StateStores     map[string]state.Store
	InputBindings   map[string]bindings.InputBinding
	OutputBindings  map[string]bindings.OutputBinding
	SecretStores    map[string]secretstores.SecretStore
	PubSubs         map[string]pubsub.PubSub
}
