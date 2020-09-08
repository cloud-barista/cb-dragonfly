package request

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	gc "github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/common"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/logger"
	pb "github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/protobuf/cbdragonfly"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/request/mon"

	"google.golang.org/grpc"
)

// ===== [ Constants and Variables ] =====

// ===== [ Types ] =====

// MONApi - MON API 구조 정의
type MONApi struct {
	gConf        *config.GrpcConfig
	conn         *grpc.ClientConn
	jaegerCloser io.Closer
	clientMON    pb.MONClient
	requestMON   *mon.MONRequest
	inType       string
	outType      string
}

// MonitoringReq - 모니터링 설정 구조 정의
type MonitoringReq struct {
	AgentInterval      int `yaml:"agent_interval" json:"agent_interval"`
	CollectorInterval  int `yaml:"collector_interval" json:"collector_interval"`
	SchedulingInterval int `yaml:"schedule_interval" json:"schedule_interval"`
	MaxHostCount       int `yaml:"max_host_count" json:"max_host_count"`
	AgentTTL           int `yaml:"agent_TTL" json:"agent_TTL"`
}

// ===== [ Implementations ] =====

// SetServerAddr - Dragonfly 서버 주소 설정
func (m *MONApi) SetServerAddr(addr string) error {
	if addr == "" {
		return errors.New("parameter is empty")
	}

	m.gConf.GSL.DragonflyCli.ServerAddr = addr
	return nil
}

// GetServerAddr - Dragonfly 서버 주소 값 조회
func (m *MONApi) GetServerAddr() (string, error) {
	return m.gConf.GSL.DragonflyCli.ServerAddr, nil
}

// SetTLSCA - TLS CA 설정
func (m *MONApi) SetTLSCA(tlsCAFile string) error {
	if tlsCAFile == "" {
		return errors.New("parameter is empty")
	}

	if m.gConf.GSL.DragonflyCli.TLS == nil {
		m.gConf.GSL.DragonflyCli.TLS = &config.TLSConfig{}
	}

	m.gConf.GSL.DragonflyCli.TLS.TLSCA = tlsCAFile
	return nil
}

// GetTLSCA - TLS CA 값 조회
func (m *MONApi) GetTLSCA() (string, error) {
	if m.gConf.GSL.DragonflyCli.TLS == nil {
		return "", nil
	}

	return m.gConf.GSL.DragonflyCli.TLS.TLSCA, nil
}

// SetTimeout - Timeout 설정
func (m *MONApi) SetTimeout(timeout time.Duration) error {
	m.gConf.GSL.DragonflyCli.Timeout = timeout
	return nil
}

// GetTimeout - Timeout 값 조회
func (m *MONApi) GetTimeout() (time.Duration, error) {
	return m.gConf.GSL.DragonflyCli.Timeout, nil
}

// SetJWTToken - JWT 인증 토큰 설정
func (m *MONApi) SetJWTToken(token string) error {
	if token == "" {
		return errors.New("parameter is empty")
	}

	if m.gConf.GSL.DragonflyCli.Interceptors == nil {
		m.gConf.GSL.DragonflyCli.Interceptors = &config.InterceptorsConfig{}
		m.gConf.GSL.DragonflyCli.Interceptors.AuthJWT = &config.AuthJWTConfig{}
	}
	if m.gConf.GSL.DragonflyCli.Interceptors.AuthJWT == nil {
		m.gConf.GSL.DragonflyCli.Interceptors.AuthJWT = &config.AuthJWTConfig{}
	}

	m.gConf.GSL.DragonflyCli.Interceptors.AuthJWT.JWTToken = token
	return nil
}

// GetJWTToken - JWT 인증 토큰 값 조회
func (m *MONApi) GetJWTToken() (string, error) {
	if m.gConf.GSL.DragonflyCli.Interceptors == nil {
		return "", nil
	}
	if m.gConf.GSL.DragonflyCli.Interceptors.AuthJWT == nil {
		return "", nil
	}

	return m.gConf.GSL.DragonflyCli.Interceptors.AuthJWT.JWTToken, nil
}

