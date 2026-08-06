package domain

type BalanceStatus string

const (
	BalanceStatusSuccess BalanceStatus = "SUCCESS"
	BalanceStatusFailed  BalanceStatus = "FAILED"
	BalanceStatusNoMoney BalanceStatus = "NOMONEY"
)
