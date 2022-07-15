package scales

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Implements Scaler Interface
type HpaOperator struct {
	clientset    kubernetes.Interface
	scaleConfigs ScaleConfigs
	logger       *logrus.Logger
	sleep        *time.Duration
}

func NewHpaOperator(clientset kubernetes.Interface, scaleConfigs ScaleConfigs, logger *logrus.Logger, sleep *time.Duration) *HpaOperator {
	return &HpaOperator{
		clientset:    clientset,
		scaleConfigs: scaleConfigs,
		logger:       logger,
		sleep:        sleep,
	}
}

func (op HpaOperator) Scale() error {
	for _, config := range op.scaleConfigs {
		deploy, err := op.clientset.AppsV1().Deployments(config.Name).Get(context.TODO(), config.Name, metav1.GetOptions{})
		if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			op.logger.Errorln(err.Error())
			return err
		}

		if errors.IsNotFound(err) {
			op.logger.Warnf("Deployment not found in namespace %s\n", config.Name)
			return nil
		}

		deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"] = strconv.Itoa(config.Max)
		deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"] = strconv.Itoa(config.Min)

		deploy, err = op.clientset.AppsV1().Deployments(config.Name).Update(context.TODO(), deploy, metav1.UpdateOptions{})
		if err == nil {
			op.logger.Printf("Success updating Deployment %s!", deploy.Name)
		}
		if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			op.logger.Errorln(err.Error())
			return err
		}
	}
	return nil
}

func (op HpaOperator) ScaleWithConcurrency() error {
	return fmt.Errorf("Not implemented.")
}
