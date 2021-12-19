package scales

type ScaleConfigs map[string]ScaleConfig

type ScaleConfig struct {
	Name        string `json:"name"`
	Min         int    `json:"min"`
	Max         int    `json:"max"`
	HpaOperator bool   `json:"hpaOperator,omitempty"`
}
