package util

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Workiva/go-datastructures/queue"
)

type TopicStructure struct {
	Policy string
	Topic  string
}

// MCIS 큐
var ringQueueOnce sync.Once
var ringQueue *queue.Queue

func GetRingQueue() *queue.Queue {
	ringQueueOnce.Do(func() {
		ringQueue = queue.New(1000)
	})
	return ringQueue
}

// MCKS 큐

var mcksRingQueueOnce sync.Once
var mcksRingQueue *queue.Queue

func GetMCKSRingQueue() *queue.Queue {
	mcksRingQueueOnce.Do(func() {
		mcksRingQueue = queue.New(1000)
	})
	return mcksRingQueue
}

func RingQueuePut(topicManagePolicy string, topic string) error {
	var topicBytes []byte
	var err error
	topicStructure := TopicStructure{
		Policy: topicManagePolicy,
		Topic:  topic,
	}
	if topicBytes, err = json.Marshal(topicStructure); err != nil {
		fmt.Println("error?")
		return err
	}
	if err = GetRingQueue().Put(topicBytes); err != nil {
		return err
	}
	return nil
}

func PutMCKSRingQueue(topicManagePolicy string, topic string) error {
	var topicBytes []byte
	var err error
	topicStructure := TopicStructure{
		Policy: topicManagePolicy,
		Topic:  topic,
	}
	if topicBytes, err = json.Marshal(topicStructure); err != nil {
		fmt.Println("error?")
		return err
	}
	if err = GetMCKSRingQueue().Put(topicBytes); err != nil {
		return err
	}
	return nil
}
