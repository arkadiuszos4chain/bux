package bux

import (
	"fmt"
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
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getTxReadyToSpend, 500, 1)

	testBeef.transactions = []*Transaction{testTx} // no parent

	printOut(parentTx, testTx, testBeef)
}

func ExampleBeefWithoutCmp() {
	// NOTICE! there is need to change encoding implementation to get BEEF without parents
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getTxReadyToSpend, 500, 1)

	testBeef.compoundMerklePaths = nil //no cmp

	printOut(parentTx, testTx, testBeef)
}

func ExampleBeefWithEmptyCmp() {
	// NOTICE! there is need to change encoding implementation to get BEEF without parents
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getTxReadyToSpend, 500, 1)

	testBeef.compoundMerklePaths = CMPSlice{} // empty cmp

	printOut(parentTx, testTx, testBeef)
}
