package manager

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest/agent"
	restconfig "github.com/cloud-barista/cb-dragonfly/pkg/api/rest/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest/metric"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
)

type APIServer struct {
	echo *echo.Echo
}

// API 서버 초기화
func NewAPIServer() (*APIServer, error) {
	e := echo.New()
	apiServer := APIServer{
		echo: e,
	}
	return &apiServer, nil
}

// 모니터링 API 서버 실행
func (apiServer *APIServer) StartAPIServer(wg *sync.WaitGroup) error {
	defer wg.Done()
	logrus.Info("Start Monitoring API Server")

	// 모니터링 API 라우팅 룰 설정
	apiServer.SetRoutingRule(apiServer.echo)

	// 모니터링 API 서버 실행
	return apiServer.echo.Start(fmt.Sprintf(":%d", config.GetInstance().APIServer.Port))
}

func (apiServer *APIServer) SetRoutingRule(e *echo.Echo) {

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// 멀티 클라우드 인프라 서비스 모니터링/실시간 모니터링 정보 조회
	e.GET("/dragonfly/ns/:ns/mcis/:mcis_id/info", metric.GetMCISMonInfo)
	e.GET("/dragonfly/ns/:ns/mcis/:mcis_id/rt-info", metric.GetMCISRealtimeMonInfo)

	// 멀티 클라우드 인프라 VM 모니터링/실시간 모니터링 정보 조회
	e.GET("/dragonfly/ns/:ns/mcis/:mcis_id/vm/:vm_id/metric/:metric_name/info", metric.GetVMMonInfo)
	e.GET("/dragonfly/ns/:ns/mcis/:mcis_id/vm/:vm_id/metric/:metric_name/rt-info", metric.GetVMLatestMonInfo)

	// 멀티 클라우드 모니터링 정책 설정
	e.PUT("/dragonfly/config", restconfig.SetMonConfig)
	e.GET("/dragonfly/config", restconfig.GetMonConfig)
	e.PUT("/dragonfly/config/reset", restconfig.ResetMonConfig)

	// 에이전트 설치 스크립트 다운로드
	//e.GET("/dragonfly/file/agent/install", apiServer.GetTelegrafInstallScript)

	// 에이전트 config, package 파일 다운로드
	//e.GET("/dragonfly/file/agent/conf", apiServer.GetTelegrafConfFile)
	//e.GET("/dragonfly/file/agent/pkg", apiServer.GetTelegrafPkgFile)

	// 에이전트 설치
	e.POST("/dragonfly/agent/install", agent.InstallTelegraf)

	// 멀티클라우드 인프라 VM 온디멘드 모니터링
	e.GET("/dragonfly/ns/:ns/mcis/:mcis_id/vm/:vm_id/agent_ip/:agent_ip", metric.OndemandAllMetric)
	e.GET("/dragonfly/ns/:ns/mcis/:mcis_id/vm/:vm_id/agent_ip/:agent_ip/metric/:metric_name", metric.OndemandMetric)

	// windows 에이전트 config, package 파일 다운로드
	e.GET("/dragonfly/installer/cbinstaller.zip", agent.GetWindowInstaller)
	e.GET("/dragonfly/file/agent/conf", agent.GetTelegrafConfFile)
	e.GET("/dragonfly/file/agent/pkg", agent.GetTelegrafPkgFile)
}
