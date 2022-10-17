package api

import (
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest/metric/mcis"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest/metric/mck8s"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest/topic"
	"net/http"
	"sync"

	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/cloud-barista/cb-dragonfly/pkg/util"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest/agent"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest/alert"
	restconfig "github.com/cloud-barista/cb-dragonfly/pkg/api/rest/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest/healthcheck"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/cloud-barista/cb-dragonfly/docs" // docs is generated by Swag CLI, you have to import it.
)

type APIServer struct {
	echo *echo.Echo
}

// NewAPIServer API 서버 초기화
func NewAPIServer() (*APIServer, error) {
	e := echo.New()
	apiServer := APIServer{
		echo: e,
	}
	return &apiServer, nil
}

// StartAPIServer 모니터링 API 서버 실행
func (apiServer *APIServer) StartAPIServer(wg *sync.WaitGroup) error {
	defer wg.Done()
	util.GetLogger().Info("start CB-Dragonfly Framework API Server")

	// 모니터링 API 라우팅 룰 설정
	apiServer.SetRoutingRule(apiServer.echo)

	// 모니터링 API 서버 실행
	return apiServer.echo.Start(fmt.Sprintf(":%d", config.GetInstance().Dragonfly.Port))
}

func (apiServer *APIServer) SetRoutingRule(e *echo.Echo) {
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	dragonfly := e.Group("/dragonfly")
	dragonfly.GET("/swagger/*", echoSwagger.WrapHandler)

	// 헬스체크
	dragonfly.GET("/healthcheck", healthcheck.Ping)

	// 멀티 클라우드 모니터링 정책 설정
	dragonfly.PUT("/config", restconfig.SetMonConfig)
	dragonfly.GET("/config", restconfig.GetMonConfig)
	dragonfly.PUT("/config/reset", restconfig.ResetMonConfig)

	// 멀티 클라우드 인프라 서비스 모니터링/실시간 모니터링 정보 조회
	//dragonfly.GET("/ns/:ns_id/mcis/:mcis_id/info", metric.GetMCISMonInfo)
	//dragonfly.GET("/ns/:ns_id/mcis/:mcis_id/rt-info", metric.GetMCISRealtimeMonInfo)

	// MCIS 모니터링 (Milkyway)
	dragonfly.GET("/ns/:ns_id/mcis/:mcis_id/vm/:vm_id/agent_ip/:agent_ip/mcis_metric/:metric_name/mcis-monitoring-info", mcis.GetMCISMetric)
	// 멀티클라우드 인프라 VM 온디멘드 모니터링
	dragonfly.GET("/ns/:ns/mcis/:mcis_id/vm/:vm_id/agent_ip/:agent_ip/metric/:metric_name/ondemand-monitoring-info", mcis.GetVMOnDemandMetric)
	// 멀티클라우드 인프라 네트워크 패킷 모니터링
	dragonfly.GET("/ns/:ns_id/mcis/:mcis_id/vm/:vm_id/watchtime/:watch_time/mcis-networkpacket-info", mcis.GetMCISOnDemandPacket)
	// 멀티클라우드 인프라 VM Process 모니터링
	dragonfly.GET("/agentip/:agent_ip/mcis-process-info", mcis.GetMCISOnDemandProcess)
	// 멀티클라우드 인프라 VM Spec 모니터링
	dragonfly.GET("/ns/:ns/mcis/:mcis_id/mcis-spec-info", mcis.GetMCISSpec)
	// 멀티 클라우드 인프라 VM 모니터링/실시간 모니터링 정보 조회
	dragonfly.GET("/ns/:ns_id/mcis/:mcis_id/vm/:vm_id/metric/:metric_name/info", mcis.GetVMMonInfo)
	// 멀티 클라우드 쿠버네티스 서비스 모니터링 정보 조회
	dragonfly.GET("/ns/:ns_id/mck8s/:mck8s_id/metric/:metric_name/info", mck8s.GetMCK8SMonInfo)

	// windows 에이전트 config, package 파일 다운로드
	dragonfly.GET("/installer/cbinstaller.zip", agent.GetWindowInstaller)
	dragonfly.GET("/file/agent/conf", agent.GetTelegrafConfFile)
	dragonfly.GET("/file/agent/pkg", agent.GetTelegrafPkgFile)

	// 에이전트 설치
	dragonfly.POST("/agent", agent.InstallTelegraf)
	// 에이전트 삭제
	dragonfly.DELETE("/agent", agent.UninstallAgent)

	// 에이전트 메타데이터
	dragonfly.GET("/agents/metadata", agent.ListAgentMetadata)
	dragonfly.GET("/agent/metadata", agent.GetAgentMetadata)
	dragonfly.PUT("/agent/metadata", agent.PutAgentMetadata)
	dragonfly.POST("/windows/agent/metadata", agent.CreateWindowAgentMetadata)
	dragonfly.DELETE("/windows/agent/metadata", agent.DeleteWindowAgentMetadata)

	// 삭제할 토픽 큐 등록 ( deployment collector 로 부터 삭제가 필요한 topic 들을 받기 위한 api )
	dragonfly.GET("/topic/delete/:topic", topic.AddDeleteTopicToQueue)

	// 알람 이벤트 핸들러 조회, 생성, 삭제
	dragonfly.GET("/alert/eventhandlers", alert.ListEventHandler)
	dragonfly.GET("/alert/eventhandler/type/:type/event/:name", alert.GetEventHandler)
	dragonfly.POST("/alert/eventhandler", alert.CreateEventHandler)
	dragonfly.PUT("/alert/eventhandler/type/:type/event/:name", alert.UpdateEventHandler)
	dragonfly.DELETE("/alert/eventhandler/type/:type/event/:name", alert.DeleteEventHandler)

	// 알람 조회, 생성, 삭제
	dragonfly.GET("/alert/tasks", alert.ListAlertTask)
	dragonfly.GET("/alert/task/:task_id", alert.GetAlertTask)
	dragonfly.POST("/alert/task", alert.CreateAlertTask)
	dragonfly.PUT("/alert/task/:task"+
		""+
		"_id", alert.UpdateAlertTask)
	dragonfly.DELETE("/alert/task/:task_id", alert.DeleteAlertTask)

	// 알람 이벤트 로그 조회, 생성
	dragonfly.GET("/alert/task/:task_id/events", alert.ListEventLog)
	dragonfly.POST("/alert/event", alert.CreateEventLog)

	e.Logger.Fatal(e.Start(":9090"))
}
