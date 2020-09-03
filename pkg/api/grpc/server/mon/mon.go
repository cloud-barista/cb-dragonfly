package mon

import (
	"context"
	"fmt"

	gc "github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/common"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/logger"
	pb "github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/protobuf/cbdragonfly"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"

	"github.com/cloud-barista/cb-dragonfly/pkg/core/agent"
	coreconfig "github.com/cloud-barista/cb-dragonfly/pkg/core/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/core/metric"
)

// ===== [ Constants and Variables ] =====

// ===== [ Types ] =====

// MONService -
type MONService struct {
}

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// GetVMMonCpuInfo - 멀티 클라우드 인프라 서비스 개별 VM CPU 모니터링 정보 조회
func (s *MONService) GetVMMonCpuInfo(ctx context.Context, req *pb.VMMonQryRequest) (*pb.CpuInfoResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.GetVMMonCpuInfo()")

	res, _, err := metric.GetVMMonInfo(req.NsId, req.McisId, req.VmId, "cpu", req.PeriodType, req.StatisticsCriteria, req.Duration)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMMonCpuInfo()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.CpuInfoResponse
	err = gc.CopySrcToDest(&res, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMMonCpuInfo()")
	}

	return &grpcObj, nil
}

// GetVMMonCpuFreqInfo - 멀티 클라우드 인프라 서비스 개별 VM CPU FREQ 모니터링 정보 조회
func (s *MONService) GetVMMonCpuFreqInfo(ctx context.Context, req *pb.VMMonQryRequest) (*pb.CpuFreqInfoResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.GetVMMonCpuFreqInfo()")

	res, _, err := metric.GetVMMonInfo(req.NsId, req.McisId, req.VmId, "cpufreq", req.PeriodType, req.StatisticsCriteria, req.Duration)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMMonCpuFreqInfo()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.CpuFreqInfoResponse
	err = gc.CopySrcToDest(&res, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMMonCpuFreqInfo()")
	}

	return &grpcObj, nil
}

// GetVMMonMemoryInfo - 멀티 클라우드 인프라 서비스 개별 VM MEMORY 모니터링 정보 조회
func (s *MONService) GetVMMonMemoryInfo(ctx context.Context, req *pb.VMMonQryRequest) (*pb.MemoryInfoResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.GetVMMonMemoryInfo()")

	res, _, err := metric.GetVMMonInfo(req.NsId, req.McisId, req.VmId, "memory", req.PeriodType, req.StatisticsCriteria, req.Duration)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMMonMemoryInfo()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.MemoryInfoResponse
	err = gc.CopySrcToDest(&res, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMMonMemoryInfo()")
	}

	return &grpcObj, nil
}

// GetVMMonDiskInfo - 멀티 클라우드 인프라 서비스 개별 VM DISK 모니터링 정보 조회
func (s *MONService) GetVMMonDiskInfo(ctx context.Context, req *pb.VMMonQryRequest) (*pb.DiskInfoResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.GetVMMonDiskInfo()")

	res, _, err := metric.GetVMMonInfo(req.NsId, req.McisId, req.VmId, "disk", req.PeriodType, req.StatisticsCriteria, req.Duration)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMMonDiskInfo()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.DiskInfoResponse
	err = gc.CopySrcToDest(&res, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMMonDiskInfo()")
	}

	return &grpcObj, nil
}

// GetVMMonNetworkInfo - 멀티 클라우드 인프라 서비스 개별 VM NETWORK 모니터링 정보 조회
func (s *MONService) GetVMMonNetworkInfo(ctx context.Context, req *pb.VMMonQryRequest) (*pb.NetworkInfoResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.GetVMMonNetworkInfo()")

	res, _, err := metric.GetVMMonInfo(req.NsId, req.McisId, req.VmId, "network", req.PeriodType, req.StatisticsCriteria, req.Duration)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMMonNetworkInfo()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.NetworkInfoResponse
	err = gc.CopySrcToDest(&res, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMMonNetworkInfo()")
	}

	return &grpcObj, nil
}

// GetVMLatestMonCpuInfo - 멀티 클라우드 인프라 서비스 개별 VM CPU 최신 모니터링 정보 조회
func (s *MONService) GetVMLatestMonCpuInfo(ctx context.Context, req *pb.VMLatestMonQryRequest) (*pb.CpuRtInfoResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.GetVMLatestMonCpuInfo()")

	res, _, err := metric.GetVMLatestMonInfo(req.NsId, req.McisId, req.VmId, "cpu", req.StatisticsCriteria)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMLatestMonCpuInfo()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.CpuRtInfoResponse
	err = gc.CopySrcToDest(&res, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMLatestMonCpuInfo()")
	}

	return &grpcObj, nil
}

// GetVMLatestMonCpuFreqInfo - 멀티 클라우드 인프라 서비스 개별 VM CPU FREQ 최신 모니터링 정보 조회
func (s *MONService) GetVMLatestMonCpuFreqInfo(ctx context.Context, req *pb.VMLatestMonQryRequest) (*pb.CpuFreqRtInfoResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.GetVMLatestMonCpuFreqInfo()")

	res, _, err := metric.GetVMLatestMonInfo(req.NsId, req.McisId, req.VmId, "cpufreq", req.StatisticsCriteria)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMLatestMonCpuFreqInfo()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.CpuFreqRtInfoResponse
	err = gc.CopySrcToDest(&res, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMLatestMonCpuFreqInfo()")
	}

	return &grpcObj, nil
}

