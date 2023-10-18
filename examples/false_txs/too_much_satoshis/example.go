package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleTooMuchSatoshis()

	fmt.Println(s)

	bux.WriteToFile("too_much_satoshis.txt", "too_much_satoshis", s)
}
