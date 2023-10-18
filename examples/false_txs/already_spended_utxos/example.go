package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleAlreadySpendedBeef()

	fmt.Println(s)

	bux.WriteToFile("already_spended_utxos.txt", "already_spended_utxos", s)
}
