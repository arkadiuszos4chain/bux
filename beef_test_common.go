package bux

import (
	"encoding/json"
	"fmt"

	"github.com/libsv/go-bc"
	"github.com/libsv/go-bt/v2"
)

// real tx feed

func getAlreadySpendedTx() *Transaction {
	// tx sended from handcash to false_tx_test@bux-wallet1.4chain.dev, all satoshis (1500) were sended back
	// https://www.whatsonchain.com/tx/1876d646b17c443084e6b1f1e58dfe4b9ed254d6141bbf016eef04424ec753e5

	tx := newTransaction("0100000002f2d523adeea458668b95ec0dffc94e131a6d763f07c89fa30e722a9c7fb4ca40000000006a473044022065cd0d4988f2eb71f08d0f18bb190b2ae8e044147680d3357d01b9908dde9f13022076a98893c14fea55c4af5e31ddc8ae67452fc1c794bebda919070f983b02e0ed4121024d20c39408b8e79960455db4eee5e06efc47dadda035d2f88337c1466a6bb60bffffffffb711324badddc2da48f5f581d7d1fb7f496bd3a50e4f8758e82515cec12f64c4060000006a47304402207bc79231e3c01ac51dd95d9463aeecc7f17f901f5bb6331bf08850f4fd81fd6a02206a3cdb0845ad1eaaa3d4c67a2f679b652e4ab94596629e1308aa025616e0af35412102c32e3c8996fc2767c7de89a90cc54aedca10212ed1f4911e383611780c3a190cffffffff03dc050000000000001976a914ad5f6eb9b38c4cba69c1ad9bc75d0b59d10fdbed88acac0d0000000000001976a91450f444b84524c8cc5b447e66f4cdfd30f0e4e79b88ac22010000000000001976a9145ce2c2ece02412dc2f635a6db58e240e43deed3988ac00000000")
	tx.BlockHeight = 813381

	// mark it as mined
	merkleProofJson := []byte("{\"index\":10280,\"txOrId\":\"1876d646b17c443084e6b1f1e58dfe4b9ed254d6141bbf016eef04424ec753e5\",\"targetType\":\"header\",\"target\":\"00c02e2621dc74a9489c36e88f6b0701c68f1fdd13b0befc53defa0c00000000000000000c5d39be19ed061cd94efeddd1141e4566496dc374cfcd970ca72c9c573243c2049026658c0610181db21bdf\",\"nodes\":[\"06bb08a7228dce296ee0b0a1ab2e9edba160280cabe54a8c0e46d83a03fbd07d\",\"a1bdb111453f5e043a830ef676a93d948f0721285167a130802472816eeea54f\",\"b007d6ec644f01f3549da5b7753238756e4b8c77019b170a08a5e8ea830bf81f\",\"4f5f6a6b27d8dee62f672810f66dc66bf3793364824ce055a6b3dd3fcaa281b3\",\"eaa34a17a499cd6c52496b0fd61e6109d5b63fdad1fe1dd4ea33238c5bfe67b0\",\"9cee07b4c286300dceabb106a3fbf937a1b7fc1ed3a9f25918bd12414df42438\",\"1a72eafcd9aaabe5e0ff16b030fad48f2c44e9fabf447316e7580fbbc197d3dd\",\"9129136a63c1e0039426d39967e4b59c3f1aab179075c8c0d8be7274b5efb825\",\"1709b68e7511390a69dcdac9f31355824a41438892c68ba37d07a362cf71bfae\",\"cc12209103d5c11bbe43d9ac1c45694df756a21f75e39f23cc25848ed423f281\",\"bad17ead9d1892a57180b72f32a8344987b4a7647f580f9937f3e85edbf2807d\",\"094bc0d0a158ee8ca09553309fff12f82a8af0c7a6f532151b0acf279a415800\",\"1ab67f45ea2d8d9ed382ef93276f9266b8e0de803d3e1ee5fb4d33f02cb740f6\",\"999bb3f098e097a8f51ecf085cd27cc22689987404c0c8eb6ff0e1917d275fb5\",\"4d91292e37bcf7861a506b1c3a7ccfb2579017fe3bb26453e2289238de2bd582\"]}")
	var mp bc.MerkleProof
	json.Unmarshal(merkleProofJson, &mp)

	tx.MerkleProof = MerkleProof(mp)

	return tx
}

