package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleValidBumpFromOtherTx()

	fmt.Println(s)

	bux.WriteToFile("incorrect_bump.txt", "incorrect_bump", s)
}
