package scales

import (
	"context"
	"log"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()

	if err != nil {

		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func GetHpaInfo(clientset *kubernetes.Clientset, scaleConfigs ScaleConfigs, logger *log.Logger) (ScaleConfigs, error) {
	currentConfig := make(ScaleConfigs)
	for name, _ := range scaleConfigs {
		hpa, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(name).Get(context.TODO(), name, metav1.GetOptions{})
		if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			logger.Println(err.Error())
			return nil, err
		}

		if errors.IsNotFound(err) {
			logger.Printf("HPA not found in namespace %s\n", name)
			continue
		}

		currentConfig[hpa.Name] = ScaleConfig{
			Min: int(*hpa.Spec.MinReplicas),
			Max: int(hpa.Spec.MaxReplicas),
		}
	}
	return currentConfig, nil
}

func updateHpa() {

}
