package bux

import (
	"github.com/libsv/go-bt/v2"
)

func ExampleAlreadySpendedBeef() string {
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getAlreadySpendedTx, 200, 0)

	return printOut(parentTx, testTx, testBeef)
}

func ExampleSomeoneElseUtxos() string {
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getSomeoneElseTx, 200, 0)
	return printOut(parentTx, testTx, testBeef)
}

func ExampleTooMuchSatoshis() string {
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getTxReadyToSpend, 5000, 0)
	return printOut(parentTx, testTx, testBeef)
}

func ExampleWithLockTime() string {
	withLockTime := func(tx *bt.Tx) {
		tx.LockTime = 99999
	}

	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestDataWithOptions(destination, getTxReadyToSpend, 200, 0, []func(*bt.Tx){withLockTime})

	return printOut(parentTx, testTx, testBeef)
}

func ExampleInputsWithLockTimeAndSequence() string {
	withLockTime := func(tx *bt.Tx) {
		tx.LockTime = 99999
	}

	withSequence := func(tx *bt.Tx) {
		for _, i := range tx.Inputs {
			i.SequenceNumber = 9999
		}
	}

	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestDataWithOptions(destination, getTxReadyToSpend, 200, 0, []func(*bt.Tx){withLockTime, withSequence})

	return printOut(parentTx, testTx, testBeef)
}

func ExampleInputsWithSequence() string {
	withSequence := func(tx *bt.Tx) {
		for _, i := range tx.Inputs {
			i.SequenceNumber = 9999
		}
	}

	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestDataWithOptions(destination, getTxReadyToSpend, 200, 0, []func(*bt.Tx){withSequence})

	return printOut(parentTx, testTx, testBeef)
}

func ExampleValidBumpFromOtherTx() string {
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getTxReadyToSpend, 200, 0)
	_, _, testBeef2 := prepareTestData(destination, getSomeoneElseTx, 200, 0)

	testBeef.bumps = testBeef2.bumps

	return printOut(parentTx, testTx, testBeef)
}
