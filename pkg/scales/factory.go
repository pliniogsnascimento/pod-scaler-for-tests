package scales

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

func getScaler(scalerType string, clientset kubernetes.Interface, logger *logrus.Logger) (Scaler, error) {
	switch scalerType {
	case "VanillaHpa":
		return NewVanillaHpa(clientset, logger), nil
	case "HpaOperator":
		return NewHpaOperator(clientset, logger), nil
	default:
		return nil, fmt.Errorf("Not valid scaler type")
	}
}
