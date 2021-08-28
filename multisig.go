package main

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"

	//"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"io"
	"os"
	"runtime/debug"

	"github.com/btcsuite/btcd/rpcclient"
	//"github.com/ethereum/go-ethereum/rpc"
	"github.com/fatih/color"
)

var (
	configFile = "config.json"
)

type config struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

//type Address interface {
//	String() string
//	EncodeAddress() string
//	ScriptAddress() []byte
//	IsForNet(*chaincfg.Params) bool
//}

type dogeAddress struct {
	addr string
}



func main() {
	fmt.Println("Started")

	cfg, err := loadCfg()
	if err != nil {
		printErrMsg(err)
		return
	}

	// Connect to Dogecoin node and auth with username and password
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
		Host:         "127.0.0.1:44555",
		User:         cfg.User,
		Pass:         cfg.Pass,
	}, nil)
	if err != nil {
		printErrMsg(err)
		return
	}

	// Everything else!
	//r, err := client.ListUnspent()

	addr, err := btcutil.DecodeAddress("2MsoE4GPg2Gn2LwGXUhupaRuLcBvwj6NvyW", &DogeTestNetParams)

	r, err := client.ListUnspentMinMaxAddresses(0, 999999999, []btcutil.Address{addr})

	if err != nil {
		printErrMsg(err)
		return
	}
	var target int64 = 5 * 1e8
	fmt.Println(gatherInputs(target, r), "doge")

	//client.CreateRawTransaction()
	
	// get transaction outputs
	// accumulate them into one big input slightly larger than the amount we want to send
	// create partially signed bitcoin transaction so that the other parties in the multisig address can sign too
	// broadcast finalized transaction
}

func gatherInputs(target int64, r []btcjson.ListUnspentResult) []btcjson.ListUnspentResult {
	result := make([]btcjson.ListUnspentResult, 0, 10)
	var current int64 = 0
	for i := range r {
		current += int64(r[i].Amount*1e8)
		result = append(result, r[i])
		//fmt.Println("It worked:", r[i].TxID, r[i].Vout, "acumulated:", current)
		if current > target {
			break
		}
	}
	return result
}

func printErrMsg(err error) {
	fmt.Println(color.RedString("error: "), err)
	debug.PrintStack()
}

func loadCfg() (*config, error) {
	// Load config to get username & password
	cfg := new(config)
	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
