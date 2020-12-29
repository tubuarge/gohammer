package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type RPCClient struct {
	Client *http.Client
}

func NewRPCClient() *RPCClient {
	rpcClient := &RPCClient{}
	httpClient := &http.Client{}

	rpcClient.Client = httpClient

	return rpcClient
}

type JSONRPCResp struct {
	ID     *json.RawMessage       `json:"id"`
	Result *json.RawMessage       `json:"result"`
	Error  map[string]interface{} `json:"error"`
}

// IsNodeUp sends a `web3_clientVersion` RPC request to the given node.
// If RPC response is not nil and there is no error returns true otherwise
// returns false.
func (r *RPCClient) IsNodeUp(nodeUrl string) (bool, error) {
	resp, err := r.doPost(nodeUrl, "web3_clientVersion", nil)
	if err != nil {
		return false, err
	}

	if resp.Result != nil {
		return true, nil
	}
	return false, nil
}

func (r *RPCClient) doPost(url, method string, params interface{}) (*JSONRPCResp, error) {
	jsonReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      0,
	}

	data, _ := json.Marshal(jsonReq)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Length", string(len(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp *JSONRPCResp
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		return nil, err
	}

	if rpcResp.Error != nil {
		return nil, errors.New(rpcResp.Error["message"].(string))
	}

	return rpcResp, err
}
