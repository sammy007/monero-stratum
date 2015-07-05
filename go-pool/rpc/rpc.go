package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"../pool"
)

type RPCClient struct {
	url    string
	client *http.Client
}

type GetBlockTemplateReply struct {
	Blob           string `json:"blocktemplate_blob"`
	Difficulty     int64  `json:"difficulty"`
	ReservedOffset int    `json:"reserved_offset"`
	Height         int64  `json:"height"`
	PrevHash       string `json:"prev_hash"`
}

type JSONRpcResp struct {
	Id     *json.RawMessage       `json:"id"`
	Result *json.RawMessage       `json:"result"`
	Error  map[string]interface{} `json:"error"`
}

func NewRPCClient(cfg *pool.Config) *RPCClient {
	url := fmt.Sprintf("http://%s:%v/json_rpc", cfg.Daemon.Host, cfg.Daemon.Port)
	rpcClient := &RPCClient{url: url}
	timeout, _ := time.ParseDuration(cfg.Daemon.Timeout)
	rpcClient.client = &http.Client{
		Timeout: timeout,
	}
	return rpcClient
}

func (r *RPCClient) GetBlockTemplate(reserveSize int, address string) (GetBlockTemplateReply, error) {
	params := map[string]interface{}{"reserve_size": reserveSize, "wallet_address": address}

	rpcResp, err := r.doPost(r.url, "getblocktemplate", params)
	var reply GetBlockTemplateReply
	if err != nil {
		return reply, err
	}
	if rpcResp.Error != nil {
		return reply, errors.New(string(rpcResp.Error["message"].(string)))
	}

	err = json.Unmarshal(*rpcResp.Result, &reply)
	return reply, err
}

func (r *RPCClient) SubmitBlock(hash string) (JSONRpcResp, error) {
	rpcResp, err := r.doPost(r.url, "submitblock", []string{hash})
	if err != nil {
		return rpcResp, err
	}
	if rpcResp.Error != nil {
		return rpcResp, errors.New(string(rpcResp.Error["message"].(string)))
	}
	return rpcResp, nil
}

func (r *RPCClient) doPost(url string, method string, params interface{}) (JSONRpcResp, error) {
	jsonReq := map[string]interface{}{"id": "0", "method": method, "params": params}
	data, _ := json.Marshal(jsonReq)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Length", (string)(len(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.client.Do(req)
	var rpcResp JSONRpcResp

	if err != nil {
		return rpcResp, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &rpcResp)
	return rpcResp, err
}
