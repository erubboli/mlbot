package main

type BalanceClient interface {
	GetPoolBalance(poolID string) (int64, error)
	GetDelegationBalance(delegationID string) (int64, error)
}

type HTTPBalanceClient struct{}

func (c *HTTPBalanceClient) GetPoolBalance(poolID string) (int64, error) {
	return getPoolBalance(poolID)
}

func (c *HTTPBalanceClient) GetDelegationBalance(delegationID string) (int64, error) {
	return getDelegationBalance(delegationID)
}
