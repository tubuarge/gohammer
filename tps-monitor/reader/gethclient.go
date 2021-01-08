package reader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const blockRequest = `{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x",false],"id":2}`
const blockNumberRequest = `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`

var blockNotFound = fmt.Errorf("block not found")

type ErrMsg struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type BlockNumberResult struct {
	Result string `json:"result"`
	Error  ErrMsg `json:"error"`
}

type BlockDataRaw struct {
	Number       string   `json:"number"`
	GasLimit     string   `json:"gasLimit"`
	GasUsed      string   `json:"gasUsed"`
	Time         string   `json:"timestamp"`
	Transactions []string `json:"transactions"`
}

type BlockDataResult struct {
	Result BlockDataRaw `json:"result"`
	Error  ErrMsg       `json:"error"`
}

type BlockData struct {
	Number   uint64
	GasLimit uint64
	GasUsed  uint64
	Time     uint64
	TxnCnt   int
}

type GethClient struct {
	httpEndpoint string
	bdCh         chan<- *BlockData
	stopCh       chan struct{}
}

func NewGethClient(ep string, bdCh chan<- *BlockData, stopc chan struct{}) *GethClient {
	return &GethClient{ep, bdCh, stopc}
}

func (g *GethClient) Start() {
	go g.readBlocksFromChain()
}

func (g *GethClient) Stop() {
	g.stopCh <- struct{}{}
}

func (g *GethClient) readBlocksFromChain() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	prevBlkNum, err := g.currentBlockNumber()
	if err != nil {
		log.Errorf("reading block number failed %v\n", err)
		return
	}

	log.Debugf("prev blockNumber=%d\n", prevBlkNum)

	totTxns := 0
	blkCount := 0
	for {
		select {
		case <-ticker.C:
			curBlkNum, err := g.currentBlockNumber()
			if err != nil {
				log.Errorf("reading block number failed %v\n", err)
				g.bdCh <- nil
				return
			}

			log.Debugf("current blockNumber=%d\n", prevBlkNum-1)

			for prevBlkNum <= curBlkNum {
				log.Debugf("reading block %d\n", prevBlkNum)
				bd, err := g.GetBlock(prevBlkNum)
				if err != nil {
					log.Errorf("reading block number %d failed with error %v", prevBlkNum, err)
					if err != blockNotFound {
						g.bdCh <- nil
						return
					}
				}
				blkCount++
				totTxns += bd.TxnCnt
				log.Debugf("totalBlocks:%d totTxns:%d\n", blkCount, totTxns)
				g.bdCh <- bd
				prevBlkNum++
			}
		case <-g.stopCh:
			g.bdCh <- nil
			log.Infof("exiting block reader")
			return
		}

	}
}

func (g *GethClient) currentBlockNumber() (uint64, error) {
	response, err := g.PostRequest(blockNumberRequest)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	var bnr BlockNumberResult
	if err := json.NewDecoder(response.Body).Decode(&bnr); err != nil {
		log.Errorf("blocknumber post request - error reading json %v\n", err)
		return 0, err
	}

	log.Debugf("BlockNumberResult:%v\n", bnr)

	blockNumber, err := hex2uint64(bnr.Result)
	if err != nil {
		log.Errorf("converting hex to int blocknumber failed %v", err)
		return 0, err
	}

	return blockNumber, nil
}

func (g *GethClient) PostRequest(blockNumberReq string) (*http.Response, error) {
	response, err := http.Post(g.httpEndpoint, "application/json", bytes.NewBuffer([]byte(blockNumberReq)))
	if err != nil {
		log.Errorf("blocknumber post request failed %v\n", err)
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("post request response failed %v\n", response.StatusCode)
		log.Errorf("err: %v\n", err)
		return nil, err
	}
	return response, nil
}

func (g *GethClient) GetBlock(bn uint64) (*BlockData, error) {
	response, err := g.PostRequest(fmt.Sprintf(blockRequest, bn))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var bdr BlockDataResult
	if err := json.NewDecoder(response.Body).Decode(&bdr); err != nil {
		log.Errorf("blockData post request - error reading json %v\n", err)
		return nil, err
	}
	log.Debugf("blockDataResult:%v\n", bdr)
	return blockResult2data(bdr.Result)
}

func (b BlockData) String() string {
	return fmt.Sprintf("block{ number:%d, txns:%d, time:%d, gasLimit:%d, gasUsed:%d}", b.Number, b.TxnCnt, b.Time, b.GasLimit, b.GasUsed)
}

func blockResult2data(r BlockDataRaw) (*BlockData, error) {
	var bd BlockData
	var err error
	// block data is empty
	if r.Number == "" {
		return nil, blockNotFound
	}
	bd.GasLimit, err = hex2uint64(r.GasLimit)
	if err != nil {
		log.Errorf("converting hex to int blockData gasLimit failed %v", err)
		return nil, err
	}
	bd.GasUsed, err = hex2uint64(r.GasUsed)
	if err != nil {
		log.Errorf("converting hex to int blockData gasUsed failed %v", err)
		return nil, err
	}

	bd.Time, err = hex2uint64(r.Time)
	if err != nil {
		log.Errorf("converting hex to int blockData Time failed %v", err)
		return nil, err
	}

	bd.Number, err = hex2uint64(r.Number)
	if err != nil {
		log.Errorf("converting hex to int blockData Number failed %v", err)
		return nil, err
	}
	bd.TxnCnt = len(r.Transactions)
	return &bd, nil
}

func hex2uint64(hexPrefixedNum string) (uint64, error) {
	return strconv.ParseUint(hexPrefixedNum, 0, 64)
}
