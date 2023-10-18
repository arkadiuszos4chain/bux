package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleBeefWithEmptyCmp()

	fmt.Println(s)

	bux.WriteToFile("with_empty_cmp.txt", "with_empty_cmp", s)
}
