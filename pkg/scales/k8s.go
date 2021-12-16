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
	"k8s.io/client-go/rest"
)

// Returns a new clientset for internal use
func GetClientset() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()

	if err != nil {

		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client := kubernetes.Interface(clientset)
	return client, nil
}

func GetHpaInfo(clientset kubernetes.Interface, scaleConfigs ScaleConfigs, logger *logrus.Logger) (ScaleConfigs, error) {
	currentConfig := make(ScaleConfigs)
	for name := range scaleConfigs {
		hpa, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(name).Get(context.TODO(), name, metav1.GetOptions{})
		if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			logger.Errorln(err.Error())
			return nil, err
		}

		if errors.IsNotFound(err) {
			logger.Warnf("HPA not found in namespace %s\n", name)
			continue
		}

		currentConfig[hpa.Name] = ScaleConfig{
			Min: int(*hpa.Spec.MinReplicas),
			Max: int(hpa.Spec.MaxReplicas),
		}
	}
	return currentConfig, nil
}

func UpdateHpa(clientset kubernetes.Interface, scaleConfigs ScaleConfigs, logger *logrus.Logger, sleep *time.Duration) error {
	// ch := make(chan ScaleConfig, 10)
	// configsUpdated := ScaleConfigs{}

	// for scaleName, scaleConfig := range scaleConfigs {
	// 	go checkDeployHpaOp(clientset, scaleName, scaleConfig, logger, ch)
	// }

	// for i := 0; i >= len(scaleConfigs); i++ {
	// 	configUpdated := <-ch
	// 	configsUpdated[configUpdated.Name] = configUpdated
	// }

	for scaleName, configs := range scaleConfigs {
		if configs.HpaOperator {
			err := updateHpaOp(clientset, scaleName, &configs, logger)
			if err != nil {
				switch err.Error() {
				case "Not HPA Operator!":
					logger.Errorf("%s %s", scaleName, err)
					continue
				default:
					return err
				}
			}
		} else {
			err := updateVanillaHpa(clientset, scaleName, &configs, logger)
			if err != nil {
				return err
			}
		}
		time.Sleep(*sleep)
	}
	logger.Infoln("Done processing update request!")
	return nil
}

// func checkDeployHpaOp(client kubernetes.Interface, scaleName string, scaleConfig ScaleConfig, logger *logrus.Logger, ch chan ScaleConfig) {
// 	deploy, err := client.AppsV1().Deployments(scaleName).Get(context.TODO(), scaleName, metav1.GetOptions{})

// 	if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
// 		logger.Errorln(err.Error())
// 		ch <- nil
// 		// return err
// 	}

// 	if errors.IsNotFound(err) {
// 		logger.Warnf("Deployment not found in namespace %s\n", scaleName)
// 		// return nil
// 	}
// 	scaleConfig.Name = scaleName

// 	_, maxOk := deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"]
// 	_, minOk := deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"]

// 	if maxOk && minOk {
// 		scaleConfig.HpaOperator = true
// 		ch <- scaleConfig
// 	} else {
// 		scaleConfig.HpaOperator = true
// 		ch <- scaleConfig
// 	}
// }

func updateHpaOp(clientset kubernetes.Interface, scaleName string, configs *ScaleConfig, logger *logrus.Logger) error {
	deploy, err := clientset.AppsV1().Deployments(scaleName).Get(context.TODO(), scaleName, metav1.GetOptions{})
	if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
		logger.Errorln(err.Error())
		return err
	}

	if errors.IsNotFound(err) {
		logger.Warnf("Deployment not found in namespace %s\n", scaleName)
		return nil
	}

	// checks if HPA Operator Annotation exists. TODO: Create specific errors.
	if _, ok := deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"]; ok {
		deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"] = strconv.Itoa(configs.Max)
	} else {
		return fmt.Errorf("Not HPA Operator!")
	}
	if _, ok := deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"]; ok {
		deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"] = strconv.Itoa(configs.Min)
	} else {
		return fmt.Errorf("Not HPA Operator!")
	}

	deploy, err = clientset.AppsV1().Deployments(scaleName).Update(context.TODO(), deploy, metav1.UpdateOptions{})
	if err == nil {
		logger.Printf("Success updating Deployment %s!", deploy.Name)
	}
	if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
		logger.Errorln(err.Error())
		return err
	}
	return nil
}

func updateVanillaHpa(clientset kubernetes.Interface, scaleName string, configs *ScaleConfig, logger *logrus.Logger) error {
	hpa, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(scaleName).Get(context.TODO(), scaleName, metav1.GetOptions{})
	if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
		logger.Errorln(err.Error())
		return err
	}

	minReplicas := int32(configs.Min)
	hpa.Spec.MinReplicas = &minReplicas
	hpa.Spec.MaxReplicas = int32(configs.Max)

	_, err = clientset.AutoscalingV1().HorizontalPodAutoscalers(scaleName).Update(context.TODO(), hpa, metav1.UpdateOptions{})
	if err == nil {
		logger.Printf("Success updating HPA %s!", hpa.Name)
	}
	if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
		logger.Errorln(err.Error())
		return err
	}
	return nil
}
