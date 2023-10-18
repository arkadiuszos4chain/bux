package bux

func ExampleRawTx() string {
	rawTx := getTxReadyToSpend().Hex
	return rawTx
}

// func ExampleEfTx() {
// 	rawTx := getTxReadyToSpend().Hex
// 	efTx := convertToEfTransaction(rawTx)
// 	fmt.Println(efTx)
// }

func ExampleBeefWithoutParents() string {
	// NOTICE! there is need to change encoding implementation to get BEEF without parents
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getTxReadyToSpend, 500, 1)

	testBeef.transactions = []*Transaction{testTx} // no parent

	return printOut(parentTx, testTx, testBeef)
}

func ExampleBeefWithoutCmp() string {
	// NOTICE! there is need to change encoding implementation to get BEEF without parents
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getTxReadyToSpend, 500, 1)

	testBeef.compoundMerklePaths = nil //no cmp

	return printOut(parentTx, testTx, testBeef)
}

func ExampleBeefWithEmptyCmp() string {
	// NOTICE! there is need to change encoding implementation to get BEEF without parents
	destination := getTestDestination()

	parentTx, testTx, testBeef := prepareTestData(destination, getTxReadyToSpend, 500, 1)

	testBeef.compoundMerklePaths = CMPSlice{} // empty cmp

	return printOut(parentTx, testTx, testBeef)
}
