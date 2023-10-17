package bux

import (
	"github.com/libsv/go-bt/v2"
)

type _falseDraftTxBuilder struct {
	pubKey      string
	paymailFrom string

	parentTx *Transaction

	inputs  []*Utxo
	outputs []*_output

	btOpts []func(*bt.Tx)
}

type _output struct {
	satoshis uint64
	to       *Destination
}

func CreateDraftTxBuilder(pubKey string, parentTx *Transaction) *_falseDraftTxBuilder {
	paymailFrom := _paymailFrom

	if parentTx.MerkleProof.TxOrID == "" {
		panic("not minded parent tx")
	}

	return &_falseDraftTxBuilder{
		pubKey:      pubKey,
		paymailFrom: paymailFrom,
		parentTx:    parentTx,

		inputs:  make([]*Utxo, 0),
		outputs: make([]*_output, 0),
		btOpts:  make([]func(*bt.Tx), 0),
	}
}

func (b *_falseDraftTxBuilder) WithInput(utxoIdx uint32) {
	output := b.parentTx.parsedTx.Outputs[utxoIdx]

	utxo := newUtxo(b.pubKey,
		b.parentTx.ID,
		output.LockingScript.String(),
		utxoIdx,
		output.Satoshis,
	)

	b.inputs = append(b.inputs, utxo)
}

func (b *_falseDraftTxBuilder) WithOutput(destination *Destination, satoshis uint64) {
	output := _output{
		satoshis: satoshis,
		to:       destination,
	}

	b.outputs = append(b.outputs, &output)
}

func (b *_falseDraftTxBuilder) WithBtOpts(opts []func(*bt.Tx)) {
	b.btOpts = append(b.btOpts, opts...)
}

func (b *_falseDraftTxBuilder) Build() *DraftTransaction {
	falseConfig := b._createConfig()

	falseDraft := newDraftTransaction(b.pubKey, falseConfig)
	falseDraft.CompoundMerklePathes = b._calculateCompoundMerklePath()
	falseDraft.Hex = b._generateHex(falseDraft)

	return falseDraft
}

func (b *_falseDraftTxBuilder) _createConfig() *TransactionConfig {
	inputs := make([]*TransactionInput, 0, len(b.inputs))
	for _, i := range b.inputs {
		inputs = append(inputs, &TransactionInput{Utxo: *i})
	}

	outputs := b._convertOutputsToConfigOutputs()

	falseConfig := TransactionConfig{
		Inputs:  inputs,
		Outputs: outputs,
	}

	return &falseConfig
}

func (b *_falseDraftTxBuilder) _convertOutputsToConfigOutputs() []*TransactionOutput {
	results := make([]*TransactionOutput, 0, len(b.outputs))

	for _, o := range b.outputs {
		output := &TransactionOutput{
			Satoshis: o.satoshis,
			To:       o.to.Address,
		}

		// Start the output script slice
		if output.Scripts == nil {
			output.Scripts = make([]*ScriptOutput, 0)
		}

		if err := output._processOutput(
			b.paymailFrom,
			"BEEF testing",
			true,
			o.to,
		); err != nil {
			panic(err)
		}

		results = append(results, output)
	}

	return results
}

func (b *_falseDraftTxBuilder) _calculateCompoundMerklePath() CMPSlice {
	merkleProofs := make(map[uint64][]MerkleProof)
	merkleProofs[b.parentTx.BlockHeight] = append(merkleProofs[b.parentTx.BlockHeight], b.parentTx.MerkleProof)

	cmps := make(CMPSlice, 0)
	for _, v := range merkleProofs {
		cmp, err := CalculateCompoundMerklePath(v)
		if err != nil {
			panic(err)
		}
		cmps = append(cmps, cmp)
	}

	return cmps
}

func (b *_falseDraftTxBuilder) _generateHex(falseDraft *DraftTransaction) string {
	inputUtxos, _, err := falseDraft.getInputsFromUtxos(b.inputs)
	if err != nil {
		panic(err)
	}

	tx := bt.NewTx()
	if err := tx.FromUTXOs(*inputUtxos...); err != nil {
		panic(err)
	}

	if err = falseDraft.addOutputsToTx(tx); err != nil {
		panic(err)
	}

	for _, opt := range b.btOpts {
		opt(tx)
	}

	return tx.String()
}