// SetConfigPath - 환경설정 파일 설정
func (m *MONApi) SetConfigPath(configFile string) error {
	logger := logger.NewLogger()

	// Viper 를 사용하는 설정 파서 생성
	parser := config.MakeParser()

	var (
		gConf config.GrpcConfig
		err   error
	)

	if configFile == "" {
		logger.Error("Please, provide the path to your configuration file")
		return errors.New("configuration file are not specified")
	}

	logger.Debug("Parsing configuration file: ", configFile)
	if gConf, err = parser.GrpcParse(configFile); err != nil {
		logger.Error("ERROR - Parsing the configuration file.\n", err.Error())
		return err
	}

	// DRAGONFLY CLIENT 필수 입력 항목 체크
	dragonflycli := gConf.GSL.DragonflyCli

	if dragonflycli == nil {
		return errors.New("dragonflycli field are not specified")
	}

	if dragonflycli.ServerAddr == "" {
		return errors.New("dragonflycli.server_addr field are not specified")
	}

	if dragonflycli.Timeout == 0 {
		dragonflycli.Timeout = 90 * time.Second
	}

	if dragonflycli.TLS != nil {
		if dragonflycli.TLS.TLSCA == "" {
			return errors.New("dragonflycli.tls.tls_ca field are not specified")
		}
	}

	if dragonflycli.Interceptors != nil {
		if dragonflycli.Interceptors.AuthJWT != nil {
			if dragonflycli.Interceptors.AuthJWT.JWTToken == "" {
				return errors.New("dragonflycli.interceptors.auth_jwt.jwt_token field are not specified")
			}
		}
		if dragonflycli.Interceptors.Opentracing != nil {
			if dragonflycli.Interceptors.Opentracing.Jaeger != nil {
				if dragonflycli.Interceptors.Opentracing.Jaeger.Endpoint == "" {
					return errors.New("dragonflycli.interceptors.opentracing.jaeger.endpoint field are not specified")
				}
			}
		}
	}

	m.gConf = &gConf
	return nil
}

// Open - 연결 설정
func (m *MONApi) Open() error {

	dragonflycli := m.gConf.GSL.DragonflyCli

	// grpc 커넥션 생성
	cbconn, closer, err := gc.NewCBConnection(dragonflycli)
	if err != nil {
		return err
	}

	if closer != nil {
		m.jaegerCloser = closer
	}

	m.conn = cbconn.Conn

	// grpc 클라이언트 생성
	m.clientMON = pb.NewMONClient(m.conn)

	// grpc 호출 Wrapper
	m.requestMON = &mon.MONRequest{Client: m.clientMON, Timeout: dragonflycli.Timeout, InType: m.inType, OutType: m.outType}

	return nil
}

// Close - 연결 종료
func (m *MONApi) Close() {
	if m.conn != nil {
		m.conn.Close()
	}
	if m.jaegerCloser != nil {
		m.jaegerCloser.Close()
	}

	m.jaegerCloser = nil
	m.conn = nil
	m.clientMON = nil
	m.requestMON = nil
}

// SetInType - 입력 문서 타입 설정 (json/yaml)
func (m *MONApi) SetInType(in string) error {
	if in == "json" {
		m.inType = in
	} else if in == "yaml" {
		m.inType = in
	} else {
		return errors.New("input type is not supported")
	}

	if m.requestMON != nil {
		m.requestMON.InType = m.inType
	}

	return nil
}

// GetInType - 입력 문서 타입 값 조회
func (m *MONApi) GetInType() (string, error) {
	return m.inType, nil
}

// SetOutType - 출력 문서 타입 설정 (json/yaml)
func (m *MONApi) SetOutType(out string) error {
	if out == "json" {
		m.outType = out
	} else if out == "yaml" {
		m.outType = out
	} else {
		return errors.New("output type is not supported")
	}

	if m.requestMON != nil {
		m.requestMON.OutType = m.outType
	}

	return nil
}

