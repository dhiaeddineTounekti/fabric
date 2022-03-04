package faultloadobject

import (
	ab "github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric/internal/pkg/comm"
	"github.com/hyperledger/fabric/orderer/common/localconfig"
	"sync"
	"time"
)

const (
	leader    string = "leader"
	notLeader        = "notLeader"
)

type faultLoadGlobalVars struct {
	server           *comm.GRPCServer
	mu               sync.Mutex
	commDelay        *communicationDelayFault
	leaderShipStatus string
	serverConfig     comm.ServerConfig
	conf             *localconfig.TopLevel
	abcServer        ab.AtomicBroadcastServer
}

type communicationDelayFault struct {
	enabled bool
	// The duration of the delay inducing in the communication layer.
	delay time.Duration
}

var faultLoadGlobalVarInstance *faultLoadGlobalVars

func init() {
	faultLoadGlobalVarInstance = &faultLoadGlobalVars{
		mu:               sync.Mutex{},
		commDelay:        &communicationDelayFault{false, time.Duration(0)},
		leaderShipStatus: notLeader,
	}
}

func SetServer(server *comm.GRPCServer) {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	faultLoadGlobalVarInstance.server = server
}

func GetServer() *comm.GRPCServer {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	return faultLoadGlobalVarInstance.server
}

func SetServerConfig(serverConfig comm.ServerConfig) {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	faultLoadGlobalVarInstance.serverConfig = serverConfig
}

func GetServerConfig() comm.ServerConfig {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	return faultLoadGlobalVarInstance.serverConfig
}

func SetConfig(conf *localconfig.TopLevel) {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	faultLoadGlobalVarInstance.conf = conf
}

func GetConf() *localconfig.TopLevel {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	return faultLoadGlobalVarInstance.conf
}

func SetAbcServer(abcServer ab.AtomicBroadcastServer) {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	faultLoadGlobalVarInstance.abcServer = abcServer
}

func GetAbcServer() ab.AtomicBroadcastServer {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	return faultLoadGlobalVarInstance.abcServer
}

func SetCommunicationDelay(delay time.Duration) {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	faultLoadGlobalVarInstance.commDelay.enabled = true
	faultLoadGlobalVarInstance.commDelay.delay = delay
}

func GetCommunicationDelay() (bool, time.Duration) {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	return faultLoadGlobalVarInstance.commDelay.enabled, faultLoadGlobalVarInstance.commDelay.delay
}

func GetLeadershipStatus() string {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	return faultLoadGlobalVarInstance.leaderShipStatus
}

func SetLeadershipStatus(isLeader bool) {
	faultLoadGlobalVarInstance.mu.Lock()
	defer faultLoadGlobalVarInstance.mu.Unlock()
	if isLeader {
		faultLoadGlobalVarInstance.leaderShipStatus = leader
	} else {
		faultLoadGlobalVarInstance.leaderShipStatus = notLeader
	}
}
