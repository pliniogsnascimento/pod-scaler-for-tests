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
	k8sHelper k8sHelperInterface
}

func NewScaleChecker(scaleConfigs *ScaleConfigs, clientset kubernetes.Interface, logger *logrus.Logger, timeout time.Duration) *ScaleTypeHelper {
	return &ScaleTypeHelper{
		ScaleConfigs: *scaleConfigs,
		clientset:    clientset,
		logger:       logger,
		ctx:          context.Background(),
		timeout:      timeout * time.Millisecond,
		helper:       newk8sHelper(clientset)
	}
}

func (s ScaleTypeHelper) ModifyHpaOpCheck() error {
	helper := s.k8sHelper
	for _, scaleConfig := range s.ScaleConfigs {
		deploy, err := helper.getDeploymentWithTimeout(scaleConfig.Name, 500)
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
