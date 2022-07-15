package scales

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// Implements Scaler Interface
type VanillaHpa struct {
	clientset kubernetes.Interface
	logger    *logrus.Logger
	k8sHelper k8sHelperInterface
}

func NewVanillaHpa(clientset kubernetes.Interface, logger *logrus.Logger) *VanillaHpa {
	return &VanillaHpa{
		clientset: clientset,
		logger:    logger,
		k8sHelper: newk8sHelper(clientset),
	}
}

func (hpa *VanillaHpa) Scale(config ScaleConfig) error {
	helper := hpa.k8sHelper
	hpaConfig, err := helper.getHpaWithTimeout(config.Name, 500)

	if err != nil {
		return err
	}

	minReplicas := int32(config.Min)
	hpaConfig.Spec.MinReplicas = &minReplicas
	hpaConfig.Spec.MaxReplicas = int32(config.Max)

	return helper.updateHpaWithTimeout(config.Name, hpaConfig, 500)
}
