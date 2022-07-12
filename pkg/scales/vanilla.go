package scales

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type VanillaHpa struct {
	clientset    kubernetes.Interface
	scaleConfigs ScaleConfigs
	logger       *logrus.Logger
	sleep        *time.Duration
}

func NewVanillaHpa(clientset kubernetes.Interface, scaleConfigs ScaleConfigs, logger *logrus.Logger, sleep *time.Duration) *VanillaHpa {
	return &VanillaHpa{
		clientset:    clientset,
		scaleConfigs: scaleConfigs,
		logger:       logger,
		sleep:        sleep,
	}
}

func (hpa *VanillaHpa) Scale() error {
	for _, config := range hpa.scaleConfigs {
		hpaConfig, err := hpa.clientset.AutoscalingV1().HorizontalPodAutoscalers(config.Name).Get(context.TODO(), config.Name, metav1.GetOptions{})
		if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			hpa.logger.Errorln(err.Error())
			return err
		}

		if errors.IsNotFound(err) {
			hpa.logger.Warnf("HPA not found in namespace %s\n", config.Name)
			return nil
		}

		minReplicas := int32(config.Min)
		hpaConfig.Spec.MinReplicas = &minReplicas
		hpaConfig.Spec.MaxReplicas = int32(config.Max)

		_, err = hpa.clientset.AutoscalingV1().HorizontalPodAutoscalers(config.Name).Update(context.TODO(), hpaConfig, metav1.UpdateOptions{})
		if err == nil {
			hpa.logger.Printf("Success updating HPA %s!", hpaConfig.Name)
		}
		if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			hpa.logger.Errorln(err.Error())
			return err
		}
	}
	return nil
}

func (hpa VanillaHpa) ScaleWithConcurrency() error {
	return fmt.Errorf("Not implemented")
}
