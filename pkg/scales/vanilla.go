package scales

import (
	"time"

	"github.com/sirupsen/logrus"
)

// Implements Scaler Interface
type vanillaHpa struct {
	logger    *logrus.Logger
	k8sHelper k8sHelperInterface
}

func newVanillaHpa(k8sHelper k8sHelperInterface, logger *logrus.Logger) *vanillaHpa {
	return &vanillaHpa{
		logger:    logger,
		k8sHelper: k8sHelper,
	}
}

func (hpa *vanillaHpa) Scale(config ScaleConfig) error {
	helper := hpa.k8sHelper
	hpaConfig, err := helper.getHpaWithTimeout(config.Name, 500*time.Millisecond)

	if err != nil {
		return err
	}

	minReplicas := int32(config.Min)
	hpaConfig.Spec.MinReplicas = &minReplicas
	hpaConfig.Spec.MaxReplicas = int32(config.Max)

	return helper.updateHpaWithTimeout(config.Name, hpaConfig, 500)
}
