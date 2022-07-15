package scales

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	fakeLogger  logrus.Logger
	deployMocks map[string]v1.Deployment
	client      = fake.NewSimpleClientset()

	fakeDeploymentModel = v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "nsname",
			Labels: map[string]string{
				"app": "myfakeapp",
			},
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "myfakeapp",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "myfakeapp",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "myapp",
							Image: "nginx",
						},
					},
				},
			},
		},
	}

	minReplicas = int32(3)

	fakeHpaModel = autoscalingv1.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "fakeHpa",
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "my-deploy",
				APIVersion: "apps/v1",
			},
			MinReplicas: &minReplicas,
			MaxReplicas: 6,
		},
	}
)

func init() {
	deployMocks = make(map[string]v1.Deployment)

	for i := 0; i <= 10; i++ {
		fakeDeployHpaOp := fakeDeploymentModel
		fakeDeployHpaOp.Name = "HpaOpDeploy" + strconv.Itoa(i)
		fakeDeployHpaOp.ObjectMeta.Annotations = map[string]string{
			"hpa.autoscaling.banzaicloud.io/maxReplicas": "3",
			"hpa.autoscaling.banzaicloud.io/minReplicas": "1",
		}

		deployMocks[fakeDeployHpaOp.Name] = fakeDeployHpaOp
	}

	fakeDeploy := fakeDeploymentModel
	fakeDeploy.Name = "NormalDeploy"

	deployMocks["NormalDeploy"] = fakeDeploy

	for _, deploy := range deployMocks {
		fmt.Printf("Creating mock deploy %s.\n", deploy.Name)
		if _, err := client.AppsV1().Deployments(deploy.Name).Create(context.TODO(), &deploy, metav1.CreateOptions{}); err != nil {
			panic("Unable to create mocks")
		}
		respectiveHpa := fakeHpaModel
		respectiveHpa.Name = deploy.Name
		respectiveHpa.Spec.ScaleTargetRef.Name = deploy.Name

		if _, err := client.AutoscalingV1().HorizontalPodAutoscalers(deploy.Name).Create(context.TODO(), &respectiveHpa, metav1.CreateOptions{}); err != nil {
			panic("Unable to create mocks")
		}
	}

	fakeLogger = *logrus.New()
	// fakeLogger.Level = logrus.PanicLevel
	fakeLogger.Level = logrus.DebugLevel
}

func TestUpdateHpaOperatorSuccess(t *testing.T) {
	scaleConfigs := &ScaleConfig{
		Min:         3,
		Max:         5,
		HpaOperator: true,
	}

	if err := updateHpaOp(client, deployMocks["HpaOpDeploy0"].Name, scaleConfigs, &fakeLogger); err != nil {
		t.Errorf("Error: %s", err)
	}
}

func TestUpdateMultipleHpaWithConcurrencySuccess(t *testing.T) {
	scaleConfigs := ScaleConfigs{
		deployMocks["HpaOpDeploy0"].Name: {
			Min:         3,
			Max:         5,
			HpaOperator: true,
		},
		deployMocks["HpaOpDeploy1"].Name: {
			Min:         3,
			Max:         5,
			HpaOperator: true,
		},
		deployMocks["HpaOpDeploy2"].Name: {
			Min:         3,
			Max:         5,
			HpaOperator: true,
		},
		deployMocks["HpaOpDeploy3"].Name: {
			Min:         3,
			Max:         5,
			HpaOperator: true,
		},
		deployMocks["HpaOpDeploy4"].Name: {
			Min:         3,
			Max:         5,
			HpaOperator: true,
		},
		deployMocks["HpaOpDeploy5"].Name: {
			Min:         3,
			Max:         5,
			HpaOperator: true,
		},
		deployMocks["HpaOpDeploy6"].Name: {
			Min:         3,
			Max:         5,
			HpaOperator: true,
		},
		deployMocks["HpaOpDeploy7"].Name: {
			Min:         3,
			Max:         5,
			HpaOperator: true,
		},
		deployMocks["HpaOpDeploy8"].Name: {
			Min:         3,
			Max:         5,
			HpaOperator: true,
		},
		deployMocks["HpaOpDeploy9"].Name: {
			Min:         3,
			Max:         5,
			HpaOperator: true,
		},
		deployMocks["NormalDeploy"].Name: {
			Min:         3,
			Max:         5,
			HpaOperator: false,
		},
	}

	sleep := time.Duration(time.Second * 1)
	UpdateHpaWithConcurrency(client, scaleConfigs, &fakeLogger, &sleep)

	checkIfUpdated(scaleConfigs, client, t)
}

func TestVanillaScaleSuccess(t *testing.T) {
	scaleConfigs := ScaleConfigs{
		deployMocks["HpaOpDeploy0"].Name: {
			Name: "HpaOpDeploy0",
			Min:  30,
			Max:  50,
		},
		deployMocks["NormalDeploy"].Name: {
			Name: "NormalDeploy",
			Min:  30,
			Max:  50,
		},
	}
	scaler := NewVanillaHpa(client, &fakeLogger)

	for _, config := range scaleConfigs {
		err := scaler.Scale(config)

		if err != nil {
			t.Errorf(err.Error())
			t.FailNow()
		}
	}

	checkIfUpdated(scaleConfigs, client, t)
}

func TestVanillaScaleError(t *testing.T) {
	scaleConfigs := ScaleConfigs{
		deployMocks["HpaOpDeploy0"].Name: {
			Min: 30,
			Max: 50,
		},
		deployMocks["NormalDeploy"].Name: {
			Min: 30,
			Max: 50,
		},
	}
	scaler := NewVanillaHpa(client, &fakeLogger)
	for _, config := range scaleConfigs {
		err := scaler.Scale(config)
		if err == nil {
			t.FailNow()
		}

		t.Log(err.Error())
	}
}

func checkIfUpdated(scaleConfigs ScaleConfigs, client kubernetes.Interface, t *testing.T) {
	for name, config := range scaleConfigs {
		deploy, _ := client.AppsV1().Deployments(name).Get(context.TODO(), name, metav1.GetOptions{})

		if config.HpaOperator && deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"] != strconv.Itoa(config.Max) {
			t.Errorf("%s max not updated: %s\n", name, deploy.Annotations["hpa.autoscaling.banzaicloud.io/maxReplicas"])
		}

		if config.HpaOperator && deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"] != strconv.Itoa(config.Min) {
			t.Errorf("%s min not updated: %s\n", name, deploy.Annotations["hpa.autoscaling.banzaicloud.io/minReplicas"])
		}
	}
}
