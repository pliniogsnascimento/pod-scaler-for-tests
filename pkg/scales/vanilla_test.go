package scales

import (
	"fmt"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	fakeHpaMocks = v1.HorizontalPodAutoscaler{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1.HorizontalPodAutoscalerSpec{},
		Status:     v1.HorizontalPodAutoscalerStatus{},
	}
)

func TestVanillaScaleSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	scaleConfig := ScaleConfig{
		Name:        "NormalDeploy",
		Min:         30,
		Max:         50,
		HpaOperator: false,
		Type:        "VanillaHpa",
	}

	k8sHelperMock := NewMockk8sHelperInterface(ctrl)
	k8sHelperMock.
		EXPECT().
		getHpaWithTimeout(scaleConfig.Name, 500*time.Millisecond).
		Return(&fakeHpaMocks, nil)

	k8sHelperMock.
		EXPECT().
		updateHpaWithTimeout(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	scaler := newVanillaHpa(k8sHelperMock, &fakeLogger)
	err := scaler.Scale(scaleConfig)

	assert.Nil(t, err)
	assert.Equal(t, int32(scaleConfig.Min), *fakeHpaMocks.Spec.MinReplicas)
	assert.Equal(t, int32(scaleConfig.Max), fakeHpaMocks.Spec.MaxReplicas)
}

func TestVanillaScaleError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	k8sHelperMock := NewMockk8sHelperInterface(ctrl)
	k8sHelperMock.
		EXPECT().
		getHpaWithTimeout(gomock.Any(), gomock.Any()).
		AnyTimes().Return(nil, fmt.Errorf("Fake error"))

	scaleConfig := ScaleConfig{
		Min: 30,
		Max: 50,
	}

	scaler := newVanillaHpa(k8sHelperMock, &fakeLogger)
	err := scaler.Scale(scaleConfig)

	assert.NotNil(t, err)
}
