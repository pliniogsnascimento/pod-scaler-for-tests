package scales

import (
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// Implements Scaler Interface
type hpaOperator struct {
	scaleConfigs ScaleConfigs
	logger       *logrus.Logger
	k8sHelper    k8sHelperInterface
}

func newHpaOperator(k8sHelper k8sHelperInterface, logger *logrus.Logger) *hpaOperator {
	return &hpaOperator{
		logger:    logger,
		k8sHelper: k8sHelper,
	}
}

func (op *hpaOperator) Scale(config ScaleConfig) error {
	deploy, err := op.k8sHelper.getDeploymentWithTimeout(config.Name, 500*time.Millisecond)

	if err != nil {
		return err
	}

	deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"] = strconv.Itoa(config.Max)
	deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"] = strconv.Itoa(config.Min)

	err = op.k8sHelper.updateDeployWithTimeout(deploy.Name, deploy, 500*time.Millisecond)

	return err
}
