package collector

import (
	"sync"
)

// 전체 호스트 목록 로컬 변수
// 고루틴에서 호스트 목록 조회 시 Mutex 처리를 위해 사용
type HostInfo struct {
	HostMap *map[string]string
	L       *sync.RWMutex
}

func (h HostInfo) GetHostById(hostId string) string {
	h.L.RLock()
	defer h.L.RUnlock()
	return (*h.HostMap)[hostId]
}

func (h HostInfo) AddHost(hostId string) {
	h.L.Lock()
	defer h.L.Unlock()
	(*h.HostMap)[hostId] = hostId
}

func (h HostInfo) DeleteHost(hostArr []string) {
	h.L.Lock()
	defer h.L.Unlock()
	for _, hostId := range hostArr {
		delete(*h.HostMap, hostId)
	}
}

func (h HostInfo) Clear() {
	h.L.Lock()
	defer h.L.Unlock()

	*h.HostMap = make(map[string]string)
}
