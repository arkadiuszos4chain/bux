package bux

import (
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

func ExampleAlreadySpendedBeef() {
	printFalseBeef(getAlreadySpendedTx, 500, 1)
}

func ExampleSomeoneElseUtxos() {
	printFalseBeef(getSomeoneElseTx, 500, 1)
}

func ExampleTooMuchSatoshis() {
	printFalseBeef(getTxReadyToSpend, 5000, 1)
}

func ExampleWithLockTime() {
	withLockTime := func(tx *bt.Tx) {
		tx.LockTime = 99999
	}

	printFalseBeefWithOptions(getTxReadyToSpend, 500, 1, []func(*bt.Tx){withLockTime})
}

func ExampleInputsWithLockTimeAndSequence() {
	withLockTime := func(tx *bt.Tx) {
		tx.LockTime = 99999
	}

	withSequence := func(tx *bt.Tx) {
		for _, i := range tx.Inputs {
			i.SequenceNumber = 9999
		}
	}

	printFalseBeefWithOptions(getTxReadyToSpend, 500, 1, []func(*bt.Tx){withLockTime, withSequence})
}

func printFalseBeef(getParentTx func() *Transaction, satoshis uint64, utxoIdx uint32) {
	printFalseBeefWithOptions(getParentTx, satoshis, utxoIdx, nil)
}

func printFalseBeefWithOptions(getParentTx func() *Transaction, satoshis uint64, utxoIdx uint32, btOpts []func(*bt.Tx)) {
	pubKey := "TODO: pubkey"
	toPaymail := "false_tx_test_helper@bux-wallet1.4chain.dev"
	validParentTx := getParentTx()

	fmt.Println("Inputs parent tx:")
	printTx(validParentTx)

	falseTx := prepareTxToTest(pubKey, toPaymail, validParentTx, satoshis, utxoIdx, btOpts)

	fmt.Println("Tx:")
	printTx(falseTx)

	beef := &beefTx{
		version:             1,
		compoundMerklePaths: falseTx.draftTransaction.CompoundMerklePathes,
		transactions:        []*Transaction{validParentTx, falseTx},
	}

	beefBytes, err := beef.toBeefBytes()
	if err != nil {
		panic(err)
	}

	beffHex := hex.EncodeToString(beefBytes)
	fmt.Println("BEEF:")
	fmt.Println(beffHex)
	fmt.Println()
}
