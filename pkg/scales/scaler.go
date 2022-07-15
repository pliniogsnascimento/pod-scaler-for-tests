package scales

import (
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
)

type scaler interface {
	Scale(config ScaleConfig) error
}

type ScaleConfigs map[string]ScaleConfig

type ScaleConfig struct {
	Name        string `json:"name"`
	Min         int    `json:"min"`
	Max         int    `json:"max"`
	HpaOperator bool   `json:"hpaOperator,omitempty"`
	Type        string `json:"type,omitempty"`
}

type scaleTypeHelper struct {
	logger    *logrus.Logger
	timeout   time.Duration
	k8sHelper k8sHelperInterface
}

func newScaleTypeHelper(k8sHelper *k8sHelper, logger *logrus.Logger, timeout time.Duration) *scaleTypeHelper {
	return &scaleTypeHelper{
		logger:    logger,
		timeout:   timeout,
		k8sHelper: k8sHelper,
	}
}

func (s scaleTypeHelper) IdentifyHpaType(scaleConfig *ScaleConfig) error {
	helper := s.k8sHelper
	deploy, err := helper.getDeploymentWithTimeout(scaleConfig.Name, s.timeout)

	if err != nil {
		return err
	}

	s.checkIfHpaOp(deploy, scaleConfig)
	return nil
}

func (s scaleTypeHelper) checkIfHpaOp(deploy *v1.Deployment, scaleConfig *ScaleConfig) {
	_, maxOk := deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"]
	_, minOk := deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"]

	if maxOk && minOk {
		s.logger.Debugf("%s uses Hpa Operator.\n", scaleConfig.Name)
		scaleConfig.HpaOperator = true
		scaleConfig.Type = "HpaOperator"
	} else {
		s.logger.Debugf("%s does not use Hpa Operator.\n", scaleConfig.Name)
		scaleConfig.HpaOperator = false
		scaleConfig.Type = "VanillaHpa"
	}
}
