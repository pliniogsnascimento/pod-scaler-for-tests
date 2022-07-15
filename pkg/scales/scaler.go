package scales

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Scaler interface {
	Scale() error
}

type ScaleConfigs map[string]ScaleConfig

type ScaleConfig struct {
	Name        string `json:"name"`
	Min         int    `json:"min"`
	Max         int    `json:"max"`
	HpaOperator bool   `json:"hpaOperator,omitempty"`
}

type ScaleTypeHelper struct {
	ScaleConfigs
	clientset kubernetes.Interface
	logger    *logrus.Logger
	ctx       context.Context
	timeout   time.Duration
}

func NewScaleChecker(scaleConfigs *ScaleConfigs, clientset kubernetes.Interface, logger *logrus.Logger, timeout time.Duration) *ScaleTypeHelper {
	return &ScaleTypeHelper{
		ScaleConfigs: *scaleConfigs,
		clientset:    clientset,
		logger:       logger,
		ctx:          context.Background(),
		timeout:      timeout * time.Millisecond,
	}
}

func (s ScaleTypeHelper) ModifyHpaOpCheck() error {
	for _, scaleConfig := range s.ScaleConfigs {
		ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
		deploy, err := s.clientset.AppsV1().Deployments(scaleConfig.Name).Get(ctx, scaleConfig.Name, metav1.GetOptions{})
		cancel()

		if errors.IsForbidden(err) || errors.IsUnauthorized(err) || errors.IsNotFound(err) {
			return err
		}

		if errors.IsNotFound(err) {
			s.logger.Warnf("Deployment not found in namespace %s\n", scaleConfig.Name)
			return fmt.Errorf("Deploy not found")
		}
		s.checkIfHpaOp(deploy, scaleConfig)
	}

	return nil
}

func (s ScaleTypeHelper) checkIfHpaOp(deploy *v1.Deployment, scaleConfig ScaleConfig) {
	_, maxOk := deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"]
	_, minOk := deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"]

	if maxOk && minOk {
		s.logger.Debugf("%s uses Hpa Operator.\n", scaleConfig.Name)
		scaleConfig.HpaOperator = true
	} else {
		s.logger.Debugf("%s does not use Hpa Operator.\n", scaleConfig.Name)
		scaleConfig.HpaOperator = false
	}
}
