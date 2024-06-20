package resources

import (
	"github.com/goal-web/contracts"
)

type ServiceProvider struct {
	list []Resource
}

func NewService(list ...Resource) contracts.ServiceProvider {
	return ServiceProvider{list: list}
}

func (s ServiceProvider) Register(application contracts.Application) {
	application.Singleton("resources", func(connection contracts.DBConnection, fs contracts.FileSystemFactory) Factory {
		factory := NewFactory(connection, fs.Disk("resources"))

		for _, res := range s.list {
			factory.Extend(res)
		}

		return factory
	})
	application.Call(func(console contracts.Console) {
		console.RegisterCommand("scan:resources", ScanResources)
	})
}

func (s ServiceProvider) Start() error {
	return nil
}

func (s ServiceProvider) Stop() {
}
