package scales

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ScalesFacade struct {
	scaleHelper   scaleTypeHelperInterface
	k8sHelper     k8sHelperInterface
	scalerFactory *scalerFactory
	logger        *logrus.Logger
}

func NewScalesFacade(logger *logrus.Logger) *ScalesFacade {
	k8sHelper := newK8sHelper()
	return &ScalesFacade{
		scaleHelper:   newScaleTypeHelper(k8sHelper, logger, 500),
		k8sHelper:     k8sHelper,
		scalerFactory: &scalerFactory{},
		logger:        logger,
	}
}

// TODO: Refactor
// Returns a new clientset for internal use
func (s *ScalesFacade) GetClientset() (kubernetes.Interface, error) {
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

// TODO: Refactor
func (s *ScalesFacade) GetHpaInfo(clientset kubernetes.Interface, scaleConfigs ScaleConfigs, logger *logrus.Logger) (ScaleConfigs, error) {
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

// Update HPA list
func (s *ScalesFacade) UpdateWithConcurrency(scaleConfigs ScaleConfigs, sleep *time.Duration) {
	scaleCh := make(chan ScaleConfig)
	errorCh := make(chan error)

	// Checks if it is Hpa Operator
	scaleHelper := s.scaleHelper
	for scaleName, scaleConfig := range scaleConfigs {
		scaleConfig.Name = scaleName
		go func(config ScaleConfig) {
			err := scaleHelper.IdentifyHpaType(&config)

			if err != nil {
				s.logger.Warnf(err.Error())
				errorCh <- err
				return
			}

			scaleCh <- config
			s.logger.Debugf("%s config sent.\n", config.Name)
		}(scaleConfig)
	}

	for i := 0; i < len(scaleConfigs); i++ {
		select {
		case configs := <-scaleCh:
			s.logger.Debugf("%s config received.\n", configs.Name)
			scaler, err := s.scalerFactory.getScaler(configs.Type, s.k8sHelper, s.logger)

			if err != nil {
				s.logger.Errorln(err)
				return
			}

			err = scaler.Scale(configs)
			if err != nil {
				s.logger.Errorln(err)
			}

			time.Sleep(*sleep)
		case err := <-errorCh:
			s.logger.Errorln(err.Error())
		}
	}
}
