package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
)

const PrecisionNAS = 18
const PrecisionNAX = 9
const naxContract = "n1etmdwczuAUCnMMvpGasfi8kwUbb2ddvRJ"
const stakingContract = "n214bLrE3nREcpRewHXF7qRDWCcaxRSiUdw"
const someAddress = "n1Jkdiq1H1HSXYJXtvDDkYm84Tmapo4hhMv"

var PrecisionDivNAS = decimal.New(1, PrecisionNAS)
var PrecisionDivNAX = decimal.New(1, PrecisionNAX)

type (
	API struct {
		url    string
		client *http.Client
	}
	AccountState struct {
		Result struct {
			Balance decimal.Decimal `json:"balance"`
			Nonce   string          `json:"nonce"`
			Type    int             `json:"type"`
			Height  string          `json:"height"`
			Pending string          `json:"pending"`
		} `json:"result"`
	}
	Block struct {
		Result struct {
			Hash          string `json:"hash"`
			ParentHash    string `json:"parent_hash"`
			Height        uint64 `json:"height,string"`
			Nonce         string `json:"nonce"`
			Coinbase      string `json:"coinbase"`
			Timestamp     int64  `json:"timestamp,string"`
			ChainID       uint64 `json:"chain_id"`
			ConsensusRoot struct {
				Timestamp   int64  `json:"timestamp,string"`
				Proposer    string `json:"proposer"`
				DynastyRoot string `json:"dynasty_root"`
			} `json:"consensus_root"`
			Miner        string        `json:"miner"`
			IsFinality   bool          `json:"is_finality"`
			Transactions []Transaction `json:"transactions"`
		} `json:"result"`
	}
	Transaction struct {
		Hash            string          `json:"hash"`
		ChainID         uint64          `json:"chain_id,string"`
		From            string          `json:"from"`
		To              string          `json:"to"`
		Value           decimal.Decimal `json:"value"`
		Nonce           int64           `json:"nonce,string"`
		Timestamp       int64           `json:"timestamp,string"`
		Type            string          `json:"type"`
		Data            string          `json:"data"`
		GasPrice        decimal.Decimal `json:"gas_price"`
		GasLimit        decimal.Decimal `json:"gas_limit"`
		ContractAddress string          `json:"contract_address"`
		Status          int             `json:"status"`
		GasUsed         decimal.Decimal `json:"gas_used"`
		BlockHeight     uint64          `json:"block_height,string"`
	}
	CallRequest struct {
		From     string       `json:"from"`
		To       string       `json:"to"`
		Value    string       `json:"value"`
		GasPrice string       `json:"gasPrice"`
		GasLimit string       `json:"gasLimit"`
		Contract CallContract `json:"contract"`
	}
	CallContract struct {
		Function string `json:"function"`
		Args     string `json:"args"`
	}
	ValidatorNode struct {
		ID   string `json:"id"`
		Info struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"info"`
		Accounts struct {
			Registrant       string `json:"registrant"`
			ConsensusManager string `json:"consensusManager"`
			GovManager       string `json:"govManager"`
			StakingAccount   string `json:"stakingAccount"`
		} `json:"accounts"`
		Type           int             `json:"type"`
		Status         int             `json:"status"`
		Online         bool            `json:"online"`
		Approved       bool            `json:"approved"`
		VoteValue      decimal.Decimal `json:"voteValue"`
		StakingValue   decimal.Decimal `json:"stakingValue"`
		StabilityIndex float64         `json:"stabilityIndex,string"`
	}
	Vote struct {
		Address string          `json:"address"`
		Value   decimal.Decimal `json:"value"`
	}

	Response struct {
		Result struct {
			Result string `json:"result"`
		} `json:"result"`
	}
)

func NewAPI(url string) *API {
	return &API{
		url:    url,
		client: &http.Client{},
	}
}

func (api *API) post(endpoint string, params interface{}, data interface{}) error {
	url := fmt.Sprintf("%s/%s", api.url, endpoint)
	var body []byte
	if params != nil {
		body, _ = json.Marshal(params)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("http.NewRequest: %s", err.Error())
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll: %s", err.Error())
	}
	err = json.Unmarshal(d, data)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %s", err.Error())
	}
	return nil
}

func (api *API) get(endpoint string, data interface{}) error {
	url := fmt.Sprintf("%s/%s", api.url, endpoint)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %s", err.Error())
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll: %s", err.Error())
	}
	err = json.Unmarshal(d, data)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %s", err.Error())
	}
	return nil
}

func (api *API) GetAccountState(address string) (state AccountState, err error) {
	err = api.post("v1/user/accountstate", map[string]interface{}{"address": address}, &state)
	return state, err
}

func (api *API) GetBlock(height uint64) (block Block, err error) {
	err = api.post("v1/user/getBlockByHeight", map[string]interface{}{
		"height":                height,
		"full_fill_transaction": true,
	}, &block)
	return block, err
}

func (api *API) GetLatestIrreversibleBlock() (block Block, err error) {
	err = api.get("v1/user/lib", &block)
	return block, err
}

func (api *API) GetNAXBalance(address string) (result decimal.Decimal, err error) {
	args, _ := json.Marshal([]string{address})
	contract := CallContract{
		Function: "balanceOf",
		Args:     string(args),
	}
	err = api.callContract(naxContract, contract, &result)
	return result, err
}

func (api *API) callContract(contractAddress string, contract CallContract, data interface{}) (err error) {
	call := CallRequest{
		From:     someAddress,
		To:       contractAddress,
		Value:    "0",
		GasPrice: "20000000000",
		GasLimit: "2000000",
		Contract: contract,
	}
	var response Response
	err = api.post("v1/user/call", call, &response)
	if err != nil {
		return fmt.Errorf("api.post: %s", err.Error())
	}
	err = json.Unmarshal([]byte(response.Result.Result), data)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %s", err.Error())
	}
	return nil
}

func (api *API) GetNodesList() (list []ValidatorNode, err error) {
	err = api.callContract(stakingContract,
		CallContract{
			Function: "getNodeList",
			Args:     "[]",
		},
		&list,
	)
	return list, nil
}

func (api *API) GetNodeVotesList(nodeID string) (list []Vote, err error) {
	args, _ := json.Marshal([]string{nodeID})
	err = api.callContract(stakingContract,
		CallContract{
			Function: "getNodeVoteStatistic",
			Args:     string(args),
		},
		&list,
	)
	return list, err
}
