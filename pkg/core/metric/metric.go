package metric

import (
	"errors"
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/collector"
	"github.com/cloud-barista/cb-dragonfly/pkg/core"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdbv1"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore"
	"github.com/influxdata/influxdb/models"
	"go.etcd.io/etcd/client"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Metric string

const (
	Cpu         = "cpu"
	CpuFreqency = "cpufreq"
	Memory      = "memory"
	Disk        = "disk"
	DiskIO      = "diskio"
	Network     = "network"
)

// 가상머신 모니터링 메트릭 조회
func GetVMMonInfo(nsId string, mcisId string, vmId string, metricName string, period string, aggregateType string, duration string) (interface{}, int, error) {

	switch metricName {

	case Cpu:

		// cpu 메트릭 조회
		cpuMetric, err := core.CoreConfig.InfluxDB.ReadMetric(vmId, Cpu, period, aggregateType, duration)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		if cpuMetric == nil {
			return nil, http.StatusNotFound, nil
		}
		resultMetric, err := metricstore.MappingMonMetric(Cpu, &cpuMetric)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return resultMetric, http.StatusOK, nil

	case CpuFreqency:

		// cpufreq 메트릭 조회
		cpuFreqMetric, err := core.CoreConfig.InfluxDB.ReadMetric(vmId, CpuFreqency, period, aggregateType, duration)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		if cpuFreqMetric == nil {
			return nil, http.StatusNotFound, nil
		}
		resultMetric, err := metricstore.MappingMonMetric(CpuFreqency, &cpuFreqMetric)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return resultMetric, http.StatusOK, nil

	case Memory:

		// memory 메트릭 조회
		memMetric, err := core.CoreConfig.InfluxDB.ReadMetric(vmId, "mem", period, aggregateType, duration)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		if memMetric == nil {
			return nil, http.StatusNotFound, nil
		}
		resultMetric, err := metricstore.MappingMonMetric(Memory, &memMetric)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return resultMetric, http.StatusOK, nil

	case Disk:

		// disk, diskio 메트릭 조회
		diskMetric, err := influxdbv1.GetInstance().ReadMetric(vmId, Disk, period, aggregateType, duration)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		diskIoMetric, err := influxdbv1.GetInstance().ReadMetric(vmId, DiskIO, period, aggregateType, duration)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		if diskMetric == nil || diskIoMetric == nil {
			return nil, http.StatusNotFound, nil
		}

		diskRow := diskMetric.(models.Row)
		diskIoRow := diskIoMetric.(models.Row)

		// Aggregate Metric
		var resultRow models.Row
		resultRow.Name = Disk
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
			if _, exist := timePointMap[timePoint]; !exist {
				timePointMap[timePoint] = timePoint
			}
		}

		// TimePoint 배열 생성
		idx := 0
		timePointArr := make([]string, len(timePointMap))
		for _, timePoint := range timePointMap {
			timePointArr[idx] = timePoint
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

		return resultRow, http.StatusOK, nil

	case Network:

		// network 메트릭 조회
		netMetric, err := core.CoreConfig.InfluxDB.ReadMetric(vmId, "net", period, aggregateType, duration)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		if netMetric == nil {
			return nil, http.StatusNotFound, nil
		}
		resultMetric, err := metricstore.MappingMonMetric(Network, &netMetric)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return resultMetric, http.StatusOK, nil

	default:
		return nil, http.StatusInternalServerError, errors.New(fmt.Sprintf("NOT FOUND METRIC : %s", metricName))
	}

	return nil, http.StatusInternalServerError, errors.New(fmt.Sprintf("NOT FOUND METRIC : %s", metricName))
}

func GetVMLatestMonInfo(nsId string, mcisId string, vmId string, metricName string, statisticsCriteria string) (interface{}, int, error) {
	resultMap := map[string]interface{}{}
	resultMap["vmId"] = vmId
	resultMap["metricName"] = metricName
	resultMap["time"] = time.Now().UTC()
	resultMap["value"] = map[string]interface{}{}

	var metricKey string
	var metricMap map[string]interface{}
	var diskMetric, diskIoMetric, result map[string]interface{}
	var diskMetricMap, diskIoMetricMap map[string]interface{}
	var err error
	var val interface{}

	aggregator := collector.Aggregator{}

	if metricName == "disk" || metricName == "diskio" {
		metricKey = "disk"
		diskMetric, err = aggregator.GetAggregateDiskMetric(vmId, metricKey, statisticsCriteria)
		if err != nil {
			// 만약 실시간 데이터가 없을 경우 empty Map 값 전달
			if err.(client.Error).Code == 100 {
				return resultMap, http.StatusOK, nil
			}
			return nil, http.StatusInternalServerError, err
		}
		// disk 메트릭 매핑
		diskMetricMap, err = realtimestore.MappingMonMetric(metricKey, diskMetric)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		// diskio 메트릭 조회
		metricKey = "diskio"
		diskIoMetric, err = aggregator.GetAggregateDiskMetric(vmId, metricKey, statisticsCriteria)
		if err != nil {
			// 만약 실시간 데이터가 없을 경우 empty Map 값 전달
			if err.(client.Error).Code == 100 {
				return resultMap, http.StatusOK, nil
			}
		}
		// diskio 메트릭 매핑
		diskIoMetricMap, err = realtimestore.MappingMonMetric(metricKey, diskIoMetric)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	} else {

		//메트릭 키 설정
		switch metricName {
		case Cpu:
			metricKey = "cpu"
		case CpuFreqency:
			metricKey = "cpufreq"
		case Memory:
			metricKey = "mem"
		case Network:
			metricKey = "net"
		default:
			return nil, http.StatusNotFound, errors.New(fmt.Sprintf("not found metric : %s", metricName))
		}
		// disk, diskio 제외한 메트릭 조회
		result, err = aggregator.GetAggregateMetric(vmId, metricKey, statisticsCriteria)
		if err != nil {
			// 만약 실시간 데이터가 없을 경우 empty Map 값 전달
			if _, ok := err.(client.Error); !ok {
				return nil, http.StatusInternalServerError, errors.New("")
			}
			if err.(client.Error).Code == 100 {
				return resultMap, http.StatusOK, nil
			}
		}
		// disk, diskio 제외한 메트릭 매핑
		metricMap, err = realtimestore.MappingMonMetric(metricKey, result)
		if _, ok := err.(client.Error); ok {
			return nil, http.StatusInternalServerError, err
		}
	}

	for metricKey, val = range metricMap {
		metricMap[metricKey] = val
	}
	for metricKey, val = range diskMetricMap {
		metricMap[metricKey] = val
	}
	for metricKey, val = range diskIoMetricMap {
		metricMap[metricKey] = val
	}
	resultMap["value"] = metricMap

	return resultMap, http.StatusOK, nil
}
