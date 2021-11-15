package scales

type ScaleConfigs map[string]ScaleConfig

type ScaleConfig struct {
	Min         int  `json:"min"`
	Max         int  `json:"max"`
	HpaOperator bool `json:"hpaOperator,omitempty"`
}
