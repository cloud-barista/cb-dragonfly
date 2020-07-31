package manager

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/core"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest/agent"
	restconfig "github.com/cloud-barista/cb-dragonfly/pkg/api/rest/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest/metric"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	metricstore "github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdbv1"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore"
)

type APIServer struct {
	echo   *echo.Echo
	config config.Config
}

// API 서버 초기화
func NewAPIServer(config config.Config, influxDB metricstore.Storage, etcd realtimestore.Storage) (*APIServer, error) {
	e := echo.New()
	apiServer := APIServer{
		echo:   e,
		config: config,
	}
	core.InitCoreConfig(config, influxDB, etcd)
	return &apiServer, nil
}

// 모니터링 API 서버 실행
func (apiServer *APIServer) StartAPIServer(wg *sync.WaitGroup) error {
	defer wg.Done()
	logrus.Info("Start Monitoring API Server")

	// 모니터링 API 라우팅 룰 설정
	apiServer.SetRoutingRule(apiServer.echo)

	// 모니터링 API 서버 실행
	return apiServer.echo.Start(fmt.Sprintf(":%d", apiServer.config.APIServer.Port))
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
	e.GET("/dragonfly/ns/:ns/mcis/:mcis_id/vm/:vm_id/metric/:metric_name/rt-info", metric.GetVMRealtimeMonInfo)

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
}

// Telegraf agent 설치 스크립트 파일 다운로드
/*func (apiServer *APIServer) GetTelegrafInstallScript(c echo.Context) error {
	// Query 파라미터 가져오기
	mcisId := c.QueryParam("mcis_id")
	vmId := c.QueryParam("vm_id")

	// Query 파라미터 값 체크
	if mcisId == "" || vmId == "" {
		err := errors.New("failed to get package. query parameter is missing")
		return c.JSON(http.StatusInternalServerError, err)
	}

	collectorServer := fmt.Sprintf("%s:%d", apiServer.config.CollectManager.CollectorIP, apiServer.config.APIServer.Port)

	rootPath := os.Getenv("CBMON_ROOT")
	filePath := rootPath + "/file/install_agent.sh"

	read, err := ioutil.ReadFile(filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// 파일 내의 변수 값 설정 (vmId, collectorServer)
	strConf := string(read)
	strConf = strings.ReplaceAll(strConf, "{{mcis_id}}", mcisId)
	strConf = strings.ReplaceAll(strConf, "{{vm_id}}", vmId)
	strConf = strings.ReplaceAll(strConf, "{{api_server}}", collectorServer)

	return c.Blob(http.StatusOK, "text/plain", []byte(strConf))
}*/

// Telegraf config 파일 다운로드
/*func (apiServer *APIServer) GetTelegrafConfFile(c echo.Context) error {
	// Query 파라미터 가져오기
	mcisId := c.QueryParam("mcis_id")
	vmId := c.QueryParam("vm_id")

	// Query 파라미터 값 체크
	if mcisId == "" || vmId == "" {
		err := errors.New("failed to get package. query parameter is missing")
		return c.JSON(http.StatusInternalServerError, err)
	}

	collectorServer := fmt.Sprintf("udp://%s:%d", apiServer.manager.Config.CollectManager.CollectorIP, apiServer.manager.Config.CollectManager.CollectorPort)
	influxDBServer := fmt.Sprintf("http://%s:8086", apiServer.manager.Config.CollectManager.CollectorIP)

	rootPath := os.Getenv("CBMON_ROOT")
	filePath := rootPath + "/file/conf/telegraf.conf"

	read, err := ioutil.ReadFile(filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// 파일 내의 변수 값 설정 (hostId, collectorServer)
	strConf := string(read)
	strConf = strings.ReplaceAll(strConf, "{{mcis_id}}", mcisId)
	strConf = strings.ReplaceAll(strConf, "{{vm_id}}", vmId)
	strConf = strings.ReplaceAll(strConf, "{{collector_server}}", collectorServer)
	strConf = strings.ReplaceAll(strConf, "{{influxdb_server}}", influxDBServer)

	return c.Blob(http.StatusOK, "text/plain", []byte(strConf))
}*/

// Telegraf package 파일 다운로드
/*func (apiServer *APIServer) GetTelegrafPkgFile(c echo.Context) error {
	// Query 파라미터 가져오기
	osType := c.QueryParam("osType")
	arch := c.QueryParam("arch")

	// Query 파라미터 값 체크
	if osType == "" || arch == "" {
		err := errors.New("failed to get package. query parameter is missing")
		return c.JSON(http.StatusInternalServerError, err)
	}

	// osType, architecture 지원 여부 체크
	osType = strings.ToLower(osType)
	if osType != "ubuntu" && osType != "centos" {
		err := errors.New("failed to get package. not supported OS type")
		return c.JSON(http.StatusInternalServerError, err)
	}
	if !strings.Contains(arch, "32") && !strings.Contains(arch, "64") {
		err := errors.New("failed to get package. not supported architecture")
		return c.JSON(http.StatusInternalServerError, err)
	}

	if strings.Contains(arch, "64") {
		arch = "x64"
	} else {
		arch = "x32"
	}

	rootPath := os.Getenv("CBMON_ROOT")
	var filePath string
	switch osType {
	case "ubuntu":
		filePath = rootPath + fmt.Sprintf("/file/pkg/%s/%s/telegraf_1.15.0~c78045c1-0_amd64.deb", osType, arch)
	case "centos":
		filePath = rootPath + fmt.Sprintf("/file/pkg/%s/%s/telegraf-1.12.0~f09f2b5-0.x86_64.rpm", osType, arch)
	default:
		err := errors.New(fmt.Sprintf("failed to get package. osType %s not supported", osType))
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.File(filePath)
}*/