// GetOutType - 출력 문서 타입 값 조회
func (m *MONApi) GetOutType() (string, error) {
	return m.outType, nil
}

// GetVMMonCpuInfo - 멀티 클라우드 인프라 서비스 개별 VM CPU 모니터링 정보 조회
func (m *MONApi) GetVMMonCpuInfo(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.GetVMMonCpuInfo()
}

// GetVMMonCpuInfoByParam - 멀티 클라우드 인프라 서비스 개별 VM CPU 모니터링 정보 조회
func (m *MONApi) GetVMMonCpuInfoByParam(nsId string, mcisId string, vmId string, periodType string, statisticsCriteria string, duration string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	m.requestMON.InData = `{"ns_id":"` + nsId + `", "mcis_id":"` + mcisId + `", "vm_id":"` + vmId + `", "periodType":"` + periodType + `", "statisticsCriteria":"` + statisticsCriteria + `", "duration":"` + duration + `"}`
	result, err := m.requestMON.GetVMMonCpuInfo()
	m.SetInType(holdType)

	return result, err
}

// GetVMMonCpuFreqInfo - 멀티 클라우드 인프라 서비스 개별 VM CPU FREQ 모니터링 정보 조회
func (m *MONApi) GetVMMonCpuFreqInfo(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.GetVMMonCpuFreqInfo()
}

// GetVMMonCpuFreqInfoByParam - 멀티 클라우드 인프라 서비스 개별 VM CPU FREQ 모니터링 정보 조회
func (m *MONApi) GetVMMonCpuFreqInfoByParam(nsId string, mcisId string, vmId string, periodType string, statisticsCriteria string, duration string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	m.requestMON.InData = `{"ns_id":"` + nsId + `", "mcis_id":"` + mcisId + `", "vm_id":"` + vmId + `", "periodType":"` + periodType + `", "statisticsCriteria":"` + statisticsCriteria + `", "duration":"` + duration + `"}`
	result, err := m.requestMON.GetVMMonCpuFreqInfo()
	m.SetInType(holdType)

	return result, err
}

// GetVMMonMemoryInfo - 멀티 클라우드 인프라 서비스 개별 VM MEMORY 모니터링 정보 조회
func (m *MONApi) GetVMMonMemoryInfo(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.GetVMMonMemoryInfo()
}

// GetVMMonMemoryInfoByParam - 멀티 클라우드 인프라 서비스 개별 VM MEMORY 모니터링 정보 조회
func (m *MONApi) GetVMMonMemoryInfoByParam(nsId string, mcisId string, vmId string, periodType string, statisticsCriteria string, duration string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	m.requestMON.InData = `{"ns_id":"` + nsId + `", "mcis_id":"` + mcisId + `", "vm_id":"` + vmId + `", "periodType":"` + periodType + `", "statisticsCriteria":"` + statisticsCriteria + `", "duration":"` + duration + `"}`
	result, err := m.requestMON.GetVMMonMemoryInfo()
	m.SetInType(holdType)

	return result, err
}

// GetVMMonDiskInfo - 멀티 클라우드 인프라 서비스 개별 VM DISK 모니터링 정보 조회
func (m *MONApi) GetVMMonDiskInfo(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.GetVMMonDiskInfo()
}

// GetVMMonDiskInfoByParam - 멀티 클라우드 인프라 서비스 개별 VM DISK 모니터링 정보 조회
func (m *MONApi) GetVMMonDiskInfoByParam(nsId string, mcisId string, vmId string, periodType string, statisticsCriteria string, duration string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	m.requestMON.InData = `{"ns_id":"` + nsId + `", "mcis_id":"` + mcisId + `", "vm_id":"` + vmId + `", "periodType":"` + periodType + `", "statisticsCriteria":"` + statisticsCriteria + `", "duration":"` + duration + `"}`
	result, err := m.requestMON.GetVMMonDiskInfo()
	m.SetInType(holdType)

	return result, err
}

