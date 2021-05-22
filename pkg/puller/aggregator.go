package puller

import (
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
)

type PullAggregator struct {
	AggregateType types.AggregateType
}

func (pa *PullAggregator) AggregateMetric() (map[string]interface{}, error) {
	return nil, nil
}

func (pa *PullAggregator) CalculateMetric() (map[string]interface{}, error) {
	return nil, nil
}
