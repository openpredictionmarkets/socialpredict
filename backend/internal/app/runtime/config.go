package runtime

import (
	configsvc "socialpredict/internal/service/config"
)

// LoadConfigService initializes runtime configuration from an explicit source asset.
func LoadConfigService(source configsvc.Source) (configsvc.Service, error) {
	return loadConfigService(configsvc.NewYAMLLoader(source))
}

func loadConfigService(loader configsvc.Loader) (configsvc.Service, error) {
	service, err := configsvc.NewService(loader)
	if err != nil {
		return nil, err
	}
	return service, nil
}
