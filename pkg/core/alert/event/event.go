package event

import (
	"encoding/json"

	cbstore "github.com/cloud-barista/cb-store"
	"github.com/cloud-barista/cb-store/config"
	icbs "github.com/cloud-barista/cb-store/interfaces"
	"github.com/sirupsen/logrus"

	"github.com/cloud-barista/cb-dragonfly/pkg/core/alert/types"
)

var cblog *logrus.Logger
var store icbs.Store

func init() {
	cblog = config.Cblogger
	store = cbstore.GetStore()
}

func CreateEventLog(eventLog types.AlertEventLog) error {
	var eventLogArr []types.AlertEventLog

	eventLogStr, err := store.Get(eventLog.Id)
	if err != nil {
		return err
	}

	if eventLogStr != nil {
		// Get event log array
		err := json.Unmarshal([]byte(eventLogStr.Value), &eventLogArr)
		if err != nil {
			return err
		}
		// Add new event log
		eventLogArr = append(eventLogArr, eventLog)
	}

	// Save event log
	newEventLogBytes, err := json.Marshal(eventLogArr)
	if err != nil {
		return err
	}
	err = store.Put(eventLog.Id, string(newEventLogBytes))
	if err != nil {
		return err
	}
	return nil
}

func ListEventLog(alertName string) ([]types.AlertEventLog, error) {
	var eventLogArr []types.AlertEventLog
	eventLogStr, err := store.Get(alertName)
	if err != nil {
		return nil, err
	}
	if eventLogStr == nil {
		return nil, nil
	}
	err = json.Unmarshal([]byte(eventLogStr.Value), &eventLogArr)
	if err != nil {
		return nil, err
	}
	return eventLogArr, nil
}
