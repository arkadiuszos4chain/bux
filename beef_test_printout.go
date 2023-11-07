package bux

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/libsv/go-bt/v2"
)

func WriteToFile(path, exampleName, content string) {
	f, err := os.Create(path)
	defer f.Close()

	if err != nil {
		panic(err)
	}

	header := fmt.Sprintf("generated with %s example", exampleName)
	f.WriteString(fmt.Sprintf("%s\n** %44s **\n%s\n", strings.Repeat("*", 50), header, strings.Repeat("*", 50)))
	f.WriteString(content)
}

func printOut(inputParentTx, testTx *Transaction, beefData *beefTx) string {
	var b strings.Builder

	b.WriteString("Inputs parent tx:\n")
	_printTx(inputParentTx, &b)

	b.WriteString("Test Tx:\n")
	_printTx(testTx, &b)

	b.WriteString("Test BEEF:\n")
	_printBeefJson(beefData, &b)

	beefBytes, err := beefData.toBeefBytes()
	if err != nil {
		panic(err)
	}

	b.WriteString("BEEF hex:\n")
	b.WriteString(hex.EncodeToString(beefBytes))
	b.WriteString("\n")

	return b.String()
}

func _printTx(tx *Transaction, b *strings.Builder) {
	b.WriteString("Raw transaction info:\n")
	btx := bux2btTxConvert(tx)
	_prettyPrint(btx, b)

	b.WriteString("Bux info:\n")
	b.WriteString("Transaction:\n")
	_prettyPrint(tx, b)
	b.WriteString("Draft:\n")
	_prettyPrint(tx.draftTransaction, b)

	b.WriteString("\n")
	b.WriteString("===============\n")
}

func _printBeefJson(bf *beefTx, b *strings.Builder) {
	js := _beefTxJs{
		Version:      bf.version,
		Bumps:        bf.bumps,
		Transactions: bf.transactions,
	}

	_prettyPrint(js, b)
}

func _prettyPrint(v interface{}, b *strings.Builder) {
	vJson, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}

	if len(vJson) == 0 {
		b.WriteString("nil\n")
	} else {
		b.Write(vJson)
		b.WriteString("\n")
	}
}

type _beefTxJs struct {
	Version      uint32   `json:"version"`
	Bumps        BUMPs    `json:"bumps"`
	Transactions []*bt.Tx `json:"transactions"`
}
