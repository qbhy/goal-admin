package goal_admin

import (
	"sync"
)

// factory 资源工厂，管理资源注册与查找
type factory struct {
	mu        sync.RWMutex
	resources map[string]Resource
}

func New() Factory {
	return &factory{resources: map[string]Resource{}}
}

// Register 注册一个资源（model 唯一标识）
func (f *factory) Register(r Resource) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.resources == nil {
		f.resources = make(map[string]Resource)
	}
	f.resources[r.GetName()] = r
}

// Get 获取资源实例
func (f *factory) Get(model string) (Resource, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	r, ok := f.resources[model]
	return r, ok
}
