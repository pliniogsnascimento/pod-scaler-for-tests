package scales

import (
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// Implements Scaler Interface
type VanillaHpa struct {
	clientset    kubernetes.Interface
	scaleConfigs ScaleConfigs
	logger       *logrus.Logger
	sleep        *time.Duration
	k8sHelper    k8sHelperInterface
}

func NewVanillaHpa(clientset kubernetes.Interface, scaleConfigs ScaleConfigs, logger *logrus.Logger, sleep *time.Duration) *VanillaHpa {
	return &VanillaHpa{
		clientset:    clientset,
		scaleConfigs: scaleConfigs,
		logger:       logger,
		sleep:        sleep,
		k8sHelper:    newk8sHelper(clientset),
	}
}

func (hpa *VanillaHpa) Scale() error {
	for _, config := range hpa.scaleConfigs {
		hpaConfig, err := hpa.k8sHelper.getHpaWithTimeout(config.Name, 500)

		if err != nil {
			return err
		}

		minReplicas := int32(config.Min)
		hpaConfig.Spec.MinReplicas = &minReplicas
		hpaConfig.Spec.MaxReplicas = int32(config.Max)

		return hpa.k8sHelper.updateHpaWithTimeout(config.Name, hpaConfig, 500)
	}
	return nil
}
