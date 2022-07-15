package scales

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// TODO: Refactor to implement new contract
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

func (op hpaOperator) Scale(config ScaleConfig) error {
	// for _, config := range op.scaleConfigs {
	// 	deploy, err := op.clientset.AppsV1().Deployments(config.Name).Get(context.TODO(), config.Name, metav1.GetOptions{})
	// 	if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
	// 		op.logger.Errorln(err.Error())
	// 		return err
	// 	}

	// 	if errors.IsNotFound(err) {
	// 		op.logger.Warnf("Deployment not found in namespace %s\n", config.Name)
	// 		return nil
	// 	}

	// 	deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"] = strconv.Itoa(config.Max)
	// 	deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"] = strconv.Itoa(config.Min)

	// 	deploy, err = op.clientset.AppsV1().Deployments(config.Name).Update(context.TODO(), deploy, metav1.UpdateOptions{})
	// 	if err == nil {
	// 		op.logger.Printf("Success updating Deployment %s!", deploy.Name)
	// 	}
	// 	if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
	// 		op.logger.Errorln(err.Error())
	// 		return err
	// 	}
	// }
	return fmt.Errorf("Not Implemented.")
}
