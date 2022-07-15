package scales

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type scalerFactoryInterface interface {
	getScaler(scalerType string, clientset kubernetes.Interface, logger *logrus.Logger) (scaler, error)
}

type scalerFactory struct{}

func (s *scalerFactory) getScaler(scalerType string, k8sHelper *k8sHelper, logger *logrus.Logger) (scaler, error) {
	switch scalerType {
	case "VanillaHpa":
		return newVanillaHpa(k8sHelper, logger), nil
	case "HpaOperator":
		return newHpaOperator(k8sHelper, logger), nil
	default:
		return nil, fmt.Errorf("Not valid scaler type")
	}
}