// GetVMLatestMonMemoryInfo - 멀티 클라우드 인프라 서비스 개별 VM MEMORY 최신 모니터링 정보 조회
func (s *MONService) GetVMLatestMonMemoryInfo(ctx context.Context, req *pb.VMLatestMonQryRequest) (*pb.MemoryRtInfoResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.GetVMLatestMonMemoryInfo()")

	res, _, err := metric.GetVMLatestMonInfo(req.NsId, req.McisId, req.VmId, "memory", req.StatisticsCriteria)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMLatestMonMemoryInfo()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.MemoryRtInfoResponse
	err = gc.CopySrcToDest(&res, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMLatestMonMemoryInfo()")
	}

	return &grpcObj, nil
}

// GetVMLatestMonDiskInfo - 멀티 클라우드 인프라 서비스 개별 VM DISK 최신 모니터링 정보 조회
func (s *MONService) GetVMLatestMonDiskInfo(ctx context.Context, req *pb.VMLatestMonQryRequest) (*pb.DiskRtInfoResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.GetVMLatestMonDiskInfo()")

	res, _, err := metric.GetVMLatestMonInfo(req.NsId, req.McisId, req.VmId, "disk", req.StatisticsCriteria)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMLatestMonDiskInfo()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.DiskRtInfoResponse
	err = gc.CopySrcToDest(&res, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMLatestMonDiskInfo()")
	}

	return &grpcObj, nil
}

// GetVMLatestMonNetworkInfo - 멀티 클라우드 인프라 서비스 개별 VM NETWORK 최신 모니터링 정보 조회
func (s *MONService) GetVMLatestMonNetworkInfo(ctx context.Context, req *pb.VMLatestMonQryRequest) (*pb.NetworkRtInfoResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.GetVMLatestMonNetworkInfo()")

	res, _, err := metric.GetVMLatestMonInfo(req.NsId, req.McisId, req.VmId, "network", req.StatisticsCriteria)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMLatestMonNetworkInfo()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.NetworkRtInfoResponse
	err = gc.CopySrcToDest(&res, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetVMLatestMonNetworkInfo()")
	}

	return &grpcObj, nil
}

// SetMonConfig - 모니터링 정책 설정
func (s *MONService) SetMonConfig(ctx context.Context, req *pb.MonitoringConfigRequest) (*pb.MonitoringConfigResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.SetMonConfig()")

	// GRPC 메시지에서 MON 객체로 복사
	var newMonConfig config.Monitoring
	err := gc.CopySrcToDest(&req.Item, &newMonConfig)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.SetMonConfig()")
	}

	monConfig, _, err := coreconfig.SetMonConfig(newMonConfig)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.SetMonConfig()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.MonitoringConfigInfo
	err = gc.CopySrcToDest(&monConfig, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.SetMonConfig()")
	}

	resp := &pb.MonitoringConfigResponse{Item: &grpcObj}
	return resp, nil
}

// GetMonConfig - 모니터링 정책 조회
func (s *MONService) GetMonConfig(ctx context.Context, req *pb.Empty) (*pb.MonitoringConfigResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.GetMonConfig()")

	monConfig, _, err := coreconfig.GetMonConfig()
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetMonConfig()")
	}

	fmt.Printf("\n grpc server ============= %v \n", monConfig)

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.MonitoringConfigInfo
	err = gc.CopySrcToDest(&monConfig, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.GetMonConfig()")
	}

	resp := &pb.MonitoringConfigResponse{Item: &grpcObj}
	return resp, nil
}

// ResetMonConfig - 모니터링 정책 초기화
func (s *MONService) ResetMonConfig(ctx context.Context, req *pb.Empty) (*pb.MonitoringConfigResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.ResetMonConfig()")

	monConfig, _, err := coreconfig.ResetMonConfig()
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.ResetMonConfig()")
	}

	// MON 객체에서 GRPC 메시지로 복사
	var grpcObj pb.MonitoringConfigInfo
	err = gc.CopySrcToDest(&monConfig, &grpcObj)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.ResetMonConfig()")
	}

	resp := &pb.MonitoringConfigResponse{Item: &grpcObj}
	return resp, nil
}

// InstallTelegraf - Telegraf 설치
func (s *MONService) InstallTelegraf(ctx context.Context, req *pb.InstallTelegrafRequest) (*pb.MessageResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MONService.InstallTelegraf()")

	if req.NsId == "" || req.McisId == "" || req.VmId == "" || req.PublicIp == "" || req.UserName == "" || req.SshKey == "" {
		return nil, gc.NewGrpcStatusErr("parameter is missing", "", "MONService.InstallTelegraf()")
	}

	_, err := agent.InstallTelegraf(req.NsId, req.McisId, req.VmId, req.PublicIp, req.UserName, req.SshKey)
	if err != nil {
		return nil, gc.ConvGrpcStatusErr(err, "", "MONService.InstallTelegraf()")
	}

	resp := &pb.MessageResponse{Message: "agent installation is finished"}
	return resp, nil
}
