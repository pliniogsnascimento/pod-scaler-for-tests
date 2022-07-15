package scales

import (
	"context"
	"time"

	v1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type k8sHelperInterface interface {
	getDeploymentWithTimeout(deployName string, timeout time.Duration) (*v1.Deployment, error)
	getHpaWithTimeout(name string, timeout time.Duration) (*autoscalingv1.HorizontalPodAutoscaler, error)
	updateHpaWithTimeout(name string, hpaConfig *autoscalingv1.HorizontalPodAutoscaler, timeout time.Duration) error
	updateDeployWithTimeout(name string, deployConfig *v1.Deployment, timeout time.Duration) error
}

type k8sHelper struct {
	clientset kubernetes.Interface
	ctx       context.Context
}

func newk8sHelper(clientset kubernetes.Interface) *k8sHelper {
	return &k8sHelper{
		clientset: clientset,
		ctx:       context.Background(),
	}
}

func (k k8sHelper) getDeploymentWithTimeout(deployName string, timeout time.Duration) (*v1.Deployment, error) {
	ctx, cancel := context.WithTimeout(k.ctx, timeout*time.Millisecond)
	defer cancel()
	deploy, err := k.clientset.AppsV1().Deployments(deployName).Get(ctx, deployName, metav1.GetOptions{})

	if k.accessOrNotFoundError(err) {
		return nil, err
	}

	return deploy, nil
}

func (k k8sHelper) getHpaWithTimeout(name string, timeout time.Duration) (*autoscalingv1.HorizontalPodAutoscaler, error) {
	ctx, cancel := context.WithTimeout(k.ctx, timeout*time.Millisecond)
	defer cancel()
	hpa, err := k.clientset.AutoscalingV1().HorizontalPodAutoscalers(name).Get(ctx, name, metav1.GetOptions{})
	if k.accessOrNotFoundError(err) {
		return nil, err
	}

	return hpa, nil
}

func (k k8sHelper) executeUpdateWithTimeout(f func(client kubernetes.Interface, ctx context.Context) error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(k.ctx, timeout*time.Millisecond)
	defer cancel()
	return f(k.clientset, ctx)
}

func (k k8sHelper) updateHpaWithTimeout(name string, hpaConfig *autoscalingv1.HorizontalPodAutoscaler, timeout time.Duration) error {
	return k.executeUpdateWithTimeout(func(client kubernetes.Interface, ctx context.Context) error {
		_, err := client.AutoscalingV1().HorizontalPodAutoscalers(name).Update(ctx, hpaConfig, metav1.UpdateOptions{})
		if k.accessOrNotFoundError(err) {
			return err
		}
		return nil
	}, timeout)
}

func (k k8sHelper) updateDeployWithTimeout(name string, deployConfig *v1.Deployment, timeout time.Duration) error {
	return k.executeUpdateWithTimeout(func(client kubernetes.Interface, ctx context.Context) error {
		_, err := client.AppsV1().Deployments(name).Update(ctx, deployConfig, metav1.UpdateOptions{})
		if k.accessOrNotFoundError(err) {
			return err
		}
		return nil
	}, timeout)
}

func (k k8sHelper) accessError(err error) bool {
	return errors.IsForbidden(err) || errors.IsUnauthorized(err)
}

func (k k8sHelper) accessOrNotFoundError(err error) bool {
	return k.accessError(err) || errors.IsForbidden(err) || errors.IsUnauthorized(err)
}