// GetVMMonNetworkInfo - 멀티 클라우드 인프라 서비스 개별 VM NETWORK 모니터링 정보 조회
func (m *MONApi) GetVMMonNetworkInfo(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.GetVMMonNetworkInfo()
}

// GetVMMonNetworkInfoByParam - 멀티 클라우드 인프라 서비스 개별 VM NETWORK 모니터링 정보 조회
func (m *MONApi) GetVMMonNetworkInfoByParam(nsId string, mcisId string, vmId string, periodType string, statisticsCriteria string, duration string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	m.requestMON.InData = `{"ns_id":"` + nsId + `", "mcis_id":"` + mcisId + `", "vm_id":"` + vmId + `", "periodType":"` + periodType + `", "statisticsCriteria":"` + statisticsCriteria + `", "duration":"` + duration + `"}`
	result, err := m.requestMON.GetVMMonNetworkInfo()
	m.SetInType(holdType)

	return result, err
}

// GetVMLatestMonCpuInfo - 멀티 클라우드 인프라 서비스 개별 VM CPU 최신 모니터링 정보 조회
func (m *MONApi) GetVMLatestMonCpuInfo(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.GetVMLatestMonCpuInfo()
}

// GetVMLatestMonCpuInfoByParam - 멀티 클라우드 인프라 서비스 개별 VM CPU 최신 모니터링 정보 조회
func (m *MONApi) GetVMLatestMonCpuInfoByParam(nsId string, mcisId string, vmId string, statisticsCriteria string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	m.requestMON.InData = `{"ns_id":"` + nsId + `", "mcis_id":"` + mcisId + `", "vm_id":"` + vmId + `", "statisticsCriteria":"` + statisticsCriteria + `"}`
	result, err := m.requestMON.GetVMLatestMonCpuInfo()
	m.SetInType(holdType)

	return result, err
}

// GetVMLatestMonCpuFreqInfo - 멀티 클라우드 인프라 서비스 개별 VM CPU FREQ 최신 모니터링 정보 조회
func (m *MONApi) GetVMLatestMonCpuFreqInfo(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.GetVMLatestMonCpuFreqInfo()
}

// GetVMLatestMonCpuFreqInfoByParam - 멀티 클라우드 인프라 서비스 개별 VM CPU FREQ 최신 모니터링 정보 조회
func (m *MONApi) GetVMLatestMonCpuFreqInfoByParam(nsId string, mcisId string, vmId string, statisticsCriteria string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	m.requestMON.InData = `{"ns_id":"` + nsId + `", "mcis_id":"` + mcisId + `", "vm_id":"` + vmId + `", "statisticsCriteria":"` + statisticsCriteria + `"}`
	result, err := m.requestMON.GetVMLatestMonCpuFreqInfo()
	m.SetInType(holdType)

	return result, err
}

// GetVMLatestMonMemoryInfo - 멀티 클라우드 인프라 서비스 개별 VM MEMORY 최신 모니터링 정보 조회
func (m *MONApi) GetVMLatestMonMemoryInfo(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.GetVMLatestMonMemoryInfo()
}

// GetVMLatestMonMemoryInfoByParam - 멀티 클라우드 인프라 서비스 개별 VM MEMORY 최신 모니터링 정보 조회
func (m *MONApi) GetVMLatestMonMemoryInfoByParam(nsId string, mcisId string, vmId string, statisticsCriteria string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	m.requestMON.InData = `{"ns_id":"` + nsId + `", "mcis_id":"` + mcisId + `", "vm_id":"` + vmId + `", "statisticsCriteria":"` + statisticsCriteria + `"}`
	result, err := m.requestMON.GetVMLatestMonMemoryInfo()
	m.SetInType(holdType)

	return result, err
}

