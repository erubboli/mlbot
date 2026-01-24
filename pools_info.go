package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

const (
	PRECISION         = 100000000000
	defaultAPIBaseURL = "https://api-server.mintlayer.org"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func getBlocks() (int64, error) {
	return getBlocksWithBaseURL(defaultAPIBaseURL)
}

func getBlocksWithBaseURL(baseURL string) (int64, error) {
	url := fmt.Sprintf("%s/api/v1/blocks", baseURL)
	resp, err := getWithRetry(url, 3)
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
	resp, err := getWithRetry(url, 3)
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
	resp, err := getWithRetry(url, 3)
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

func getWithRetry(url string, attempts int) (*http.Response, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		resp, err := httpClient.Get(url)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isRetryableHTTPError(err) {
			break
		}
		time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
	}
	return nil, lastErr
}

func isRetryableHTTPError(err error) bool {
	if errors.Is(err, io.EOF) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}
	return false
}
