package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloud-barista/cb-spider/cloud-control-manager/vm-ssh"
	"github.com/google/uuid"
	"github.com/influxdata/influxdb1-client/models"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/client"

	"github.com/cloud-barista/cb-dragonfly/pkg/collector"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
)

type APIServer struct {
	echo       *echo.Echo
	config     Config
	aggregator *collector.Aggregator
	Etcd       realtimestore.Storage
	InfluxDB   metricstore.Storage
}

// API 서버 초기화
func NewAPIServer(config Config, aggregator *collector.Aggregator, influxDB metricstore.Storage, etcd realtimestore.Storage) (*APIServer, error) {
	e := echo.New()
	apiServer := APIServer{
		echo:       e,
		config:     config,
		aggregator: aggregator,
		InfluxDB:   influxDB,
		Etcd:       etcd,
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
	return apiServer.echo.Start(fmt.Sprintf(":%d", apiServer.config.APIServer.Port))
}

func (apiServer *APIServer) SetRoutingRule(e *echo.Echo) {

	// 멀티 클라우드 인프라 서비스 모니터링/실시간 모니터링 정보 조회
	e.GET("/dragonfly/mcis/:mcis_id/info", apiServer.GetMCISMonInfo)
	e.GET("/dragonfly/mcis/:mcis_id/rt-info", apiServer.GetMCISRealtimeMonInfo)

	// 멀티 클라우드 인프라 VM 모니터링/실시간 모니터링 정보 조회
	e.GET("/dragonfly/mcis/:mcis_id/vm/:vm_id/metric/:metric_name/info", apiServer.GetVMMonInfo)
	e.GET("/dragonfly/mcis/:mcis_id/vm/:vm_id/metric/:metric_name/rt-info", apiServer.GetVMRealtimeMonInfo)

	// 멀티 클라우드 모니터링 정책 설정
	e.PUT("/dragonfly/config", apiServer.SetMonConfig)
	e.GET("/dragonfly/config", apiServer.GetMonConfig)
	e.PUT("/dragonfly/config/reset", apiServer.ResetMonConfig)

	// 에이전트 설치 스크립트 다운로드
	e.GET("/dragonfly/file/agent/install", apiServer.GetTelegrafInstallScript)

	// 에이전트 config, package 파일 다운로드
	e.GET("/dragonfly/file/agent/conf", apiServer.GetTelegrafConfFile)
	e.GET("/dragonfly/file/agent/pkg", apiServer.GetTelegrafPkgFile)

	// 에이전트 설치
	e.POST("/dragonfly/agent/install", apiServer.InstallTelegraf)
}

// 멀티 클라우드 인프라 서비스 모니터링 정보 조회
func (apiServer *APIServer) GetMCISMonInfo(c echo.Context) error {
	mcisId := c.Param("mcis_id")
	fmt.Println(mcisId)

	// TODO: MCIS 서비스 모니터링 정보 조회 기능 개발

	return c.JSON(http.StatusOK, "")
}

// 멀티 클라우드 인프라 서비스(MCIS) 실시간 모니터링 정보 조회
func (apiServer *APIServer) GetMCISRealtimeMonInfo(c echo.Context) error {
	mcisId := c.Param("mcis_id")
	fmt.Println(mcisId)

	// TODO: MCIS 서비스 실시간 모니터링 정보 조회 기능 개발

	return c.JSON(http.StatusOK, "")
}

// 멀티 클라우드 인프라 VM 모니터링 정보 조회
func (apiServer *APIServer) GetVMMonInfo(c echo.Context) error {

	// Path 파라미터 가져오기
	//mcisId := c.Param("mcis_id")
	vmId := c.Param("vm_id")
	metricName := c.Param("metric_name")

	// Query 파라미터 가져오기
	period := c.QueryParam("periodType")
	aggregateType := c.QueryParam("statisticsCriteria")
	duration := c.QueryParam("duration")

	var metricKey string

	switch metricName {
	case "cpu":

		// cpu 메트릭 조회
		metricKey = "cpu"
		cpuMetric, err := apiServer.InfluxDB.ReadMetric(vmId, metricKey, period, aggregateType, duration)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		resultMetric, err := metricstore.MappingMonMetric(metricKey, &cpuMetric)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		return c.JSON(http.StatusOK, resultMetric)

	case "memory":

		// memory 메트릭 조회
		metricKey := "mem"
		memMetric, err := apiServer.InfluxDB.ReadMetric(vmId, metricKey, period, aggregateType, duration)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		resultMetric, err := metricstore.MappingMonMetric(metricKey, &memMetric)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		return c.JSON(http.StatusOK, resultMetric)

	case "disk":

		// disk, diskio 메트릭 조회
		metricKey := "disk"
		diskMetric, err := apiServer.InfluxDB.ReadMetric(vmId, metricKey, period, aggregateType, duration)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		metricKey = "diskio"
		diskIoMetric, err := apiServer.InfluxDB.ReadMetric(vmId, metricKey, period, aggregateType, duration)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		if diskMetric == nil || diskIoMetric == nil {
			return c.JSON(http.StatusNotFound, err)
		}

		diskRow := diskMetric.(models.Row)
		diskIoRow := diskIoMetric.(models.Row)

		// Aggregate Column Info
		//diskRow.Columns = append(diskRow.Columns, diskIoRow.Columns[1:]...)

		// Aggregate Monitoring data
		/*initArr := make([]interface{}, 4)
		for idx := range diskRow.Values {
			if idx <= len(diskIoRow.Values) {
				diskRow.Values[idx] = append(diskRow.Values[idx], diskIoRow.Values[idx][1:]...)
				continue
			}
			diskRow.Values[idx] = append(diskRow.Values[idx], initArr...)
		}*/

		// Aggregate Metric
		var resultRow models.Row
		resultRow.Name = "disk"
		resultRow.Tags = diskRow.Tags
		resultRow.Columns = append(resultRow.Columns, diskRow.Columns[0:]...)
		resultRow.Columns = append(resultRow.Columns, diskIoRow.Columns[1:]...)

		// TimePoint 맵 생성 (disk, diskio 메트릭)
		timePointMap := make(map[string]string, len(diskRow.Values))
		for _, val := range diskRow.Values {
			timePoint := val[0].(string)
			timePointMap[timePoint] = timePoint
		}
		for _, val := range diskIoRow.Values {
			timePoint := val[0].(string)
			if tp, exist := timePointMap[timePoint]; !exist {
				timePointMap[tp] = tp
			}
		}

		// TimePoint 배열 생성
		idx := 0
		timePointArr := make([]string, len(timePointMap))
		for _, tp := range timePointMap {
			timePointArr[idx] = tp
			idx++
		}
		sort.Strings(timePointArr)

		// TimePoint 배열 기준 모니터링 메트릭 Aggregate
		for _, tp := range timePointArr {

			metricVal := make([]interface{}, 1)
			metricVal[0] = tp

			// disk 메트릭 aggregate
			diskMetricAdded := false
			for idx, val := range diskRow.Values {
				t := val[0].(string)
				if strings.EqualFold(t, tp) {
					metricVal = append(metricVal, val[1:]...)
					diskMetricAdded = true
					break
				}
				// 해당 TimePoint에 해당하는 disk 메트릭이 없을 경우 0으로 값 초기화
				if !diskMetricAdded && (idx == len(diskRow.Values)-1) {
					initVal := make([]interface{}, len(val)-1)
					for i := range initVal {
						initVal[i] = 0
					}
					metricVal = append(metricVal, initVal...)
				}
			}

			// diskio 메트릭 aggregate
			diskIoMetricAdded := false
			for idx, val := range diskIoRow.Values {
				t := val[0].(string)
				if strings.EqualFold(t, tp) {
					metricVal = append(metricVal, val[1:]...)
					diskIoMetricAdded = true
					break
				}
				// 해당 TimePoint에 해당하는 disk 메트릭이 없을 경우 0으로 값 초기화
				if !diskIoMetricAdded && (idx == len(diskIoRow.Values)-1) {
					initVal := make([]interface{}, len(val)-1)
					for i := range initVal {
						initVal[i] = 0
					}
					metricVal = append(metricVal, initVal...)
				}
			}

			resultRow.Values = append(resultRow.Values, metricVal)
		}

		return c.JSON(http.StatusOK, resultRow)

	case "network":

		// network 메트릭 조회
		metricKey := "net"
		netMetric, err := apiServer.InfluxDB.ReadMetric(vmId, metricKey, period, aggregateType, duration)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		if netMetric == nil {
			return c.JSON(http.StatusNotFound, err)
		}

		resultMetric, err := metricstore.MappingMonMetric(metricKey, &netMetric)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		return c.JSON(http.StatusOK, resultMetric)

	default:
		return c.JSON(http.StatusNotFound, fmt.Sprintf("not found metric : %s", metricName))
	}
}

// 멀티 클라우드 인프라 VM 실시간 모니터링 정보 조회
func (apiServer *APIServer) GetVMRealtimeMonInfo(c echo.Context) error {

	// Path 파라미터 가져오기
	//mcisId := c.Param("mcis_id")
	vmId := c.Param("vm_id")
	metricName := c.Param("metric_name")

	// Query 파라미터 가져오기
	aggregateType := c.QueryParam("statisticsCriteria")

	apiServer.aggregator.Etcd = apiServer.Etcd

	resultMap := map[string]interface{}{}
	resultMap["vmId"] = vmId
	resultMap["metricName"] = metricName
	resultMap["time"] = time.Now().UTC()
	resultMap["value"] = map[string]interface{}{}

	var metricMap map[string]interface{}

	switch metricName {
	case "cpu":
		// cpu 메트릭 조회
		metricKey := "cpu"
		result, err := apiServer.aggregator.GetAggregateMetric(vmId, metricKey, aggregateType)
		if err != nil {
			// 만약 실시간 데이터가 없을 경우 empty Map 값 전달
			if err.(client.Error).Code == 100 {
				return c.JSON(http.StatusOK, resultMap)
			}
		}
		// cpu 메트릭 매핑
		metricMap, err = realtimestore.MappingMonMetric(metricKey, result)
		if _, ok := err.(client.Error); ok {
			return c.JSON(http.StatusInternalServerError, err)
		}
		resultMap["value"] = metricMap

	case "memory":
		// memory 메트릭 조회
		metricKey := "mem"
		result, err := apiServer.aggregator.GetAggregateMetric(vmId, metricKey, aggregateType)
		if err != nil {
			// 만약 실시간 데이터가 없을 경우 empty Map 값 전달
			if err.(client.Error).Code == 100 {
				return c.JSON(http.StatusOK, resultMap)
			}
		}
		// memory 메트릭 매핑
		metricMap, err = realtimestore.MappingMonMetric(metricKey, result)
		if _, ok := err.(client.Error); ok {
			return c.JSON(http.StatusInternalServerError, err)
		}
		resultMap["value"] = metricMap

	case "disk":
		// disk 메트릭 조회
		metricKey := "disk"
		diskMetric, err := apiServer.aggregator.GetAggregateDiskMetric(vmId, metricKey, aggregateType)
		if err != nil {
			// 만약 실시간 데이터가 없을 경우 empty Map 값 전달
			if err.(client.Error).Code == 100 {
				return c.JSON(http.StatusOK, resultMap)
			}
			return c.JSON(http.StatusInternalServerError, err)
		}
		// disk 메트릭 매핑
		diskMetricMap, err := realtimestore.MappingMonMetric(metricKey, diskMetric)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		// diskio 메트릭 조회
		metricKey = "diskio"
		diskIoMetric, err := apiServer.aggregator.GetAggregateDiskMetric(vmId, metricKey, aggregateType)
		if err != nil {
			// 만약 실시간 데이터가 없을 경우 empty Map 값 전달
			if err.(client.Error).Code == 100 {
				return c.JSON(http.StatusOK, resultMap)
			}
		}
		// diskio 메트릭 매핑
		diskIoMetricMap, err := realtimestore.MappingMonMetric(metricKey, diskIoMetric)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		// disk, diskio 메트릭 통합
		metricMap := map[string]interface{}{}
		for metricKey, val := range diskMetricMap {
			metricMap[metricKey] = val
		}
		for metricKey, val := range diskIoMetricMap {
			metricMap[metricKey] = val
		}
		resultMap["value"] = metricMap
	case "network":
		// network 메트릭 조회
		metricKey := "net"
		result, err := apiServer.aggregator.GetAggregateMetric(vmId, metricKey, aggregateType)
		if err != nil {
			// 만약 실시간 데이터가 없을 경우 empty Map 값 전달
			if err.(client.Error).Code == 100 {
				return c.JSON(http.StatusOK, resultMap)
			}
		}
		// network 메트릭 매핑
		metricMap, err = realtimestore.MappingMonMetric(metricKey, result)
		if _, ok := err.(client.Error); ok {
			return c.JSON(http.StatusInternalServerError, err)
		}
		resultMap["value"] = metricMap
	default:
		return c.JSON(http.StatusNotFound, fmt.Sprintf("not found metric : %s", metricName))
	}

	return c.JSON(http.StatusOK, resultMap)
}

// 모니터링 정책 설정
func (apiServer *APIServer) SetMonConfig(c echo.Context) error {

	// form 파라미터 정보 가져오기
	agentInterval, err := strconv.Atoi(c.FormValue("agent_interval"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	collectorInterval, err := strconv.Atoi(c.FormValue("collector_interval"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	schedulingInterval, err := strconv.Atoi(c.FormValue("schedule_interval"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	maxHostCnt, err := strconv.Atoi(c.FormValue("max_host_count"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// etcd 저장소에 모니터링 정책 정보 저장
	monConfig := MonConfig{
		AgentInterval:      agentInterval,
		CollectorInterval:  collectorInterval,
		SchedulingInterval: schedulingInterval,
		MaxHostCount:       maxHostCnt,
	}

	// TODO: 구조체 map[string]interface{} 타입으로 Unmarshal
	// TODO: 추후에 별도의 map 변환 함수 (toMap() 개발)
	reqBodyBytes := new(bytes.Buffer)
	if err = json.NewEncoder(reqBodyBytes).Encode(monConfig); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	byteData := reqBodyBytes.Bytes()

	jsonMap := map[string]interface{}{}
	if err = json.Unmarshal(byteData, &jsonMap); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// etcd 저장소에 모니터링 정책 저장 후 결과 값 반환
	err = apiServer.Etcd.WriteMetric("/mon/config", jsonMap)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, monConfig)
}

// 모니터링 정책 조회
func (apiServer *APIServer) GetMonConfig(c echo.Context) error {

	// etcd 저장소에서 모니터링 정책 정보 조회
	node, err := apiServer.Etcd.ReadMetric("/mon/config")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// 모니터링 정책 정보 구조체 매핑
	monConfig := &MonConfig{}
	if err := json.Unmarshal([]byte(node.Value), monConfig); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, monConfig)
}

// 모니터링 정책 초기화
func (apiServer *APIServer) ResetMonConfig(c echo.Context) error {

	// config 파일 기준 모니터링 정책 초기화
	monConfig := MonConfig{
		AgentInterval:      apiServer.config.Monitoring.AgentInterval,
		CollectorInterval:  apiServer.config.Monitoring.CollectorInterval,
		SchedulingInterval: apiServer.config.Monitoring.ScheduleInterval,
		MaxHostCount:       apiServer.config.Monitoring.MaxHostCount,
	}

	// TODO: 구조체 map[string]interface{} 타입으로 Unmarshal
	// TODO: 추후에 별도의 map 변환 함수 (toMap() 개발)
	reqBodyBytes := new(bytes.Buffer)
	if err := json.NewEncoder(reqBodyBytes).Encode(monConfig); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	byteData := reqBodyBytes.Bytes()

	jsonMap := map[string]interface{}{}
	if err := json.Unmarshal(byteData, &jsonMap); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// etcd 저장소에 모니터링 정책 저장 후 결과 값 반환
	if err := apiServer.Etcd.WriteMetric("/mon/config", jsonMap); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, monConfig)
}

// Telegraf agent 설치 스크립트 파일 다운로드
func (apiServer *APIServer) GetTelegrafInstallScript(c echo.Context) error {
	// Query 파라미터 가져오기
	mcisId := c.QueryParam("mcis_id")
	vmId := c.QueryParam("vm_id")

	// Query 파라미터 값 체크
	if mcisId == "" || vmId == "" {
		err := errors.New("failed to get package. query parameter is missing")
		return c.JSON(http.StatusInternalServerError, err)
	}

	collectorServer := fmt.Sprintf("%s:%d", apiServer.config.CollectManager.CollectorIP, apiServer.config.APIServer.Port)

	rootPath := os.Getenv("CBMON_PATH")
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
}

// Telegraf config 파일 다운로드
func (apiServer *APIServer) GetTelegrafConfFile(c echo.Context) error {
	// Query 파라미터 가져오기
	mcisId := c.QueryParam("mcis_id")
	vmId := c.QueryParam("vm_id")

	// Query 파라미터 값 체크
	if mcisId == "" || vmId == "" {
		err := errors.New("failed to get package. query parameter is missing")
		return c.JSON(http.StatusInternalServerError, err)
	}

	collectorServer := fmt.Sprintf("udp://%s:%d", apiServer.config.CollectManager.CollectorIP, apiServer.config.CollectManager.CollectorPort)

	rootPath := os.Getenv("CBMON_PATH")
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
	strConf = strings.ReplaceAll(strConf, "{{influxdb_server}}", apiServer.config.InfluxDB.EndpointUrl)

	return c.Blob(http.StatusOK, "text/plain", []byte(strConf))
}

// Telegraf package 파일 다운로드
func (apiServer *APIServer) GetTelegrafPkgFile(c echo.Context) error {
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

	rootPath := os.Getenv("CBMON_PATH")
	var filePath string
	switch osType {
	case "ubuntu":
		filePath = rootPath + fmt.Sprintf("/file/pkg/%s/%s/telegraf_1.12.0~f09f2b5-0_amd64.deb", osType, arch)
	case "centos":
		filePath = rootPath + fmt.Sprintf("/file/pkg/%s/%s/telegraf-1.12.0~f09f2b5-0.x86_64.rpm", osType, arch)
	default:
		err := errors.New(fmt.Sprintf("failed to get package. osType %s not supported", osType))
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.File(filePath)
}

func (apiServer *APIServer) createTelegrafConfigFile(mcisId string, vmId string) (string, error) {

	collectorServer := fmt.Sprintf("udp://%s:%d", apiServer.config.CollectManager.CollectorIP, apiServer.config.CollectManager.CollectorPort)

	rootPath := os.Getenv("CBMON_PATH")
	filePath := rootPath + "/file/conf/telegraf.conf"

	read, err := ioutil.ReadFile(filePath)
	if err != nil {
		// ERROR 정보 출럭
		logrus.Error("failed to read telegraf.conf file.")
		return "", err
	}

	// 파일 내의 변수 값 설정 (hostId, collectorServer)
	strConf := string(read)
	strConf = strings.ReplaceAll(strConf, "{{mcis_id}}", mcisId)
	strConf = strings.ReplaceAll(strConf, "{{vm_id}}", vmId)
	strConf = strings.ReplaceAll(strConf, "{{collector_server}}", collectorServer)
	strConf = strings.ReplaceAll(strConf, "{{influxdb_server}}", apiServer.config.InfluxDB.EndpointUrl)

	// telegraf.conf 파일 생성
	telegrafFilePath := rootPath + "/file/conf/"
	createFileName := "telegraf-" + uuid.New().String() + ".conf"
	telegrafConfFile := telegrafFilePath + createFileName

	err = ioutil.WriteFile(telegrafConfFile, []byte(strConf), os.FileMode(777))
	if err != nil {
		logrus.Error("failed to create telegraf.conf file.")
		return "", err
	}

	return telegrafConfFile, err
}

func (apiServer *APIServer) InstallTelegraf(c echo.Context) error {
	// form 파라미터 값 가져오기
	mcisId := c.FormValue("mcis_id")
	vmId := c.FormValue("vm_id")
	publicIp := c.FormValue("public_ip")
	userName := c.FormValue("user_name")
	sshKey := c.FormValue("ssh_key")

	// form 파라미터 값 체크
	if mcisId == "" || vmId == "" || publicIp == "" || userName == "" || sshKey == "" {
		err := errors.New("failed to get package. query parameter is missing")
		return c.JSON(http.StatusInternalServerError, err)
	}

	sshInfo := sshrun.SSHInfo{
		ServerPort: publicIp + ":22", //serverEndPoint
		UserName:   userName,         //userName
		PrivateKey: []byte(sshKey),   //[]byte(privateKey)
	}

	// {사용자계정}/cb-dragonfly 폴더 생성
	createFolderCmd := fmt.Sprintf("mkdir $HOME/cb-dragonfly")
	if _, err := sshrun.SSHRun(sshInfo, createFolderCmd); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// 리눅스 OS 환경 체크
	osType, err := sshrun.SSHRun(sshInfo, "hostnamectl | grep 'Operating System' | awk '{print $3}' | tr 'a-z' 'A-Z'")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	rootPath := os.Getenv("CBMON_PATH")

	var sourceFile, targetFile, installCmd string
	if strings.Contains(osType, "CENTOS") {
		sourceFile = rootPath + "/file/pkg/centos/x64/telegraf-1.12.0~f09f2b5-0.x86_64.rpm"
		targetFile = fmt.Sprintf("$HOME/cb-dragonfly/cb-agent.rpm")
		installCmd = fmt.Sprintf("sudo rpm -ivh $HOME/cb-dragonfly/cb-agent.rpm")
	} else if strings.Contains(osType, "UBUNTU") {
		sourceFile = rootPath + "/file/pkg/ubuntu/x64/telegraf_1.12.0~f09f2b5-0_amd64.deb"
		targetFile = fmt.Sprintf("$HOME/cb-dragonfly/cb-agent.deb")
		installCmd = fmt.Sprintf("sudo dpkg -i $HOME/cb-dragonfly/cb-agent.deb")
	}

	// 에이전트 설치 패키지 다운로드
	if err := sshrun.SSHCopy(sshInfo, sourceFile, targetFile); err != nil {
		return err
	}
	// 패키지 설치 실행
	if _, err := sshrun.SSHRun(sshInfo, installCmd); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// 공통 서비스 활성화 및 실행

	// 설치시 자동 생성되는 telegraf_conf 파일 제거
	if _, err := sshrun.SSHRun(sshInfo, "sudo rm /etc/telegraf/telegraf.conf"); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// telegraf_conf 파일 복사
	telegrafConfSourceFile, err := apiServer.createTelegrafConfigFile(mcisId, vmId)
	telegrafConfTargetFile := "$HOME/cb-dragonfly/telegraf.conf"
	if err != nil {
		return err
	}
	if err := sshrun.SSHCopy(sshInfo, telegrafConfSourceFile, telegrafConfTargetFile); err != nil {
		return err
	}

	// telegraf_conf 파일 이동
	if _, err := sshrun.SSHRun(sshInfo, "sudo chown root:root $HOME/cb-dragonfly/telegraf.conf"); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	if _, err := sshrun.SSHRun(sshInfo, "sudo mv $HOME/cb-dragonfly/telegraf.conf /etc/telegraf/"); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// telegraf UUId conf 파일 삭제
	err = os.Remove(telegrafConfSourceFile)
	if err != nil {
		logrus.Error("failed to remove telegraf_UUID_conf File")
	}

	// 에이전트 설치에 사용한 파일 폴더 채로 제거
	removeRpmCmd := fmt.Sprintf("sudo rm -rf $HOME/cb-dragonfly")
	if _, err := sshrun.SSHRun(sshInfo, removeRpmCmd); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// 설치 스크립트 다운로드
	/*apiEndpoint := fmt.Sprintf("%s:%d", apiServer.config.CollectManager.CollectorIP, apiServer.config.APIServer.Port)
	downloadCmd := fmt.Sprintf("wget -O agent_install.sh \"http://%s/mon/file/agent/install?mcis_id=%s&vm_id=%s\"", apiEndpoint, mcisId, vmId)
	if _, err := util.RunCommand(publicIp, userName, sshKey, downloadCmd); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}*/

	// 설치 스크립트 실행 권한 추가
	/*chmodCmd := fmt.Sprintf("chmod +x agent_install.sh")
	if _, err := util.RunCommand(publicIp, userName, sshKey, chmodCmd); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}*/

	// 설치 스크립트 실행
	/*execCmd := fmt.Sprintf("bash agent_install.sh")
	if _, err := util.RunCommand(publicIp, userName, sshKey, execCmd); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}*/

	// 정상 설치 확인
	checkCmd := "telegraf --version"
	if result, err := util.RunCommand(publicIp, userName, sshKey, checkCmd); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	} else {
		if strings.Contains(*result, "command not found") {
			err = errors.New("failed to install agent")
			return c.JSON(http.StatusInternalServerError, err)
		}

		response := echo.Map{}
		response["message"] = "agent installation is finished"
		return c.JSON(http.StatusOK, response)
	}
}
