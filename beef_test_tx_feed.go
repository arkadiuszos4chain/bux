package bux

import (
	"encoding/json"

	"github.com/libsv/go-bc"
)

// real tx feed
func getAlreadySpendedTx() *Transaction {
	// tx sended from handcash to false_tx_test@bux-wallet1.4chain.dev, all satoshis (1500) were sended back
	// https://www.whatsonchain.com/tx/2427e53681a3f21d295497888bceb4c97d1a0cec2ac9e4f016d61dd2128b0f15
	itx := newIncomingTransaction(
		"2427e53681a3f21d295497888bceb4c97d1a0cec2ac9e4f016d61dd2128b0f15",
		"010000000224bb8a3d4f0b2035288bfedb062f12781d3e6eed1755763032e20e323d192f2d000000006a473044022029b16927ef8b7c6553f23d5368886681ace96daaf03ca666285e10dce7a5fa3302204a551a4c25789e23490d0534a70838caa1955b397f41fec82375f897095655e04121036add0f78cbf6f7d83c07b5d8db1d416fbf491544b7b16bc3bc0607af35ffdb8effffffff1c25c0777758266bd12bba8be518722c0a3c41a5ae691528d8374c4a55643ddc010000006a47304402201baa3d559ecda0b76f605780e2dc164b6fb627b6e6c07e9a35cb7dd8c9e83fc30220128d8d9dece522c8e34b042a26def20bb13a5774130ca64eea9bd3f04eaa993b4121020c34c26466cc70fd54348d98a63fddbd3cf6075c98f16a24a7fba5cc83697052ffffffff03dc050000000000001976a914fc1c2fdddd7e90d0a88f3a5d22c20260c281fc3b88acac0d0000000000001976a914817ebd5a7b67bc411ad232af8e74574e29a6ab3788acf9000000000000001976a914a02f449196c8f3f14461b240efa86c470327928088ac00000000",
	)

	tx := newTransactionFromIncomingTransaction(itx)
	tx.BlockHeight = 814395

	// mark it as mined
	merkleProofJson := []byte("{\"index\":11173,\"txOrId\":\"2427e53681a3f21d295497888bceb4c97d1a0cec2ac9e4f016d61dd2128b0f15\",\"targetType\":\"header\",\"target\":\"008001206b36aa94cff773fad885c9b969a618691bc2a6c34fe576010000000000000000d75bd60b27a1a4083b674970c264efbca62b35ff77df7f9ba18597a6c21e1b0b33bb2f6571d90b18cfe9dbce\",\"nodes\":[\"786a84ba80d035dbf401768008864af6064e321e97608f9234ff16d0e46c6a10\",\"9f41dbb064e154b777184807e85f8b0c5feaf86a885713f5f803d9f5cc232359\",\"602d93502ec3dde3101c10f0a6381d5fb92f0af9edbf7f659dbbd4291c8c8eb9\",\"0eaba2c4e27a6baee2e7323a255a457b700ed64ec49c7c539cb7996d8194980f\",\"c8971afa566b3aa958af6e37d9b8d5f203b647771246c6b014c4b31f0adf7bbb\",\"ce33a2712e665af1105af6e4ccc4de58d82d7502586f6806898d4eb839d1f53c\",\"38e941a638af1f5118c1808811675ec36fa1b47311e9f95314bd4b3d0bd7b0bf\",\"1d7182cbc6fe65b660aa0fd7e367318bfee610c0d9b7e41da4c4fe67656460c5\",\"fd75e8cb577a5e6e8c4cc0b38820f5988284cfb17be8faef60ef9a927c34475f\",\"a178170a1afc33fb5e7572e60a578fefc8e2ac7437ef11455d42ebfaf87fb7c8\",\"caa0fca6fefceddcb4f27c5ec12dfa97726f523d94d58f59823ef98ad18d7f87\",\"84d3107c4c18dbf5d21f6a38ebf6776ca19b8a9c9916a0e52ec8f1899a75c4f0\",\"f14062918a762cc1829a50a2c3cdebda87eb43738a8dc9a78c8ae1c039094421\",\"863237cc986324525fe84ab09709fc701037ff47d03134dcdb1abe0835480d6e\",\"575496a5650fcf48638fc3b6ba56020886a0826608711f51ce7d47d5778fae4f\",\"81f6e4c7978604f5155c712c379e48b94ac892271d8bee794391fc4599c9d1c9\"]}")
	var mp bc.MerkleProof
	json.Unmarshal(merkleProofJson, &mp)

	tx.MerkleProof = MerkleProof(mp)

	return tx
}

