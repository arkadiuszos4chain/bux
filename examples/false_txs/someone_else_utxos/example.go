package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleSomeoneElseUtxos()

	fmt.Println(s)

	bux.WriteToFile("someone_else_utxos.txt", "someone_else_utxos", s)
}
