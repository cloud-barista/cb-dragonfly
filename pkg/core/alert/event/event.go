package event

import (
	"encoding/json"
	"strings"

	"github.com/cloud-barista/cb-dragonfly/pkg/core/alert/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/localstore"
)

func CreateEventLog(eventLog types.AlertEventLog) error {
	var eventLogArr []types.AlertEventLog

	eventLogStr := localstore.GetInstance().StoreGet(eventLog.Id)

	if eventLogStr != "" {
		// Get event log array
		err := json.Unmarshal([]byte(eventLogStr), &eventLogArr)
		if err != nil {
			return err
		}
	}

	// Add new event log
	eventLogArr = append(eventLogArr, eventLog)

	// Save event log
	newEventLogBytes, err := json.Marshal(eventLogArr)
	if err != nil {
		return err
	}
	err = localstore.GetInstance().StorePut(eventLog.Id, string(newEventLogBytes))
	if err != nil {
		return err
	}
	return nil
}

func ListEventLog(taskId string, logLevel string) ([]types.AlertEventLog, error) {
	var eventLogArr []types.AlertEventLog
	eventLogStr := localstore.GetInstance().StoreGet(taskId)
	if eventLogStr == "" {
		return []types.AlertEventLog{}, nil
	}
	err := json.Unmarshal([]byte(eventLogStr), &eventLogArr)
	if err != nil {
		return nil, err
	}
	if logLevel == "" {
		return eventLogArr, nil
	}

	filterdEventLogArr := []types.AlertEventLog{}
	for _, log := range eventLogArr {
		if strings.EqualFold(log.Level, logLevel) {
			filterdEventLogArr = append(filterdEventLogArr, log)
		}
	}
	return filterdEventLogArr, nil
}

func DeleteEventLog(taskId string) error {
	return localstore.GetInstance().StoreDelete(taskId)
}
