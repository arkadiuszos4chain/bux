package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleInputsWithLockTimeAndSequence()

	fmt.Println(s)

	bux.WriteToFile("with_lock_time_n_sequence.txt", "with_lock_time_n_sequence", s)
}
