package scales

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestIdentifyHpaType_VanillaSuccess(t *testing.T) {
	vanillaScaleConfig := ScaleConfig{
		Name: deployMocks["NormalDeploy"].Name,
		Min:  30,
		Max:  50,
	}
	deployMock := deployMocks["NormalDeploy"]

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockk8sHelperInterface(ctrl)
	m.
		EXPECT().
		getDeploymentWithTimeout(gomock.Any(), gomock.Any()).
		Return(&deployMock, nil)

	scaleHelper := newScaleTypeHelper(m, &fakeLogger, 500)
	err := scaleHelper.IdentifyHpaType(&vanillaScaleConfig)

	assert.Empty(t, err)
	assert.Equal(t, false, vanillaScaleConfig.HpaOperator)
	assert.Equal(t, "VanillaHpa", vanillaScaleConfig.Type)
}

func TestIdentifyHpaType_OperatorSuccess(t *testing.T) {
	vanillaScaleConfig := ScaleConfig{
		Name: deployMocks["HpaOpDeploy0"].Name,
		Min:  30,
		Max:  50,
	}
	deployMock := deployMocks["HpaOpDeploy0"]

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockk8sHelperInterface(ctrl)
	m.
		EXPECT().
		getDeploymentWithTimeout(gomock.Any(), gomock.Any()).
		Return(&deployMock, nil)

	scaleHelper := newScaleTypeHelper(m, &fakeLogger, 500)
	err := scaleHelper.IdentifyHpaType(&vanillaScaleConfig)

	assert.Empty(t, err)
	assert.Equal(t, true, vanillaScaleConfig.HpaOperator)
	assert.Equal(t, "HpaOperator", vanillaScaleConfig.Type)
}

func TestIdentifyHpaType_Fail(t *testing.T) {
	vanillaScaleConfig := ScaleConfig{
		Name: deployMocks["HpaOpDeploy0"].Name,
		Min:  30,
		Max:  50,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockk8sHelperInterface(ctrl)
	m.
		EXPECT().
		getDeploymentWithTimeout(gomock.Any(), gomock.Any()).
		Return(nil, fmt.Errorf("Fake error"))

	scaleHelper := newScaleTypeHelper(m, &fakeLogger, 500)
	err := scaleHelper.IdentifyHpaType(&vanillaScaleConfig)

	assert.NotNil(t, err)
}

func TestScaleHpaOperator_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	k8sHelperMock := NewMockk8sHelperInterface(ctrl)

	fakeDeploy := fakeDeploymentModel
	fakeDeploy.Annotations = map[string]string{}

	k8sHelperMock.
		EXPECT().
		getDeploymentWithTimeout(gomock.Any(), gomock.Any()).
		Return(&fakeDeploy, nil).
		AnyTimes()

	k8sHelperMock.
		EXPECT().
		updateDeployWithTimeout(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	config := ScaleConfig{
		Min:         3,
		Max:         5,
		HpaOperator: true,
	}

	hpaOp := newHpaOperator(k8sHelperMock, &fakeLogger)
	hpaOp.Scale(config)
}
