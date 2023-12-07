package bux

import (
	"context"
	"fmt"

	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bt/v2"
)

func Example_Unmined() {
	store := &tstore{}
	store.d = make([]*Transaction, 0)

	// mined transactions
	mt_1 := getTxReadyToSpend()  // 1000 satoshis
	mt_2 := getTxReadyToSpend1() // 1000 satoshis
	mt_3 := getTxReadyToSpend2() // 500 satoshis
	mt_4 := getTxReadyToSpend3() // 500 satoshis

	store.add(mt_1)
	store.add(mt_2)
	store.add(mt_3)
	store.add(mt_4)

	// [mt_1, mt_2] => ut_1
	ut_1 := createTxF(1500,
		txInput{mt_1, 0, &Destination{Num: 13}},
		txInput{mt_2, 0, &Destination{Num: 18}},
	) // change  500

	store.add(ut_1)
	//fmt.Println(ut_1)

	// [ut_1] => ut_2
	ut_2 := createTxF(100, txInput{ut_1, 1, &Destination{Num: 33}}) // change 400
	store.add(ut_2)

	// [ut_2] => ut_3
	ut_3 := createTxF(100, txInput{ut_2, 1, &Destination{Num: 33}}) // change 300
	store.add(ut_3)

	// [mt_3] => ut_4
	ut_4 := createTxF(150, txInput{mt_3, 0, &Destination{Num: 22}}) // change 350
	store.add(ut_4)

	// [ut_3 (change), ut_4(change), mt_4] => resutl tx

	// valid
	result_to_test_tx := createTxF(1100,
		txInput{ut_3, 1, &Destination{Num: 33}},
		txInput{ut_4, 1, &Destination{Num: 33}},
		txInput{mt_4, 0, &Destination{Num: 27}},
	)

	// // spend too much satoshis
	// result_to_test_tx = createTxF(5100,
	// 	txInput{ut_3, 1, &Destination{Num: 33}},
	// 	txInput{ut_4, 1, &Destination{Num: 33}},
	// 	txInput{mt_4, 0, &Destination{Num: 27}},
	// )

	// // spend someone else utxo
	// result_to_test_tx = createTxF(800,
	// 	txInput{ut_3, 1, &Destination{Num: 33}},
	// 	txInput{ut_4, 0, &Destination{Num: 33}},
	// 	txInput{mt_4, 0, &Destination{Num: 27}},
	// )

	beef, err := ToBeef(context.Background(), result_to_test_tx, store)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(beef)
	}
}

func createTxF(satoshis uint64, inputs ...txInput) *Transaction {
	from := &Destination{
		Type:          "pubkeyhash",
		Address:       "false_tx_test@bux-wallet1.4chain.dev",
		Chain:         0,
		LockingScript: "76a9148360b65a14e83a68aef2fae3f9cf472a90dd4da488ac",
		Num:           33,
	}

	to := &Destination{
		Type:          "pubkeyhash",
		Address:       "false_tx_test_helper@bux-wallet1.4chain.dev",
		Chain:         0,
		LockingScript: "76a914ddfbaa2cd75b86cf24135603b5ddd2b1968bc62d88ac",
		Num:           142,
	}

	b := &_txbuilder{
		pubKey:   _pubKey,
		xpriv:    _xpriv,
		from:     from,
		to:       to,
		satoshis: satoshis,
		inputs:   inputs,
	}

	return b.Build()
}

type txInput struct {
	tx        *Transaction
	outputIdx uint32

	from *Destination
}

// repo/store
type tstore struct {
	d []*Transaction
}

func (s *tstore) add(t *Transaction) {
	s.d = append(s.d, t)
}
func (s *tstore) GetTransactionsByIDs(ctx context.Context, txIDs []string) ([]*Transaction, error) {
	res := make([]*Transaction, 0)

	for _, id := range txIDs {

		for _, tx := range s.d {
			if tx.ID == id {
				res = append(res, tx)

				break
			}

		}

	}

	return res, nil
}

// builders

type _txbuilder struct {
	pubKey, xpriv string

	from, to *Destination

	satoshis uint64
	inputs   []txInput
}

