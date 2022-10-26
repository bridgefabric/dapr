package wasm

import (
	"net"
	"sync"
	"time"

	"github.com/dapr/kit/logger"
	"go.uber.org/atomic"
	"google.golang.org/grpc"

	daprCredentials "github.com/dapr/dapr/pkg/credentials"
	"github.com/dapr/dapr/pkg/placement/hashing"
	v1pb "github.com/dapr/dapr/pkg/proto/placement/v1"
)

var log = logger.NewLogger("dapr.runtime.actor.internal.placement")

const (
	lockOperation   = "lock"
	unlockOperation = "unlock"
	updateOperation = "update"

	placementReconnectInterval    = 500 * time.Millisecond
	statusReportHeartbeatInterval = 11 * time.Second

	grpcServiceConfig = `{"loadBalancingPolicy":"round_robin"}`
)

// ActorPlacement maintains membership of actor instances and consistent hash
// tables to discover the actor while interacting with Placement service.
//
//nolint:nosnakecase
type ActorPlacement struct {
	actorTypes []string
	appID      string
	// runtimeHostname is the address and port of the runtime
	runtimeHostName string

	// serverAddr is the list of placement addresses.
	serverAddr []string
	// serverIndex is the current index of placement servers in serverAddr.
	serverIndex atomic.Int32

	// clientCert is the workload certificate to connect placement.
	clientCert *daprCredentials.CertChain

	// clientLock is the lock for client conn and stream.
	clientLock *sync.RWMutex
	// clientConn is the gRPC client connection.
	clientConn *grpc.ClientConn
	// clientStream is the client side stream.
	clientStream v1pb.Placement_ReportDaprStatusClient
	// streamConnAlive is the status of stream connection alive.
	streamConnAlive bool
	// streamConnectedCond is the condition variable for goroutines waiting for or announcing
	// that the stream between runtime and placement is connected.
	streamConnectedCond *sync.Cond

	// placementTables is the consistent hashing table map to
	// look up Dapr runtime host address to locate actor.
	placementTables *hashing.ConsistentHashTables
	// placementTableLock is the lock for placementTables.
	placementTableLock *sync.RWMutex

	// unblockSignal is the channel to unblock table locking.
	unblockSignal chan struct{}
	// tableIsBlocked is the status of table lock.
	tableIsBlocked *atomic.Bool
	// operationUpdateLock is the lock for three stage commit.
	operationUpdateLock *sync.Mutex

	// appHealthFn is the user app health check callback.
	appHealthFn func() bool
	// afterTableUpdateFn is function for post processing done after table updates,
	// such as draining actors and resetting reminders.
	afterTableUpdateFn func()

	// shutdown is the flag when runtime is being shutdown.
	shutdown atomic.Bool
	// shutdownConnLoop is the wait group to wait until all connection loop are done
	shutdownConnLoop sync.WaitGroup
}

func addDNSResolverPrefix(addr []string) []string {
	resolvers := make([]string, 0, len(addr))
	for _, a := range addr {
		prefix := ""
		host, _, err := net.SplitHostPort(a)
		if err == nil && net.ParseIP(host) == nil {
			prefix = "dns:///"
		}
		resolvers = append(resolvers, prefix+a)
	}
	return resolvers
}

// NewActorPlacement initializes ActorPlacement for the actor service.
func NewActorPlacement(
	serverAddr []string, clientCert *daprCredentials.CertChain,
	appID, runtimeHostName string, actorTypes []string,
	appHealthFn func() bool,
	afterTableUpdateFn func(),
) *ActorPlacement {
	return &ActorPlacement{
		actorTypes:      actorTypes,
		appID:           appID,
		runtimeHostName: runtimeHostName,
		serverAddr:      addDNSResolverPrefix(serverAddr),

		clientCert: clientCert,

		clientLock:          &sync.RWMutex{},
		streamConnAlive:     false,
		streamConnectedCond: sync.NewCond(&sync.Mutex{}),

		placementTableLock: &sync.RWMutex{},
		placementTables:    &hashing.ConsistentHashTables{Entries: make(map[string]*hashing.Consistent)},

		operationUpdateLock: &sync.Mutex{},
		tableIsBlocked:      atomic.NewBool(false),
		appHealthFn:         appHealthFn,
		afterTableUpdateFn:  afterTableUpdateFn,
	}
}

// Start connects placement service to register to membership and send heartbeat
// to report the current member status periodically.
func (p *ActorPlacement) Start() {
	// todo connect to contract
}

// Stop shuts down server stream gracefully.
func (p *ActorPlacement) Stop() {
	// todo
}

// WaitUntilPlacementTableIsReady waits until placement table is until table lock is unlocked.
func (p *ActorPlacement) WaitUntilPlacementTableIsReady() {
}

// LookupActor resolves to actor service instance address using consistent hashing table.
func (p *ActorPlacement) LookupActor(actorType, actorID string) (string, string) {
	return p.runtimeHostName, p.appID
}
