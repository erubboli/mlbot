package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/tidwall/gjson"
)

const (
	PRECISION         = 100000000000
	defaultAPIBaseURL = "https://api-server.mintlayer.org"
)

func getBlocks() (int64, error) {
	return getBlocksWithBaseURL(defaultAPIBaseURL)
}

func getBlocksWithBaseURL(baseURL string) (int64, error) {
	url := fmt.Sprintf("%s/api/v1/blocks", baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	blocks := gjson.GetBytes(body, "blocks").Int()
	return blocks, nil
}

func getPoolBalance(poolID string) (int64, error) {
	return getPoolBalanceWithBaseURL(defaultAPIBaseURL, poolID)
}

func getPoolBalanceWithBaseURL(baseURL, poolID string) (int64, error) {
	url := fmt.Sprintf("%s/api/v2/pool/%s", baseURL, poolID)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	atoms_balance := gjson.GetBytes(body, "staker_balance.atoms").Int()
	ml_balance := atoms_balance / PRECISION
	return ml_balance, nil
}

func getDelegationBalance(delegationID string) (int64, error) {
	return getDelegationBalanceWithBaseURL(defaultAPIBaseURL, delegationID)
}

func getDelegationBalanceWithBaseURL(baseURL, delegationID string) (int64, error) {
	url := fmt.Sprintf("%s/api/v2/delegation/%s", baseURL, delegationID)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	atoms_balance := gjson.GetBytes(body, "balance.atoms").Int()
	ml_balance := atoms_balance / PRECISION
	return ml_balance, nil
}
