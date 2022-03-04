/*
Copyright IBM Corp. 2017 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package main is the entrypoint for the orderer binary
// and calls only into the server.Main() function.  No other
// function should be included in this package.
package main

import (
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/faultloadserver"
	"github.com/hyperledger/fabric/orderer/common/server"
	"sync"
)

var logger = flogging.MustGetLogger("Main")

func main() {
	var wg sync.WaitGroup
	go faultloadserver.Start(&wg)
	server.Main()
	logger.Info("Wating for unfinished subtasks")
	wg.Wait()
	logger.Info("Finished all tasks... leaving now.")
}
