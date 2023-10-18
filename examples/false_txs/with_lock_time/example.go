package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleWithLockTime()

	fmt.Println(s)

	bux.WriteToFile("with_lock_time.txt", "with_lock_time", s)
}