func getSomeoneElseTx() *Transaction {
	// tx sended from handcash to false_tx_test_helper@bux-wallet1.4chain.dev (1000 satoshis)
	// https://www.whatsonchain.com/tx/82187e52b7e057122d4871852d1062238522588eb658617cf48c7ac02e388a08

	itx := newIncomingTransaction(
		"82187e52b7e057122d4871852d1062238522588eb658617cf48c7ac02e388a08",
		"0100000002704273c86298166ac351c3aa9ac90a8029e4213b5f1b03c3bbf4bc5fb09cdd43010000006a4730440220398d6389e8a156a3c6c1ca355e446d844fd480193a93af832afd1c87d0f04784022050091076b8f7405b37ce6e795d1b92526396ac2b14f08e91649b908e711e2b044121030ef6975d46dbab4b632ef62fdbe97de56d183be1acc0be641d2c400ae01cf136ffffffff2f41ed6a2488ac3ba4a3c330a15fa8193af87f0192aa59935e6c6401d92dc3a00a0000006a47304402200ad9cf0dc9c90a4c58b08910740b4a8b3e1a7e37db1bc5f656361b93f412883d0220380b6b3d587103fc8bf3fe7bed19ab375766984c67ebb7d43c993bcd199f32a441210205ef4171f58213b5a2ddf16ac6038c10a2a8c3edc1e6275cb943af4bb3a58182ffffffff03e8030000000000001976a9148a8c4546a95e6fc8d18076a9980d59fd882b4e6988acf4010000000000001976a914c7662da5e0a6a179141a7872045538126f1e954288acf5000000000000001976a914765bdf10934f5aac894cf8a3795c9eeb494c013488ac00000000",
	)

	tx := newTransactionFromIncomingTransaction(itx)
	tx.BlockHeight = 814414

	// mark it as mined
	merkleProofJson := []byte("{\"index\":37008,\"txOrId\":\"82187e52b7e057122d4871852d1062238522588eb658617cf48c7ac02e388a08\",\"targetType\":\"header\",\"target\":\"00e0ff2f886610c430f0b901065a1ae1258e8bfcadce03b2e8db2c040000000000000000d031f2c8b6ba0dfc710202a5d242a79636b7d4b1bcefec7e828cfc4a03516fd04aee2f6554be0b1854e634c6\",\"nodes\":[\"3228f78cfd3c96262ec521225f1b9dd6326b4d3e245d1551bb06258f2101cb65\",\"05267706279d2e5ebcf89ed0645d4283108c7e850cdb84aeb0974738ae447a8d\",\"5ad82f15dd63ed2ca1e319e91404c918b4a5e71312acdb0ad7c83136a38af360\",\"15c69ba0b2a2d8f503637db4053a2d551e86ef669fa25091c7a0a764ee32f15f\",\"b3841a856957ef968213b160d5a22cdc4db1bde9dc05fe9d43b906cf21af0bcc\",\"51ce3503a1bbe73a5b1594884e752b567aa6190f7ab4cdbed3e1a63262b2cc63\",\"bc50e5cc34225b3e3bbf9e0b7444cf4d90af70af115c8872832c0f5a45fc53e1\",\"9187bd785e2737c028e9faa81801193818deebdc6a5de6dd79c3dcd169ead577\",\"d143aab223e0c9c4025f58921f93043560efae88277f856951a28416f3e46302\",\"5cd848118c319ba93286e71e089b5585c8464dbf57b6a2cfea5e3e0b2e2972de\",\"ed41d9af8c4cfada67651a3db852abd27f97ead6200972f3e7b999ca31f8a568\",\"0f45052de20d4bdd15b81838fc26a67d5c36f66a05a8d602b756903fd8910b6a\",\"df4d66a947d89e258b818493f3460d88e7dd0b203de95958a74585b656ce7698\",\"c3719736ebc72175f49c82520d16fe83f983a139ec86577b92bcac5ce9c50b99\",\"d73795a577f57306eec549b9fdabeadf47bf50d0510a01b690f50503e594e74f\",\"d059074b8118309d1a8e74a4074ee45a1f903925911719bcd8a4ae58a344ad4d\"]}")
	var mp bc.MerkleProof
	json.Unmarshal(merkleProofJson, &mp)

	tx.MerkleProof = MerkleProof(mp)

	return tx
}

