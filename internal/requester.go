package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var client = &http.Client{}
var clientUrl = "https://rest.bd.evmos.org:1317/"
var PreHeight = "5074186"
var PostHeight = "5074187"

func GetBalances(address string, height string) BalancesResponse {
	balance_start := "cosmos/bank/v1beta1/balances/"
	url := balance_start + address
	body := makeRequest(url, height)
	m := &BalancesResponse{}
	err := json.Unmarshal(body, &m)
	if err != nil {
		fmt.Println("Error parsing the body", body, err)
		panic("Stop processing")
	}
	return *m
}

func GetEvmosBalance(address string, height string) BalanceResponse {
	balance_start := "cosmos/bank/v1beta1/balances/"
	balance_end := "/by_denom?denom=aevmos"
	url := balance_start + address + balance_end
	body := makeRequest(url, height)
	m := &BalanceResponse{}
	err := json.Unmarshal(body, &m)
	if err != nil {
		fmt.Println("Error parsing the body", body, err)
		panic("Stop processing")
	}
	return *m
}

func GetAccount(address, height string) AccountResponse {
	account_start := "cosmos/auth/v1beta1/accounts/"
	url := account_start + address
	body := makeRequest(url, height)
	m := &AccountResponse{}
	err := json.Unmarshal(body, &m)
	if err != nil {
		fmt.Println("Error parsing the body", body, err)
		panic("Stop processing")
	}
	return *m
}

func makeRequest(endpoint string, height string) []byte {
	req, _ := http.NewRequest("GET", clientUrl+endpoint, nil)
	req.Header.Set("x-cosmos-block-height", height)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error http request", endpoint, err)
		panic("Stop processing")
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading the response body", endpoint, err)
		panic("Stop processing")
	}
	return body
}
