package bux

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

func printOut(inputParentTx, testTx *Transaction, beefData *beefTx) {
	fmt.Println("Inputs parent tx:")
	_printTx(inputParentTx)

	fmt.Println("Test Tx:")
	_printTx(testTx)

	fmt.Println("Test BEEF:")
	_printBeefJson(beefData)

	beefBytes, err := beefData.toBeefBytes()
	if err != nil {
		panic(err)
	}

	fmt.Println("BEEF hex:")
	fmt.Println(hex.EncodeToString(beefBytes))
	fmt.Println()

}

func _printTx(tx *Transaction) {
	fmt.Println("Print out tx:")
	fmt.Println()

	fmt.Println("Raw transaction info:")
	btx, _ := bt.NewTxFromString(tx.Hex)
	_prettyPrint(btx)

	fmt.Println("Bux info:")
	_prettyPrint(tx)
	_prettyPrint(tx.draftTransaction)

	fmt.Println()
	fmt.Println("===============")
}

func _printBeefJson(bf *beefTx) {
	js := _beefTxJs{
		Version:             bf.version,
		CompoundMerklePaths: bf.compoundMerklePaths,
		Transactions:        bf.transactions,
	}

	_prettyPrint(js)
}

func _prettyPrint(v interface{}) {
	vJson, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}

	fmt.Printf("%s\n", vJson)
}

type _beefTxJs struct {
	Version             uint32         `json:"version"`
	CompoundMerklePaths CMPSlice       `json:"compoundMerklePaths"`
	Transactions        []*Transaction `json:"transactions"`
}
