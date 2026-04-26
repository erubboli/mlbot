package main

import (
	"context"
	"strconv"
	"time"

	"github.com/mintlayer/go-sdk/indexer"
)

type BalanceClient interface {
	GetPoolBalance(poolID string) (int64, error)
	GetDelegationBalance(delegationID string) (int64, error)
}

type HTTPBalanceClient struct {
	idxClient *indexer.Client
}

func NewHTTPBalanceClient(baseURL string) *HTTPBalanceClient {
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	return &HTTPBalanceClient{
		idxClient: indexer.New(baseURL, indexer.WithTimeout(10*time.Second)),
	}
}

func (c *HTTPBalanceClient) GetPoolBalance(poolID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := c.idxClient.GetPool(ctx, poolID)
	if err != nil {
		return 0, err
	}
	atoms, err := strconv.ParseInt(pool.StakerBalance.Atoms, 10, 64)
	if err != nil {
		return 0, err
	}
	return atoms / PRECISION, nil
}

func (c *HTTPBalanceClient) GetDelegationBalance(delegationID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	delegation, err := c.idxClient.GetDelegation(ctx, delegationID)
	if err != nil {
		return 0, err
	}
	atoms, err := strconv.ParseInt(delegation.Balance.Atoms, 10, 64)
	if err != nil {
		return 0, err
	}
	return atoms / PRECISION, nil
}