func (b *_txbuilder) Build() *Transaction {
	draft := b._createDraft()
	return b._createTransaction(draft)
}

func (b *_txbuilder) _createDraft() *DraftTransaction {
	draftB := _draftBuilder{
		pubKey:  b.pubKey,
		from:    b.from,
		inputs:  make([]*TransactionInput, 0),
		outputs: make([]*TransactionOutput, 0),
	}

	for _, i := range b.inputs {
		draftB.WithInput(&i)
	}

	draftB.WithOutputTo(b.satoshis, b.to)

	draft := draftB.Build()
	return draft
}

func (b *_txbuilder) _createTransaction(draft *DraftTransaction) *Transaction {
	xpriv, err := bitcoin.GenerateHDKeyFromString(b.xpriv)
	if err != nil {
		panic(err)
	}

	signedHex, err := _signInputs(draft, xpriv)
	if err != nil {
		panic(err)
	}

	tx := newTransaction(signedHex)
	tx.draftTransaction = draft

	return tx
}

type _draftBuilder struct {
	pubKey string
	from   *Destination

	inputs    []*TransactionInput
	inputSato uint64

	outputs    []*TransactionOutput
	outputSato uint64
}

func (b *_draftBuilder) WithInput(i *txInput) {
	prevTxOut := i.tx.parsedTx.Outputs[i.outputIdx]

	outReceiverDst := Destination{
		Chain:         i.from.Chain,
		Num:           i.from.Num,
		LockingScript: prevTxOut.LockingScript.String(),
	}

	utxo := newUtxo(
		b.pubKey, // chyba niepotrzebne
		i.tx.ID,
		prevTxOut.LockingScript.String(),
		i.outputIdx,
		prevTxOut.Satoshis,
	)
	//fmt.Println(output.LockingScript.String())

	b.inputs = append(b.inputs, &TransactionInput{Utxo: *utxo, Destination: outReceiverDst})

	b.inputSato += prevTxOut.Satoshis
}

func (b *_draftBuilder) WithOutputTo(satoshis uint64, to *Destination) {

	out := &TransactionOutput{
		Satoshis: satoshis,
		To:       to.Address,
		Scripts:  make([]*ScriptOutput, 0),
	}

	if err := out._processOutput(
		b.from.Address,
		"BEEF testing",
		true,
		to,
	); err != nil {
		panic(err)
	}

	b.outputs = append(b.outputs, out)
	b.outputSato += out.Satoshis

}

func (b *_draftBuilder) Build() *DraftTransaction {
	if b.inputSato > b.outputSato {
		b._addChangeOutput()
	}

	config := &TransactionConfig{
		Inputs:  b.inputs,
		Outputs: b.outputs,
	}

	draft := newDraftTransaction(b.pubKey, config)
	draft.Hex = b._generateHex(draft)

	return draft
}

func (b *_draftBuilder) _addChangeOutput() {
	changeOutput := &TransactionOutput{
		Satoshis:     b.inputSato - b.outputSato - 1, // fee,
		To:           b.from.Address,
		Scripts:      make([]*ScriptOutput, 0),
		UseForChange: true,
	}

	if err := changeOutput._processOutput(
		b.from.Address,
		"BEEF testing",
		true,
		b.from,
	); err != nil {
		panic(err)
	}

	b.outputs = append(b.outputs, changeOutput)
}

func (b *_draftBuilder) _generateHex(draft *DraftTransaction) string {
	reserverdUtxos := make([]*Utxo, 0)

	for _, i := range draft.Configuration.Inputs {
		reserverdUtxos = append(reserverdUtxos, &i.Utxo)
	}

	inputUtxos, _, err := draft.getInputsFromUtxos(reserverdUtxos)
	if err != nil {
		panic(err)
	}

	tx := bt.NewTx()
	if err := tx.FromUTXOs(*inputUtxos...); err != nil {
		panic(err)
	}

	if err = draft.addOutputsToTx(tx); err != nil {
		panic(err)
	}

	// for _, opt := range b.btOpts {
	// 	opt(tx)
	// }

	return tx.String()
}
