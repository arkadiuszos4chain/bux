package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleBeefWithoutCmp()

	fmt.Println(s)

	bux.WriteToFile("without_cmp.txt", "without_cmp", s)
}