func getSomeoneElseTx() *Transaction {
	// tx sended from handcash to false_tx_test_helper@bux-wallet1.4chain.dev (1000 satoshis)
	// https://www.whatsonchain.com/tx/43dd9cb05fbcf4bbc3031b5f3b21e429800ac99aaac351c36a169862c8734270

	tx := newTransaction("010000000236f45e6fbf3feff044c5d83315c3032118e2e2bc8cb51841026b60362903ad6b010000006a4730440220443da8827fb2914f4468c3051b2601ab300b6d0f6f9f1d3aff5b410dc0a750ba022079581b51ae6aecf58c9ac4901af2c25745379f35c6185ff1048296c3294fa8af412102e64563b3abf6198463557bf91a65a8460b521e43ce14e1740b42bf693569518bffffffff4bd909585ac05f905b90526c3134a9ea194624650b2823f70f28dcbf9f550dfa060000006a47304402205d50e1375f2fddc97ffaccee772b64fb30e45cbf4179f99d0a4164aca277b28b02200d4879b351e34a3282e65b9c5fca7b15873cdb3ed54c9e9b9e72a73eae9ae775412103ba85a8d6a3cc4c7ae2b3e1a5ea69a0a3afbcda34778965d3487e6ab4f50e3fa0ffffffff03e8030000000000001976a91465ed0bc03b7086e24392dd669440124b9e18809e88acdc050000000000001976a91467a7d9685d1dffa5aa589118c6a6e5be4d39285c88acd2260000000000001976a914c23b02b8b15c530dd61a96eb79a9c2c0776311b988ac00000000")
	tx.BlockHeight = 813392

	// mark it as mined
	merkleProofJson := []byte("{\"index\":2249,\"txOrId\":\"43dd9cb05fbcf4bbc3031b5f3b21e429800ac99aaac351c36a169862c8734270\",\"targetType\":\"header\",\"target\":\"00200120f6a7b7e0b511ae3d1a381f616e906a129f9b943a20feb101000000000000000092cc53712907d7c112ab3f7fb366e87c63f06041bda26d3df96d158774ca1e6b67a126659ca80f1833545630\",\"nodes\":[\"e5b249d701803b15c45db7b9833aec08be8c92e20c15e0b480a866338105f9ee\",\"3a2288227bcb73bfcb6a150528a9cc7f324b32a86591f6b7b33846e858424d61\",\"d18988bfd176523bcfef080d86f5503edc163e8216da002a56396a7d8e7c1a1a\",\"6c4c33ee4753948a85b20e55b67afef6d5d56dd46c99307f61ad225e770e247f\",\"f3c0d5351dd4348be18e9810394455fa3b98fb0b9736861b36b4a1316690670c\",\"8fd6f6267ecb6b90cd60ad7962945969b999ad69a6ca60a65361e4bff23ccb9f\",\"61d2bc392d249030074030092809bc494009caccde7ecf603e3b16d86b748418\",\"5a4b842e847daa5ace9e937496f074f70c39c976e60c200d30353229658a5d69\",\"911a0890cb9f42a3e050cfb6713d9e726423143187c8b4a617e386a81b46b229\",\"9cc0711a4de16193a887c908ef0c286c1b213b49c4a20bd945b092badcd7127f\",\"c4e2aa7d17400375de17085bb2d2203ff3c4a5ae42ab2ef8e8809c0c83e3e3a5\",\"8e17e7ce316c452274a611358b164afedc804b11badf544976f128acdc4ea2cb\"]}")
	var mp bc.MerkleProof
	json.Unmarshal(merkleProofJson, &mp)

	tx.MerkleProof = MerkleProof(mp)

	return tx
}

