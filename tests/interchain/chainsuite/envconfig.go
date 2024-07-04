package chainsuite

import (
	"github.com/kelseyhightower/envconfig"
)

const prefix = "TEST"

type Environment struct {
	DockerRegistry      string `envconfig:"DOCKER_REGISTRY"`
	GaiaImageName       string `envconfig:"GAIA_IMAGE_NAME" default:"gaia"`
	OldGaiaImageVersion string `envconfig:"OLD_GAIA_IMAGE_VERSION"`
	NewGaiaImageVersion string `envconfig:"NEW_GAIA_IMAGE_VERSION"`
	UpgradeName         string `envconfig:"UPGRADE_NAME"`
}

func GetEnvironment() Environment {
	var env Environment
	envconfig.MustProcess(prefix, &env)
	return env
}
