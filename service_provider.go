package goal_admin

import (
	"sync"

	"github.com/goal-web/contracts"
	"github.com/qbhy/goal-admin/resources"
)

var (
	// defaultFactory 全局单例工厂
	defaultFactory Factory
	once           sync.Once
)

// Default 获取全局工厂
func Default() Factory {
	once.Do(func() {
		defaultFactory = New()

		// 注册默认提供的资源
		defaultFactory.Register(resources.NewExportResource())
	})
	return defaultFactory
}

func NewService() contracts.ServiceProvider {
	return &ServiceProvider{}
}

type ServiceProvider struct {
	app contracts.Application
}

func (s ServiceProvider) Register(application contracts.Application) {
	application.Singleton("resources", func() Factory {
		return Default()
	})
	s.app = application
}

func (s ServiceProvider) Start() error {
	return nil
}

func (s ServiceProvider) Stop() {
}
