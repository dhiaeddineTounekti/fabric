package faultloadserver

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/common/flogging"
	flo "github.com/hyperledger/fabric/faultloadobject"
	"net/http"
	"os/exec"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

const (
	indifferent string = "indifferent"
)

type faultLoadServerVars struct {
	mux     *http.ServeMux
	address string
	port    string
}

type faultLoadBody struct {
	Condition     string `json:"condition"`
	FaultDuration int    `json:"faultDuration"`
}

type crashFault struct {
	FaultLoad faultLoadBody `json:"faultLoadBody"`
}

type communicationFault struct {
	FaultLoad          faultLoadBody `json:"faultLoadBody"`
	CommunicationDelay int           `json:"communicationDelay"`
	PackageDropRate    int           `json:"packageDropRate"`
}

var (
	// structure containing the server configuration
	faultLoadServerVarsInstance *faultLoadServerVars
	logger                      = flogging.MustGetLogger("faultload-server")
	globalWg                    *sync.WaitGroup
)

func init() {
	faultLoadServerVarsInstance = &faultLoadServerVars{
		mux:     createHandlers(),
		address: "",
		port:    "7052",
	}
}

func createHandlers() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/fault_load/crash", crashFaultHandler)
	mux.HandleFunc("/fault_load/communication", communicationFaultHandler)
	return mux
}

func Start(wg *sync.WaitGroup) {
	globalWg = wg
	listeningAddress := fmt.Sprintf("%s:%s", faultLoadServerVarsInstance.address, faultLoadServerVarsInstance.port)
	logger.Infof("Fault load server listening on %s", listeningAddress)

	if err := http.ListenAndServe(listeningAddress, faultLoadServerVarsInstance.mux); err != nil {
		logger.Errorw(err.Error())
	}
}

func crashFaultHandler(w http.ResponseWriter, r *http.Request) {
	globalWg.Add(1)
	defer globalWg.Done()
	var faultLoadConfig crashFault
	err := json.NewDecoder(r.Body).Decode(&faultLoadConfig)
	if err != nil {
		httpBadRequest(w, err)
		return
	}
	if shouldApplyFault(faultLoadConfig.FaultLoad) {
		logger.Infof("Crashing server after receiving request from: %s", r.RemoteAddr)

		globalWg.Add(1)
		go func() {
			defer globalWg.Done()

			commands := fmt.Sprintf("sleep %d && orderer", faultLoadConfig.FaultLoad.FaultDuration)
			cmd := exec.Command("/bin/sh", "-c", commands)
			err := cmd.Start()
			if err != nil {
				logger.Fatal("error crashing the server", err.Error())
			}

			flo.GetServer().Stop()

		}()
		if _, err := w.Write([]byte("Done executing the crash")); err != nil {
			logger.Error(err.Error())
		}
		return
	}

	logger.Infof("Ignoring Request intended for %s and I am %s", faultLoadConfig.FaultLoad.Condition, flo.GetLeadershipStatus())

}

func communicationFaultHandler(w http.ResponseWriter, r *http.Request) {
	globalWg.Add(1)
	defer globalWg.Done()
	var communicationFaultBody communicationFault

	err := json.NewDecoder(r.Body).Decode(&communicationFaultBody)
	if err != nil {
		httpBadRequest(w, err)
		return
	}

	if shouldApplyFault(communicationFaultBody.FaultLoad) {
		logger.Infof("Crashing server after receiving request from: %s", r.RemoteAddr)

		go func() {
			commands := fmt.Sprintf("tc qdisc add dev eth0 root netem loss %d%% delay %dms", communicationFaultBody.PackageDropRate, communicationFaultBody.CommunicationDelay)
			cmd := exec.Command("/bin/sh", "-c", commands)
			err := cmd.Run()
			if err != nil {
				logger.Fatal("Error executing the commands:", commands, err.Error())
			}

			time.Sleep(time.Second * time.Duration(communicationFaultBody.FaultLoad.FaultDuration))

			commands = fmt.Sprintf("tc qdisc del dev eth0 root")
			cmd = exec.Command("/bin/sh", "-c", commands)
			err = cmd.Run()
			if err != nil {
				logger.Fatal("Error crashing the server", err.Error())
			}
		}()
	}

	logger.Infof("Ignoring Request intended for %s and I am %s", communicationFaultBody.FaultLoad.Condition, flo.GetLeadershipStatus())

}

func httpBadRequest(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
	logger.Error(err.Error())
}

func GetUnexportedField(name string, object interface{}) interface{} {
	field := GetFieldByName(name, object)
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func SetUnexportedField(name string, object interface{}, value interface{}) {
	field := GetFieldByName(name, object)
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}

func GetFieldByName(name string, object interface{}) reflect.Value {
	return reflect.ValueOf(object).Elem().FieldByName(name)
}

func shouldApplyFault(faultLoadBody faultLoadBody) bool {
	return faultLoadBody.Condition == flo.GetLeadershipStatus() || faultLoadBody.Condition == indifferent
}
