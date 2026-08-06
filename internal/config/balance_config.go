package config

type BalanceConfig struct {
	Addr string `yaml:"addr" env:"BALANCE_ADDR" env-required:"true"`
}
