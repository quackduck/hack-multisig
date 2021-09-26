package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/fatih/color"
	"io"
	"os"
	"runtime/debug"
)

var (
	configFile = "config.json"
	address = "2MsoE4GPg2Gn2LwGXUhupaRuLcBvwj6NvyW"
)

type rpcConfig struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

func main() {
	cfg, err := loadCfg()
	if err != nil {
		printErr(err)
		return
	}

	// Connect to Dogecoin node and auth with username and password
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
		Host:         "localhost:44555",
		User:         cfg.User,
		Pass:         cfg.Pass,
	}, nil)
	if err != nil {
		printErr(err)
		return
	}

	addressFrom, err := btcutil.DecodeAddress(address, &DogeTestNetParams)

	r, err := client.ListUnspentMinMaxAddresses(0, 999999999, []btcutil.Address{addressFrom})

	if err != nil {
		printErr(err)
		return
	}
	target, err := readAmount()
	if err != nil {
		printErr(err)
		return
	}
	unspentOutputs, err := gatherInputs(target, r)
	if err != nil {
		printErr(err)
		return
	}
	fmt.Println(toJSON(unspentOutputs))

	inputs := make([]btcjson.TransactionInput, len(unspentOutputs))
	for i, v := range unspentOutputs {
		inputs[i] = btcjson.TransactionInput{
			Txid: v.TxID,
			Vout: v.Vout,
		}
	}

	outputs := make(map[btcutil.Address]btcutil.Amount)
	addr, err := readAddress()
	if err != nil {
		printErr(err)
		return
	}
	outputs[addr] = target

	locktime := int64(0)

	//fmt.Println(len(outputs))
	//fmt.Println(toJSON(inputs), outputs)
	transaction, err := client.CreatePSBTTransaction(inputs, outputs, &(locktime))
	if err != nil {
		printErr(err)
		return
	}
	fmt.Println(transaction)

	//base64decoded, err := base64.StdEncoding.DecodeString(transaction)
	//if err != nil {
	//	printErr(err)
	//	return
	//}
	s, err := client.DecodePSBT(transaction)
	if err != nil {
		printErr(err)
		return
	}
	fmt.Println(s)
	//client.CreateRawTransaction()

	// get transaction outputs
	// accumulate them into one big input slightly larger than the amount we want to send
	// create partially signed bitcoin transaction so that the other parties in the multisig address can sign too
	// broadcast finalized transaction
}

func readAmount() (btcutil.Amount, error) {
	amt := 0.0
	fmt.Print("Enter amount to send: ")
	color.Set(color.FgHiCyan)
	_, err := fmt.Scanf("%f", &amt)
	color.Unset()
	if err != nil {
		return 0, err
	}
	amount, err := btcutil.NewAmount(amt)
	if err != nil {
		return 0, err
	}

	return amount, nil
}


func readAddress() (btcutil.Address, error) {
	addr := ""
	var decodeAddress btcutil.Address
	var err error
	for true {
		color.Set(color.FgHiCyan)
		fmt.Print("Enter address to send to: ")
		_, err = fmt.Scanf("%s", &addr)
		if err != nil {
			return nil, err
		}
		color.Unset()
		if addr == "" {
			return nil, errors.New("user doesn't wanna do a transaction :(((")
		}
		decodeAddress, err = btcutil.DecodeAddress(addr, &DogeTestNetParams)
		if err == nil {
			break
		}
	}

	return decodeAddress, nil
}

func gatherInputs(target btcutil.Amount, r []btcjson.ListUnspentResult) ([]btcjson.ListUnspentResult, error) {
	if target < 1 {
		return nil, errors.New("silly human, one can't send that much money")
	}
	result := make([]btcjson.ListUnspentResult, 0, 10)
	// TODO: sort to pick as many small unspent outputs as possible?
	var current float64 = 0
	for i := range r {
		current += r[i].Amount
		result = append(result, r[i])
		//fmt.Println("It worked:", r[i].TxID, r[i].Vout, "acumulated:", current)
		if current > target.ToBTC() {
			break
		}
	}
	if current <= target.ToBTC() {
		return nil, errors.New("not enough funds")
	}
	return result, nil
}

func printErr(err error) {
	fmt.Println(color.RedString("error:"), err)
	debug.PrintStack()
}

func loadCfg() (*rpcConfig, error) {
	// Load config to get username & password
	cfg := new(rpcConfig)
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

func toJSON(v interface{}) string {
	//s, _ := json.Marshal(v)
	s, _ := json.MarshalIndent(v, "", "   ")
	return string(s)
}