func getTxReadyToSpend() *Transaction {
	// tx sended from handcash to false_tx_test@bux-wallet1.4chain.dev (1000 satoshis)
	// https://www.whatsonchain.com/tx/6bad032936606b024118b58cbce2e2182103c31533d8c544f0ef3fbf6f5ef436
	tx := newTransaction("0100000002e553c74e4204ef6e01bf1b14d654d29e4bfe8de5f1b1e68430447cb146d67618010000006a473044022067ab22e75b00eadf7b7907cd88aa3ad9b29d4c3f69204afd6e95c3b39beda30c022077bf08a5f0cd33b1bf13c258ad613cc0e591c42ad37b8924bed50c65104f8a86412102e9fcc4d4b2bf1e444f5bbd9bf07c41b8748fb0130f9a5c94a72137a2180d7df3ffffffff34cf06b8008bbcdd1a0e8203cd73c876fc47857a2d6b558078e5216a43ff06c6060000006a4730440220593f00937f92d46d7752b26944b85b68c625cc471d30a4d8f166854bf6c6afc2022013114c1a02b9c442f5ade3826ef361e062fba0b7133f740522f4074fdbe8e4ae41210245357d97c4177e469b965e2dbff92751b414823442fc6ecdc6ee12214cfaee8dffffffff03e8030000000000001976a91453c1a5432bcae1b4894bd9baa3cfb828549c382888acc4090000000000001976a9145539528a3a5abc5c4227ba40bfb471245c838c4c88accd4d0000000000001976a914fa1b182e3c73355d08d4f1eb3023e039f7f117c888ac00000000")
	tx.BlockHeight = 813384

	// mark it as mined
	merkleProofJson := []byte("{\"index\":23837,\"txOrId\":\"6bad032936606b024118b58cbce2e2182103c31533d8c544f0ef3fbf6f5ef436\",\"targetType\":\"header\",\"target\":\"00e0ff2775e0a301d31c556df9023a3211915448a84fffb37ca68e08000000000000000078ce01186d8e33b4da80329a705226d3a13099e527861efa54440b2118eb128629972665be091018cdebf757\",\"nodes\":[\"f4673751fd7eee3332043ac9c7c96e268c53238f23be6205a55764e04d84cc71\",\"04a68bb95545d4850a26fa75ee2ecc913ef37bbf794d3c1bd9f4abe87e37b937\",\"78e99356097514a315ce9c2fa0272c60096f6c36ef7b9978c6c4437fc00683e5\",\"c573f71278eae27462dff1b6318fbabefa29a6b6d427598e4ce88b6fb970a9b4\",\"8d40cba905f8cc90d4f2925b420d33125affec1c27a8a7226d5619e5f770720c\",\"1f89b7ee0fff2ad66400e2784ca817a29ebe2700d1c62d4f525d24994c4205fa\",\"c3c2e00020325dd129d2d0f6509380dec7df09f36be02afc6539796cb043735c\",\"46dd19fa8e7d545226e277cbebe8b204dd2c3789ddec5f2153f0a162c64a8f71\",\"c0fbed07b637f5e6bc1e643846a283a7ff8bfb3f2827cfe3eb8f320983f6ca0f\",\"92886151b4ba502a93ecc480f229b342cae18448aee078fe82aedb946d27e201\",\"f40c401a99bd453364ae24c21aa9fb5d5d871713873475183f400742f1bb5eba\",\"79b778c55f9cf2ee1691e0a01ca3b350ace52f0bf0f8af1eef268f3789b626ed\",\"d208383ea23041649b5d271dcfb0e398f6e5e08ebd482054cb10d8fc518a331a\",\"c56cace5be989e875676b58f0bce74ec7a362341fd5f37232a9ddc9e6010a48f\",\"b375fda2cea55c72f51e97556afbd9001ab0c2c4d06688c0c38b4295230d8496\",\"0576e1a65ddcc3bf71f4ad88b01b696a7b7d11371c7df9ae94cad3ec2c8a9230\"]}")
	var mp bc.MerkleProof
	json.Unmarshal(merkleProofJson, &mp)

	tx.MerkleProof = MerkleProof(mp)

	return tx
}

//
func prepareTxToTest(pubKey, toPaymail string, parentTx *Transaction, satoshis uint64, utxoIdx uint32, btOpts []func(*bt.Tx)) *Transaction {
	falseInput := TransactionInput{
		Utxo: *newUtxoFromTxID(parentTx.ID, utxoIdx),
	}

	// false output
	falseOutput := TransactionOutput{
		Satoshis: satoshis,
		To:       toPaymail,
	}

	falseConfig := TransactionConfig{
		Inputs:  []*TransactionInput{&falseInput},
		Outputs: []*TransactionOutput{&falseOutput},
	}
	falseDraft := newDraftTransaction(pubKey, &falseConfig)

	// prepare markle proof
	merkleProofs := make(map[uint64][]MerkleProof)
	if parentTx.MerkleProof.TxOrID == "" {
		panic("not minded parent tx")
	}

	merkleProofs[parentTx.BlockHeight] = append(merkleProofs[parentTx.BlockHeight], parentTx.MerkleProof)
	for _, v := range merkleProofs {
		cmp, err := CalculateCompoundMerklePath(v)
		if err != nil {
			panic(err)
		}
		falseDraft.CompoundMerklePathes = append(falseDraft.CompoundMerklePathes, cmp)
	}

	// create bt tx for hex
	inputUtxos, _, err := falseDraft.getInputsFromUtxos([]*Utxo{&falseInput.Utxo})
	if err != nil {
		panic(err)
	}

	tx := bt.NewTx()
	if err := tx.FromUTXOs(*inputUtxos...); err != nil {
		panic(err)
	}

	// Add the outputs to the bt transaction
	if err = falseDraft.addOutputsToTx(tx); err != nil {
		panic(err)
	}

	if btOpts != nil {
		for _, opt := range btOpts {
			opt(tx)
		}
	}

	falseDraft.Hex = tx.String()

	falseTx := newTransaction(falseDraft.Hex)
	falseTx.draftTransaction = falseDraft

	return falseTx
}

// utils
func printTx(tx *Transaction) {
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

// PrettyPrint prints JSON in a friendly format
func _prettyPrint(v interface{}) {
	vJson, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}

	fmt.Printf("%s\n", vJson)
}
