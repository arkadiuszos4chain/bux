package bux

import (
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

const _xpriv string = ""
const _pubKey string = "TODO"
const _paymailFrom string = ""

// "interface"
func prepareTestData(destination *Destination, getParentTx func() *Transaction, satoshis uint64, utxoIdx uint32) (inputParentTx, testTx *Transaction, testTxBeefData *beefTx) {
	return prepareTestDataWithOptions(destination, getParentTx, satoshis, utxoIdx, nil)
}

func prepareTestDataWithOptions(destination *Destination, getParentTx func() *Transaction, satoshis uint64, utxoIdx uint32, btOpts []func(*bt.Tx)) (inputParentTx, testTx *Transaction, testTxBeefData *beefTx) {
	xpr, prErr := bitcoin.GenerateHDKeyFromString(_xpriv)
	if prErr != nil {
		panic(prErr)
	}
	pubKey, xperr := bitcoin.GetExtendedPublicKey(xpr)
	if xperr != nil {
		panic(xperr)
	}
	inputParentTx = getParentTx()
	testTx = _prepareTxToTest(pubKey, inputParentTx, destination, satoshis, utxoIdx, btOpts)

	btParentTx := bux2btTxConvert(inputParentTx)
	btTestTx := bux2btTxConvert(testTx)

	testTxBeefData = &beefTx{
		version:      1,
		bumps:        testTx.draftTransaction.BUMPs,
		transactions: []*bt.Tx{btParentTx, btTestTx},
	}

	return
}

func bux2btTxConvert(tx *Transaction) *bt.Tx {
	var btTx *bt.Tx
	var err error

	btTx, err = bt.NewTxFromString(tx.Hex)
	if err != nil {
		panic(err)
	}

	return btTx
}

// utils
func _prepareTxToTest(pubKey string, parentTx *Transaction, destination *Destination, satoshis uint64, utxoIdx uint32, btOpts []func(*bt.Tx)) *Transaction {
	dtBuilder := CreateDraftTxBuilder(pubKey, parentTx)

	dtBuilder.WithInput(utxoIdx)
	dtBuilder.WithOutput(destination, satoshis)
	dtBuilder.WithBtOpts(btOpts)

	falseDraft := dtBuilder.Build()

	xpriv, err := bitcoin.GenerateHDKeyFromString(_xpriv)
	if err != nil {
		panic(err)
	}

	signedHex, err := _signInputs(falseDraft, xpriv)
	if err != nil {
		panic(err)
	}

	falseTx := newTransaction(signedHex)
	falseTx.draftTransaction = falseDraft

	return falseTx
}

func _signInputs(dt *DraftTransaction, xPriv *bip32.ExtendedKey) (signedHex string, resError error) {
	var err error
	// Start a bt draft transaction
	var txDraft *bt.Tx
	if txDraft, err = bt.NewTxFromString(dt.Hex); err != nil {
		resError = err
		return
	}

	// Sign the inputs
	for index, input := range dt.Configuration.Inputs {

		// Get the locking script
		var ls *bscript.Script
		if ls, err = bscript.NewFromHexString(
			input.Destination.LockingScript,
		); err != nil {
			resError = err
			return
		}
		txDraft.Inputs[index].PreviousTxScript = ls
		txDraft.Inputs[index].PreviousTxSatoshis = input.Satoshis

		// Derive the child key (chain)
		var chainKey *bip32.ExtendedKey
		if chainKey, err = xPriv.Child(
			input.Destination.Chain,
		); err != nil {
			resError = err
			return
		}

		// Derive the child key (num)
		var numKey *bip32.ExtendedKey
		if numKey, err = chainKey.Child(
			input.Destination.Num,
		); err != nil {
			resError = err
			return
		}

		// Get the private key
		var privateKey *bec.PrivateKey
		if privateKey, err = bitcoin.GetPrivateKeyFromHDKey(
			numKey,
		); err != nil {
			resError = err
			return
		}

		// Get the unlocking script
		var s *bscript.Script
		if s, err = _getUnlockingScript(
			txDraft, uint32(index), privateKey,
		); err != nil {
			resError = err
			return
		}

		// Insert the locking script
		if err = txDraft.InsertInputUnlockingScript(
			uint32(index), s,
		); err != nil {
			resError = err
			return
		}
	}

	// Return the signed hex
	signedHex = txDraft.String()
	return
}

// GetUnlockingScript will generate an unlocking script
func _getUnlockingScript(tx *bt.Tx, inputIndex uint32, privateKey *bec.PrivateKey) (*bscript.Script, error) {
	sigHashFlags := sighash.AllForkID

	sigHash, err := tx.CalcInputSignatureHash(inputIndex, sigHashFlags)
	if err != nil {
		return nil, err
	}

	var sig *bec.Signature
	if sig, err = privateKey.Sign(sigHash); err != nil {
		return nil, err
	}

	pubKey := privateKey.PubKey().SerialiseCompressed()
	signature := sig.Serialise()

	var script *bscript.Script
	if script, err = bscript.NewP2PKHUnlockingScript(pubKey, signature, sigHashFlags); err != nil {
		return nil, err
	}

	return script, nil
}