// GetVMLatestMonDiskInfo - 멀티 클라우드 인프라 서비스 개별 VM DISK 최신 모니터링 정보 조회
func (m *MONApi) GetVMLatestMonDiskInfo(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.GetVMLatestMonDiskInfo()
}

// GetVMLatestMonDiskInfoByParam - 멀티 클라우드 인프라 서비스 개별 VM DISK 최신 모니터링 정보 조회
func (m *MONApi) GetVMLatestMonDiskInfoByParam(nsId string, mcisId string, vmId string, statisticsCriteria string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	m.requestMON.InData = `{"ns_id":"` + nsId + `", "mcis_id":"` + mcisId + `", "vm_id":"` + vmId + `", "statisticsCriteria":"` + statisticsCriteria + `"}`
	result, err := m.requestMON.GetVMLatestMonDiskInfo()
	m.SetInType(holdType)

	return result, err
}

// GetVMLatestMonNetworkInfo - 멀티 클라우드 인프라 서비스 개별 VM NETWORK 최신 모니터링 정보 조회
func (m *MONApi) GetVMLatestMonNetworkInfo(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.GetVMLatestMonNetworkInfo()
}

// GetVMLatestMonNetworkInfoByParam - 멀티 클라우드 인프라 서비스 개별 VM NETWORK 최신 모니터링 정보 조회
func (m *MONApi) GetVMLatestMonNetworkInfoByParam(nsId string, mcisId string, vmId string, statisticsCriteria string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	m.requestMON.InData = `{"ns_id":"` + nsId + `", "mcis_id":"` + mcisId + `", "vm_id":"` + vmId + `", "statisticsCriteria":"` + statisticsCriteria + `"}`
	result, err := m.requestMON.GetVMLatestMonNetworkInfo()
	m.SetInType(holdType)

	return result, err
}

// SetMonConfig - 모니터링 정책 설정
func (m *MONApi) SetMonConfig(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.SetMonConfig()
}

// SetMonConfigByParam - 모니터링 정책 설정
func (m *MONApi) SetMonConfigByParam(req *MonitoringReq) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	j, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	m.requestMON.InData = string(j)
	result, err := m.requestMON.SetMonConfig()
	m.SetInType(holdType)

	return result, err
}

// GetMonConfig - 모니터링 정책 조회
func (m *MONApi) GetMonConfig() (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	return m.requestMON.GetMonConfig()
}

// ResetMonConfig - 모니터링 정책 초기화
func (m *MONApi) ResetMonConfig() (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	return m.requestMON.ResetMonConfig()
}

// InstallTelegraf - Telegraf 설치
func (m *MONApi) InstallTelegraf(doc string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	m.requestMON.InData = doc
	return m.requestMON.InstallTelegraf()
}

// InstallTelegrafByParam - Telegraf 설치
func (m *MONApi) InstallTelegrafByParam(nsId string, mcisId string, vmId string, publicIp string, userName string, sshKey string) (string, error) {
	if m.requestMON == nil {
		return "", errors.New("The Open() function must be called")
	}

	holdType, _ := m.GetInType()
	m.SetInType("json")
	m.requestMON.InData = `{"ns_id":"` + nsId + `", "mcis_id":"` + mcisId + `", "vm_id":"` + vmId + `", "public_ip":"` + publicIp + `", "user_name":"` + userName + `", "ssh_key":"` + sshKey + `"}`
	result, err := m.requestMON.InstallTelegraf()
	m.SetInType(holdType)

	return result, err
}

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// NewMONManager - MON API 객체 생성
func NewMONManager() (m *MONApi) {

	m = &MONApi{}
	m.gConf = &config.GrpcConfig{}
	m.gConf.GSL.DragonflyCli = &config.GrpcClientConfig{}

	m.jaegerCloser = nil
	m.conn = nil
	m.clientMON = nil
	m.requestMON = nil

	m.inType = "json"
	m.outType = "json"

	return
}
