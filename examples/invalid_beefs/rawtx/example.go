package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleRawTx()

	fmt.Println(s)

	bux.WriteToFile("rawtx.txt", "rawtx", s)
}
