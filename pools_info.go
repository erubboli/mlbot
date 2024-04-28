package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/tidwall/gjson"
)

const PRECISION = 100000000000

func getBlocks() (int64, error) {
	url := "https://api-server.mintlayer.org/api/v1/blocks"
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
	url := fmt.Sprintf("https://api-server.mintlayer.org/api/v2/pool/%s", poolID)
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
	url := fmt.Sprintf("https://api-server.mintlayer.org/api/v2/delegation/%s", delegationID)
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
