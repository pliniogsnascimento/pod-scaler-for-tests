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

func UpdateHpaWithConcurrency(clientset kubernetes.Interface, scaleConfigs ScaleConfigs, logger *logrus.Logger, sleep *time.Duration) {
	scaleCh := make(chan ScaleConfig)
	chQuit := make(chan error)

	// Checks if it is Hpa Operator
	go func() {
		for scaleName, scaleConfig := range scaleConfigs {
			err := checkDeployHpaOp(clientset, scaleName, &scaleConfig, logger)

			if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
				logger.Errorln(err.Error())
				chQuit <- err
				return
			}

			if errors.IsNotFound(err) {
				logger.Warnf("Deployment not found in namespace %s\n", scaleName)
				continue
			}

			scaleConfig.Name = scaleName
			scaleCh <- scaleConfig
			logger.Debugf("%s config sent.\n", scaleConfig.Name)
		}
		chQuit <- nil
	}()

	for {
		select {
		case configs := <-scaleCh:
			logger.Debugf("%s config received.\n", configs.Name)
			if configs.HpaOperator {
				err := updateHpaOp(clientset, configs.Name, &configs, logger)
				if err != nil {
					logger.Errorln(err)
					return
				}
			} else {
				err := updateVanillaHpa(clientset, configs.Name, &configs, logger)
				if err != nil {
					logger.Errorln(err)
					return
				}
			}
			time.Sleep(*sleep)
		case err := <-chQuit:
			close(scaleCh)
			if err != nil {
				logger.Errorln(err)
				return
			}
			logger.Debugln("Channels were closed!")
			return
		default:
			continue
		}
	}
}

// Deprecated: This method is deprecated by using no concurrent implementation.
func UpdateHpa(clientset kubernetes.Interface, scaleConfigs ScaleConfigs, logger *logrus.Logger, sleep *time.Duration) error {
	// Checks if it is Hpa Operator
	for scaleName, scaleConfig := range scaleConfigs {
		err := checkDeployHpaOp(clientset, scaleName, &scaleConfig, logger)
		if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			logger.Errorln(err.Error())
			return err
		}

		if errors.IsNotFound(err) {
			logger.Warnf("Deployment not found in namespace %s\n", scaleName)
			continue
		}
		scaleConfigs[scaleName] = scaleConfig
	}

	for scaleName, configs := range scaleConfigs {
		if configs.HpaOperator {
			err := updateHpaOp(clientset, scaleName, &configs, logger)
			if err != nil {
				return err
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

func checkDeployHpaOp(client kubernetes.Interface, scaleName string, scaleConfig *ScaleConfig, logger *logrus.Logger) error {
	deploy, err := client.AppsV1().Deployments(scaleName).Get(context.TODO(), scaleName, metav1.GetOptions{})

	if errors.IsForbidden(err) || errors.IsUnauthorized(err) || errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		logger.Warnf("Deployment not found in namespace %s\n", scaleName)
		return fmt.Errorf("Deploy not found")
	}
	_, maxOk := deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"]
	_, minOk := deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"]

	if maxOk && minOk {
		logger.Debugf("%s uses Hpa Operator.\n", scaleName)
		scaleConfig.HpaOperator = true
	} else {
		logger.Debugf("%s does not use Hpa Operator.\n", scaleName)
		scaleConfig.HpaOperator = false
	}

	return nil
}

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

	deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"] = strconv.Itoa(configs.Max)
	deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"] = strconv.Itoa(configs.Min)

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

	if errors.IsNotFound(err) {
		logger.Warnf("HPA not found in namespace %s\n", scaleName)
		return nil
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

// Consider using this function
// func checkKubernetesErrors(err error, logger *logrus.Logger, deployName string) error {
// 	if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
// 		logger.Errorln(err.Error())
// 		return err
// 	}

// 	if errors.IsNotFound(err) {
// 		logger.Warnf("Deployment not found in namespace %s\n", deployName)
// 	}
// 	return nil
// }
