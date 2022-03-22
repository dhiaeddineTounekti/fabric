package faultloadserver

import (
	"context"
	"encoding/json"
	"fmt"
	ab "github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/common/grpclogging"
	"github.com/hyperledger/fabric/common/grpcmetrics"
	"github.com/hyperledger/fabric/common/metrics"
	"github.com/hyperledger/fabric/common/metrics/disabled"
	"github.com/hyperledger/fabric/common/metrics/prometheus"
	"github.com/hyperledger/fabric/faultloadobject"
	"github.com/hyperledger/fabric/internal/pkg/comm"
	"github.com/hyperledger/fabric/orderer/common/localconfig"
	"github.com/hyperledger/fabric/orderer/common/multichannel"
	"github.com/hyperledger/fabric/orderer/common/server"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "General server start",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &localconfig.TopLevel{
				General: localconfig.General{
					ConnectionTimeout: 7 * time.Second,
					TLS: localconfig.TLS{
						Enabled:            false,
						ClientAuthRequired: true,
						Certificate:        "main.go",
						PrivateKey:         "main.go",
						RootCAs:            []string{"main.go"},
						ClientRootCAs:      []string{"main.go"},
					},
				},
			}
			sc := initializeServerConfig(conf, nil)
			grpcServer := InitializeGrpcServer(conf, sc)
			server1 := server.NewServer(
				&multichannel.Registrar{},
				&prometheus.Provider{},
				&conf.Debug,
				conf.General.Authentication.TimeWindow,
				false,
				conf.General.Authentication.NoExpirationChecks,
			)
			ab.RegisterAtomicBroadcastServer(grpcServer.Server(), server1)

			faultloadobject.SetServer(grpcServer)

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := grpcServer.Start()
				if err != nil {
					fmt.Println(err)
				}
			}()
			go func() {
				Start(&wg)
			}()
			wg.Wait()
		})
	}
}

func Test_crashFaultHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "crash fault handler",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crashFaultHandler(tt.args.w, tt.args.r)
		})
	}
}

func Test_createHandlers(t *testing.T) {
	tests := []struct {
		name string
		want *http.ServeMux
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createHandlers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createHandlers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_httpBadRequest(t *testing.T) {
	type args struct {
		w   http.ResponseWriter
		err error
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpBadRequest(tt.args.w, tt.args.err)
		})
	}
}

func Test_parseRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    faultLoadBody
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRequest(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func initializeServerConfig(conf *localconfig.TopLevel, metricsProvider metrics.Provider) comm.ServerConfig {
	// secure server config
	secureOpts := comm.SecureOptions{
		UseTLS:            conf.General.TLS.Enabled,
		RequireClientCert: conf.General.TLS.ClientAuthRequired,
		TimeShift:         conf.General.TLS.TLSHandshakeTimeShift,
	}
	// check to see if TLS is enabled
	if secureOpts.UseTLS {
		msg := "TLS"
		// load crypto material from files
		serverCertificate, err := ioutil.ReadFile(conf.General.TLS.Certificate)
		if err != nil {
			logger.Fatalf("Failed to load server Certificate file '%s' (%s)",
				conf.General.TLS.Certificate, err)
		}
		serverKey, err := ioutil.ReadFile(conf.General.TLS.PrivateKey)
		if err != nil {
			logger.Fatalf("Failed to load PrivateKey file '%s' (%s)",
				conf.General.TLS.PrivateKey, err)
		}
		var serverRootCAs, clientRootCAs [][]byte
		for _, serverRoot := range conf.General.TLS.RootCAs {
			root, err := ioutil.ReadFile(serverRoot)
			if err != nil {
				logger.Fatalf("Failed to load ServerRootCAs file '%s' (%s)",
					err, serverRoot)
			}
			serverRootCAs = append(serverRootCAs, root)
		}
		if secureOpts.RequireClientCert {
			for _, clientRoot := range conf.General.TLS.ClientRootCAs {
				root, err := ioutil.ReadFile(clientRoot)
				if err != nil {
					logger.Fatalf("Failed to load ClientRootCAs file '%s' (%s)",
						err, clientRoot)
				}
				clientRootCAs = append(clientRootCAs, root)
			}
			msg = "mutual TLS"
		}
		secureOpts.Key = serverKey
		secureOpts.Certificate = serverCertificate
		secureOpts.ServerRootCAs = serverRootCAs
		secureOpts.ClientRootCAs = clientRootCAs
		logger.Infof("Starting orderer with %s enabled", msg)
	}
	kaOpts := comm.DefaultKeepaliveOptions
	// keepalive settings
	// ServerMinInterval must be greater than 0
	if conf.General.Keepalive.ServerMinInterval > time.Duration(0) {
		kaOpts.ServerMinInterval = conf.General.Keepalive.ServerMinInterval
	}
	kaOpts.ServerInterval = conf.General.Keepalive.ServerInterval
	kaOpts.ServerTimeout = conf.General.Keepalive.ServerTimeout

	commLogger := flogging.MustGetLogger("core.comm").With("server", "Orderer")

	if metricsProvider == nil {
		metricsProvider = &disabled.Provider{}
	}

	return comm.ServerConfig{
		SecOpts:            secureOpts,
		KaOpts:             kaOpts,
		Logger:             commLogger,
		ServerStatsHandler: comm.NewServerStatsHandler(metricsProvider),
		ConnectionTimeout:  conf.General.ConnectionTimeout,
		StreamInterceptors: []grpc.StreamServerInterceptor{
			grpcmetrics.StreamServerInterceptor(grpcmetrics.NewStreamMetrics(metricsProvider)),
			grpclogging.StreamServerInterceptor(flogging.MustGetLogger("comm.grpc.server").Zap()),
		},
		UnaryInterceptors: []grpc.UnaryServerInterceptor{
			grpcmetrics.UnaryServerInterceptor(grpcmetrics.NewUnaryMetrics(metricsProvider)),
			grpclogging.UnaryServerInterceptor(
				flogging.MustGetLogger("comm.grpc.server").Zap(),
				grpclogging.WithLeveler(grpclogging.LevelerFunc(grpcLeveler)),
			),
		},
		MaxRecvMsgSize: int(conf.General.MaxRecvMsgSize),
		MaxSendMsgSize: int(conf.General.MaxSendMsgSize),
	}
}

func grpcLeveler(ctx context.Context, fullMethod string) zapcore.Level {
	switch fullMethod {
	case "/orderer.Cluster/Step":
		return flogging.DisabledLevel
	default:
		return zapcore.InfoLevel
	}
}

func InitializeGrpcServer(conf *localconfig.TopLevel, serverConfig comm.ServerConfig) *comm.GRPCServer {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", conf.General.ListenAddress, conf.General.ListenPort))
	if err != nil {
		logger.Fatal("Failed to listen:", err)
	}

	// Create GRPC server - return if an error occurs
	grpcServer, err := comm.NewGRPCServerFromListener(lis, serverConfig)
	if err != nil {
		logger.Fatal("Failed to return new GRPC server:", err)
	}

	return grpcServer
}

func parseRequest(r *http.Request) (faultLoadBody, error) {
	var faultConfig faultLoadBody
	err := json.NewDecoder(r.Body).Decode(&faultConfig)
	return faultConfig, err
}
