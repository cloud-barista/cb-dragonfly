package server

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	gc "github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/common"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/logger"
	pb "github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/protobuf/cbdragonfly"
	grpc_mon "github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/server/mon"

	"google.golang.org/grpc/reflection"
)

// RunServer - 모니터링 GRPC 서버 실행
func RunServer(wg *sync.WaitGroup) {
	defer wg.Done()
	logger := logger.NewLogger()

	configPath := os.Getenv("CBMON_ROOT") + "/conf/grpc_conf.yaml"
	gConf, err := configLoad(configPath)
	if err != nil {
		logger.Error("failed to load config : ", err)
		return
	}

	dragonflysrv := gConf.GSL.DragonflySrv

	conn, err := net.Listen("tcp", dragonflysrv.Addr)
	if err != nil {
		logger.Error("failed to listen: ", err)
		return
	}

	cbserver, closer, err := gc.NewCBServer(dragonflysrv)
	if err != nil {
		logger.Error("failed to create grpc server: ", err)
		return
	}

	if closer != nil {
		defer closer.Close()
	}

	gs := cbserver.Server
	pb.RegisterMONServer(gs, &grpc_mon.MONService{})

	if dragonflysrv.Reflection == "enable" {
		if dragonflysrv.Interceptors.AuthJWT != nil {
			fmt.Printf("\n\n*** you can run reflection when jwt auth interceptor is not used ***\n\n")
		} else {
			reflection.Register(gs)
		}
	}

	fmt.Printf("\n[CB-Dragonfly: Cloud-Barista Integrated Monitoring Framework]")
	fmt.Printf("\n   Initiating GRPC API Server....__^..^__....")
	fmt.Printf("\n\n => grpc server started on %s\n\n", dragonflysrv.Addr)

	if err := gs.Serve(conn); err != nil {
		logger.Error("failed to serve: ", err)
	}
}

func configLoad(cf string) (config.GrpcConfig, error) {
	logger := logger.NewLogger()

	// Viper 를 사용하는 설정 파서 생성
	parser := config.MakeParser()

	var (
		gConf config.GrpcConfig
		err   error
	)

	if cf == "" {
		logger.Error("Please, provide the path to your configuration file")
		return gConf, errors.New("configuration file are not specified")
	}

	logger.Debug("Parsing configuration file: ", cf)
	if gConf, err = parser.GrpcParse(cf); err != nil {
		logger.Error("ERROR - Parsing the configuration file.\n", err.Error())
		return gConf, err
	}

	// Command line 에 지정된 옵션을 설정에 적용 (우선권)

	// DRAGONFLY 필수 입력 항목 체크
	dragonflysrv := gConf.GSL.DragonflySrv

	if dragonflysrv == nil {
		return gConf, errors.New("dragonflysrv field are not specified")
	}

	if dragonflysrv.Addr == "" {
		return gConf, errors.New("dragonflysrv.addr field are not specified")
	}

	if dragonflysrv.TLS != nil {
		if dragonflysrv.TLS.TLSCert == "" {
			return gConf, errors.New("dragonflysrv.tls.tls_cert field are not specified")
		}
		if dragonflysrv.TLS.TLSKey == "" {
			return gConf, errors.New("dragonflysrv.tls.tls_key field are not specified")
		}
	}

	if dragonflysrv.Interceptors != nil {
		if dragonflysrv.Interceptors.AuthJWT != nil {
			if dragonflysrv.Interceptors.AuthJWT.JWTKey == "" {
				return gConf, errors.New("dragonflysrv.interceptors.auth_jwt.jwt_key field are not specified")
			}
		}
		if dragonflysrv.Interceptors.PrometheusMetrics != nil {
			if dragonflysrv.Interceptors.PrometheusMetrics.ListenPort == 0 {
				return gConf, errors.New("dragonflysrv.interceptors.prometheus_metrics.listen_port field are not specified")
			}
		}
		if dragonflysrv.Interceptors.Opentracing != nil {
			if dragonflysrv.Interceptors.Opentracing.Jaeger != nil {
				if dragonflysrv.Interceptors.Opentracing.Jaeger.Endpoint == "" {
					return gConf, errors.New("dragonflysrv.interceptors.opentracing.jaeger.endpoint field are not specified")
				}
			}
		}
	}

	return gConf, nil
}
