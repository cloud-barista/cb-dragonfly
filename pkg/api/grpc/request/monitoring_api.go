package request

import (
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/protobuf/cbdragonfly"
)

const (
	TimeoutMinutes = 5
	ConvertType    = "json"
)

var once sync.Once
var monApi MonitoringAPI

type MonitoringAPI struct {
	conn       *grpc.ClientConn
	monClient  *pb.MONClient
	monRequest *MonitoringRequest
	timeout    time.Duration
	inType     string
	outType    string
}

//func InitMonitoringAPI(timeout *time.Duration, inType string, outType string) MonitoringAPI {
func InitMonitoringAPI() *MonitoringAPI {
	// initialize monitoring api
	once.Do(func() {
		monApi = MonitoringAPI{
			timeout: TimeoutMinutes * time.Minute,
			inType:  ConvertType,
			outType: ConvertType,
		}
	})
	return &monApi
}

func GetMonitoringAPI() *MonitoringAPI {
	return InitMonitoringAPI()
}

func (monApi *MonitoringAPI) Open() error {
	// connect to grpc server
	clientConn, err := ConnectGRPC()
	if err != nil {
		return err
	}
	monApi.conn = clientConn

	// connect to grpc service (monitoring)
	monClient := ConnectClient(clientConn)
	monApi.monClient = monClient

	// initialize monitoring request
	monReq := NewMonitoringRequest(monClient, monApi.timeout, monApi.inType, monApi.outType)
	monApi.monRequest = monReq

	return nil
}

func (monApi *MonitoringAPI) Close() {
	// disconnect to grpc server
	if monApi.conn != nil {
		monApi.conn.Close()
	}
	// finalize monitoring api properties
	monApi.conn = nil
	monApi.monClient = nil
	monApi.monRequest = nil
}

func (monApi *MonitoringAPI) GetMonitoringConfig() (string, error) {
	return monApi.monRequest.GetMonitoringConfig()
}

func (monApi *MonitoringAPI) SetMonitoringConfig(reqParams pb.MonitoringConfigInfo) (string, error) {
	return monApi.monRequest.SetMonitoringConfig(reqParams)
}

func (monApi *MonitoringAPI) ResetMonitoringConfig() (string, error) {
	return monApi.monRequest.ResetMonitoringConfig()
}

func (monApi *MonitoringAPI) GetVMMonInfo(metricName string, vmMonQueryRequest pb.VMMonQryRequest) (string, error) {
	return monApi.monRequest.GetVMMonInfo(metricName, vmMonQueryRequest)
}

func (monApi *MonitoringAPI) GetVMOnDemandMonInfo(metricName string, vmMonQueryRequest pb.VMOnDemandMonQryRequest) (string, error) {
	return monApi.monRequest.GetVMOnDemandMonInfo(metricName, vmMonQueryRequest)
}

func (monApi *MonitoringAPI) GetMCISMonInfo(mcisMonQueryRequest pb.VMMCISMonQryRequest) (string, error) {
	return monApi.monRequest.GetMCISMonInfo(mcisMonQueryRequest)
}
