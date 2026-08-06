package config

import "time"

type WorkersConfig struct {
	PendingOrdersInterval  time.Duration `yaml:"pendingOrdersInterval" env-default:"5s"`
	PendingOrdersBatchSize int           `yaml:"pendingOrdersBatchSize" env-default:"10"`

	ProcessingOrderTimeout time.Duration `yaml:"processingOrderTimeout" env-default:"5m"`
	RecoveryOrderInterval  time.Duration `yaml:"recoveryOrderInterval" env-default:"3m"`
	RecoveryOrderBatchSize int           `yaml:"recoveryOrderBatchSize" env-default:"10"`
}
