package bux

import (
	"encoding/hex"
	"fmt"

	//"github.com/GorillaPool/go-junglebus"

	"github.com/libsv/go-bt/v2"
)

func ExampleRawTx() {
	rawTx := getTxReadyToSpend().Hex
	fmt.Println(rawTx)
}

// func ExampleEfTx() {
// 	rawTx := getTxReadyToSpend().Hex
// 	efTx := convertToEfTransaction(rawTx)
// 	fmt.Println(efTx)
// }

func ExampleBeefWithoutParents() {
	// NOTICE! there is need to change encoding implementation to get BEEF without parents
	getInvalidBeefWithoutParents(getTxReadyToSpend, 500, 1, nil)
}

func ExampleBeefWithoutCmp() {
	// NOTICE! there is need to change encoding implementation to get BEEF without parents
	getInvalidBeefWithoutCmp(getTxReadyToSpend, 500, 1, nil)
}

func ExampleBeefWithEmptyCmp() {
	// NOTICE! there is need to change encoding implementation to get BEEF without parents
	getInvalidBeefWithEmptyCmp(getTxReadyToSpend, 500, 1, nil)
}

// func convertToEfTransaction(rawTx string) string {
// 	junglebusClient, err := junglebus.New(
// 		junglebus.WithHTTP("https://junglebus.gorillapool.io"),
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	transaction, err := bt.NewTxFromString(rawTx)
// 	if err != nil {
// 		panic(err)
// 	}

// 	for _, input := range transaction.Inputs {
// 		if err = updateUtxoWithMissingData(junglebusClient, input); err != nil {
// 			panic(err)
// 		}
// 	}

// 	return hex.EncodeToString(transaction.ExtendedBytes())

// }

// func updateUtxoWithMissingData(jbc *junglebus.Client, input *bt.Input) error {
// 	txid := input.PreviousTxIDStr()

// 	tx, err := jbc.GetTransaction(context.Background(), txid)
// 	if err != nil {
// 		return err
// 	}

// 	actualTx, err := bt.NewTxFromBytes(tx.Transaction)
// 	if err != nil {
// 		return err
// 	}

// 	o := actualTx.Outputs[input.PreviousTxOutIndex]
// 	input.PreviousTxScript = o.LockingScript
// 	input.PreviousTxSatoshis = o.Satoshis
// 	return nil
// }

func getInvalidBeefWithoutParents(getParentTx func() *Transaction, satoshis uint64, utxoIdx uint32, btOpts []func(*bt.Tx)) {
	pubKey := "TODO: pubkey"
	toPaymail := "TODO: paymail"
	validParentTx := getParentTx()

	printTx(validParentTx)

	falseTx := prepareTxToTest(pubKey, toPaymail, validParentTx, satoshis, utxoIdx, btOpts)

	printTx(falseTx)

	beef := &beefTx{
		version:             1,
		compoundMerklePaths: falseTx.draftTransaction.CompoundMerklePathes,
		transactions:        []*Transaction{falseTx}, // no parent
	}

	beefBytes, err := beef.toBeefBytes()
	if err != nil {
		panic(err)
	}

	beffHex := hex.EncodeToString(beefBytes)
	fmt.Println(beffHex)
	fmt.Println()
}

func getInvalidBeefWithoutCmp(getParentTx func() *Transaction, satoshis uint64, utxoIdx uint32, btOpts []func(*bt.Tx)) {
	pubKey := "TODO: pubkey"
	toPaymail := "TODO: paymail"
	validParentTx := getParentTx()

	printTx(validParentTx)

	falseTx := prepareTxToTest(pubKey, toPaymail, validParentTx, satoshis, utxoIdx, btOpts)

	printTx(falseTx)

	beef := &beefTx{
		version:             1,
		compoundMerklePaths: nil, //no cmp,
		transactions:        []*Transaction{validParentTx, falseTx},
	}

	beefBytes, err := beef.toBeefBytes()
	if err != nil {
		panic(err)
	}

	beffHex := hex.EncodeToString(beefBytes)
	fmt.Println(beffHex)
	fmt.Println()
}

func getInvalidBeefWithEmptyCmp(getParentTx func() *Transaction, satoshis uint64, utxoIdx uint32, btOpts []func(*bt.Tx)) {
	pubKey := "TODO: pubkey"
	toPaymail := "TODO: paymail"
	validParentTx := getParentTx()

	printTx(validParentTx)

	falseTx := prepareTxToTest(pubKey, toPaymail, validParentTx, satoshis, utxoIdx, btOpts)

	printTx(falseTx)

	beef := &beefTx{
		version:             1,
		compoundMerklePaths: CMPSlice{}, // empty cmp
		transactions:        []*Transaction{validParentTx, falseTx},
	}

	beefBytes, err := beef.toBeefBytes()
	if err != nil {
		panic(err)
	}

	beffHex := hex.EncodeToString(beefBytes)
	fmt.Println(beffHex)
	fmt.Println()
}
