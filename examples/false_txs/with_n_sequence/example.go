package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleInputsWithSequence()

	fmt.Println(s)

	bux.WriteToFile("with_n_sequence.txt", "with_n_sequence", s)
}
