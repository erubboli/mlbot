package main

import "strings"

type BalanceClient interface {
	GetPoolBalance(poolID string) (int64, error)
	GetDelegationBalance(delegationID string) (int64, error)
}

type HTTPBalanceClient struct {
	baseURL string
}

func NewHTTPBalanceClient(baseURL string) *HTTPBalanceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	return &HTTPBalanceClient{baseURL: baseURL}
}

func (c *HTTPBalanceClient) GetPoolBalance(poolID string) (int64, error) {
	return getPoolBalanceWithBaseURL(c.baseURL, poolID)
}

func (c *HTTPBalanceClient) GetDelegationBalance(delegationID string) (int64, error) {
	return getDelegationBalanceWithBaseURL(c.baseURL, delegationID)
}
