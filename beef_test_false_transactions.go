package bux

import (
	"github.com/libsv/go-bt/v2"
)

func ExampleAlreadySpendedBeef() {
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getAlreadySpendedTx, 500, 1)
	printOut(parentTx, testTx, testBeef)
}

func ExampleSomeoneElseUtxos() {
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getSomeoneElseTx, 500, 1)
	printOut(parentTx, testTx, testBeef)
}

func ExampleTooMuchSatoshis() {
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getTxReadyToSpend, 5000, 1)
	printOut(parentTx, testTx, testBeef)
}

func ExampleWithLockTime() {
	withLockTime := func(tx *bt.Tx) {
		tx.LockTime = 99999
	}

	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestDataWithOptions(destination, getTxReadyToSpend, 500, 1, []func(*bt.Tx){withLockTime})
	printOut(parentTx, testTx, testBeef)
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

	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestDataWithOptions(destination, getTxReadyToSpend, 500, 1, []func(*bt.Tx){withLockTime, withSequence})
	printOut(parentTx, testTx, testBeef)
}
