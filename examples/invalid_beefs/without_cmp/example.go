package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleBeefWithoutBumps()

	fmt.Println(s)

	bux.WriteToFile("without_bumps.txt", "without_bumps", s)
}
