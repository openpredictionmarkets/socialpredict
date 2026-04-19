package runtime

import (
	configsvc "socialpredict/internal/service/config"
	"socialpredict/setup"
)

// LoadConfigService initializes runtime configuration from the embedded setup source.
func LoadConfigService() (configsvc.Service, error) {
	return configsvc.NewService(configsvc.LoaderFunc(setup.LoadEconomicsConfig))
}
