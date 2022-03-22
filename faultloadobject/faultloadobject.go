package faultloadobject

import (
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/internal/pkg/comm"
	"sync"
)

const (
	leader    string = "leader"
	notLeader        = "notLeader"
)

type faultLoadGlobalVars struct {
	server           *comm.GRPCServer
	mu               sync.Mutex
	leaderShipStatus string
}

var faultLoadGlobalVarInstance *faultLoadGlobalVars
var logger = flogging.MustGetLogger("faultload-object")

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
		logger.Info("Current node is set to Leader")
	} else {
		faultLoadGlobalVarInstance.leaderShipStatus = notLeader
		logger.Info("Current node is set to Follower")
	}
}

func init() {
	faultLoadGlobalVarInstance = &faultLoadGlobalVars{
		server:           &comm.GRPCServer{},
		mu:               sync.Mutex{},
		leaderShipStatus: notLeader,
	}
}
