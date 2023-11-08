package main

import (
	"fmt"

	"github.com/BuxOrg/bux"
)

func main() {
	s := bux.ExampleBeefWithEmptyBumps()

	fmt.Println(s)

	bux.WriteToFile("with_empty_bumps.txt", "with_empty_bumps", s)
}