func getTxReadyToSpend() *Transaction {
	// tx sended from handcash to false_tx_test@bux-wallet1.4chain.dev (1000 satoshis)
	// https://www.whatsonchain.com/tx/0c165cfc63963848fcbb00f6da64e078b99cfb88c46293313728c8fd0265e38a
	itx := newIncomingTransaction(
		"0c165cfc63963848fcbb00f6da64e078b99cfb88c46293313728c8fd0265e38a",
		"01000000027b0a1b12c7c9e48015e78d3a08a4d62e439387df7e0d7a810ebd4af37661daaa000000006a47304402207d972759afba7c0ffa6cfbbf39a31c2aeede1dae28d8841db56c6dd1197d56a20220076a390948c235ba8e72b8e43a7b4d4119f1a81a77032aa6e7b7a51be5e13845412103f78ec31cf94ca8d75fb1333ad9fc884e2d489422034a1efc9d66a3b72eddca0fffffffff7f36874f858fb43ffcf4f9e3047825619bad0e92d4b9ad4ba5111d1101cbddfe010000006a473044022043f048043d56eb6f75024808b78f18808b7ab45609e4c4c319e3a27f8246fc3002204b67766b62f58bf6f30ea608eaba76b8524ed49f67a90f80ac08a9b96a6922cd41210254a583c1c51a06e10fab79ddf922915da5f5c1791ef87739f40cb68638397248ffffffff03e8030000000000001976a914b08f70bc5010fb026de018f19e7792385a146b4a88acf3010000000000001976a9147d48635f889372c3da12d75ce246c59f4ab907ed88acf7000000000000001976a914b8fbd58685b6920d8f9a8f1b274d8696708b51b088ac00000000",
	)

	tx := newTransactionFromIncomingTransaction(itx)
	tx.BlockHeight = 814414

	// mark it as mined
	merkleProofJson := []byte("{\"index\":26524,\"txOrId\":\"0c165cfc63963848fcbb00f6da64e078b99cfb88c46293313728c8fd0265e38a\",\"targetType\":\"header\",\"target\":\"00e0ff2f886610c430f0b901065a1ae1258e8bfcadce03b2e8db2c040000000000000000d031f2c8b6ba0dfc710202a5d242a79636b7d4b1bcefec7e828cfc4a03516fd04aee2f6554be0b1854e634c6\",\"nodes\":[\"bdf86816f48280ea612349b0435b585a571a11cd0d97a5838ca53088723213db\",\"3c7a05468cbec89e38656d1bcf3f1f1a772878dbb5f76def16e5aa54697068e5\",\"7b5528b537beadc00987d6eba885517dfcdcfc3f5e76c8558f98cc138d02a601\",\"754d063eeaea31b8e79bc445e1a8c2e346f225b951e7ab1c97e117a2094fc61c\",\"a176885e1b00b39562db4f70a8b73b50d0e28a3fef774805db0ca2262612cf9c\",\"ad47ee74a8c7cbe59511f5293345e3fcc083cbed77bc550bf65f17336796eaae\",\"e9d7a9cc739b34267f8f504cc62a5b9a4b070319ee024ce8d6ff733db5507ff6\",\"68aafc3a48df5eb62faf58871f3e395adcc0f078b5a8dd501cc6d0ee7b4ce76c\",\"c1c8c5e42b80e4b09107f76a2ca96cdbd19d9805004d52621d14dd7be17524e3\",\"7a27692852731ae2722fbb3d3be96450299a82f2dca7a4a1c36cf294f46281b8\",\"77bb554c763fc856063fb1e4e268102ad6f73b4050e4803a6a5bd20df8d338a9\",\"2c34bd37df48b4958c334b75e662e4f682d500dcbfb508c5371ca4fee4c6ea6f\",\"7856e409b5c1a80af8c5bc0b6dc1bb657004ee0cd053fe9a2b3be26b354854bf\",\"444ddb0b389e53fd9237f3473835db1312aecad3ec7a731a31ee37a43174e3c2\",\"eb7a0020601b5f4ba40650df9658edcdbff814e02caf482470e2708729ef5a5d\",\"32a54c346b51c6a273b53a5939c7fb05cbaf6db5b936f325dbfc0330514f4891\"]}")
	var mp bc.MerkleProof
	json.Unmarshal(merkleProofJson, &mp)

	tx.MerkleProof = MerkleProof(mp)

	return tx
}

// TODO
func getTestDestination() *Destination {
	//toPaymail := "false_tx_test_helper@bux-wallet1.4chain.dev"

	dst := Destination{
		Address:       "false_tx_test_helper@bux-wallet1.4chain.dev",
		Chain:         0,
		LockingScript: "76a9148a8c4546a95e6fc8d18076a9980d59fd882b4e6988ac",
		Num:           3,
	}

	return &dst
}
