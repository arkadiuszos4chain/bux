package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleBeefWithoutParents()

	fmt.Println(s)

	bux.WriteToFile("without_parents.txt", "without_parents", s)
}
