package bux

import (
	"context"
	"fmt"
	"testing"

	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bk/bip32"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RevertTransaction(t *testing.T) {
	t.Run("revert transaction", func(t *testing.T) {
		ctx, client, transaction, _, deferMe := initRevertTransactionData(t)
		defer deferMe()

		//
		// Revert the transaction
		//
		err := client.RevertTransaction(ctx, transaction.ID)
		require.NoError(t, err)

		// check transaction was reverted
		var tx *Transaction
		tx, err = client.GetTransaction(ctx, testXPubID, transaction.ID)
		require.NoError(t, err)
		assert.Equal(t, transaction.ID, tx.ID)
		assert.Len(t, tx.XpubInIDs, 1) // XpubInIDs should have been set to reverted
		assert.Equal(t, "reverted", tx.XpubInIDs[0])
		assert.Len(t, tx.XpubOutIDs, 1) // XpubInIDs should have been set to reverted
		assert.Equal(t, "reverted", tx.XpubOutIDs[0])
		assert.Len(t, tx.XpubOutputValue, 1) // XpubInIDs should have been set to reverted
		assert.Equal(t, int64(0), tx.XpubOutputValue["reverted"])

		// check the balance of the xpub
		var xpub *Xpub
		xpub, err = client.GetXpubByID(ctx, testXPubID)
		require.NoError(t, err)
		assert.Equal(t, uint64(100000), xpub.CurrentBalance) // 100000 was initial value

		// check sync transaction was canceled
		var syncTx *SyncTransaction
		syncTx, err = GetSyncTransactionByID(ctx, transaction.ID, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Equal(t, SyncStatusCanceled, syncTx.BroadcastStatus)

		// check utxos where reverted
		var utxos []*Utxo
		conditions := &map[string]interface{}{
			xPubIDField: transaction.XPubID,
		}
		utxos, err = client.GetUtxos(ctx, nil, conditions, nil, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Len(t, utxos, 2) // only original
		for _, utxo := range utxos {
			if utxo.TransactionID == transaction.ID {
				assert.True(t, utxo.SpendingTxID.Valid)
				assert.Equal(t, "deleted", utxo.SpendingTxID.String)
			} else {
				assert.False(t, utxo.SpendingTxID.Valid)
				assert.Equal(t, "", utxo.SpendingTxID.String)
			}
		}
	})

	t.Run("disallow revert spent transaction", func(t *testing.T) {
		ctx, client, transaction, xPriv, deferMe := initRevertTransactionData(t)
		defer deferMe()

		// we need a draft transaction, otherwise we cannot revert
		draftTransaction := newDraftTransaction(
			testXPub, &TransactionConfig{
				Outputs: []*TransactionOutput{{
					To:       "1A1PjKqjWMNBzTVdcBru27EV1PHcXWc63W", // random address
					Satoshis: 1000,
				}},
				ChangeNumberOfDestinations: 1,
				Sync: &SyncConfig{
					Broadcast:        true,
					BroadcastInstant: false,
					PaymailP2P:       false,
					SyncOnChain:      false,
				},
			},
			append(client.DefaultModelOptions(), New())...,
		)
		// this gets inputs etc.
		err := draftTransaction.Save(ctx)
		require.NoError(t, err)

		var hex string
		hex, err = draftTransaction.SignInputs(xPriv)
		require.NoError(t, err)
		assert.NotEmpty(t, hex)

		var secondTransaction *Transaction
		secondTransaction, err = client.RecordTransaction(ctx, testXPub, hex, draftTransaction.ID, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.NotEmpty(t, secondTransaction)

		//
		// Revert the transaction
		//
		err = client.RevertTransaction(ctx, transaction.ID)
		require.Equal(t, "utxo of this transaction has been spent, cannot revert", err.Error())
	})

	t.Run("revert spend to internal address", func(t *testing.T) {
		ctx, client, _, xPriv, deferMe := initRevertTransactionData(t)
		defer deferMe()

		testXPub2 := "xpub661MyMwAqRbcFGX8a3K99DKPZahQBj1z8DsMTE7gqKtYj9yaWv45nkjHYcWdwUcQkGdZMv62HVKNCF4MNqXK2oiRKcfSE7U7iu5hAcyMzUS"
		xPub := newXpub(testXPub2, append(client.DefaultModelOptions(), New())...)
		err := xPub.Save(ctx)
		require.NoError(t, err)

		var destination *Destination
		destination, err = xPub.getNewDestination(ctx, utils.ChainExternal, utils.ScriptTypePubKeyHash, client.DefaultModelOptions(New())...)
		require.NoError(t, err)
		require.NotNil(t, destination)

		err = destination.Save(ctx)
		require.NoError(t, err)

		// we need a draft transaction, otherwise we cannot revert
		draftTransaction := newDraftTransaction(
			testXPub, &TransactionConfig{
				Outputs: []*TransactionOutput{{
					To:       destination.Address,
					Satoshis: 1000,
				}},
				ChangeNumberOfDestinations: 1,
				Sync: &SyncConfig{
					Broadcast:        true,
					BroadcastInstant: false,
					PaymailP2P:       false,
					SyncOnChain:      false,
				},
			},
			append(client.DefaultModelOptions(), New())...,
		)
		// this gets inputs etc.
		err = draftTransaction.Save(ctx)
		require.NoError(t, err)

		var hex string
		hex, err = draftTransaction.SignInputs(xPriv)
		require.NoError(t, err)
		assert.NotEmpty(t, hex)

		var transaction *Transaction
		transaction, err = client.RecordTransaction(ctx, testXPub, hex, draftTransaction.ID, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.NotEmpty(t, transaction)
		assert.Len(t, transaction.XpubOutIDs, 2)
		assert.Equal(t, int64(1000), transaction.XpubOutputValue[xPub.ID])

		xPub, err = client.GetXpub(ctx, testXPub2)
		require.NoError(t, err)
		assert.Equal(t, uint64(1000), xPub.CurrentBalance)

		var utxos []*Utxo
		utxos, err = client.GetUtxosByXpubID(ctx, xPub.ID, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, utxos, 1)
		assert.Equal(t, uint64(1000), utxos[0].Satoshis)
		assert.False(t, utxos[0].SpendingTxID.Valid)

		//
		// Revert the transaction
		//
		err = client.RevertTransaction(ctx, transaction.ID)
		require.NoError(t, err)

		// check the destination xpub / utxos etc
		xPub, err = client.GetXpub(ctx, testXPub2)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), xPub.CurrentBalance)

		utxos, err = client.GetUtxosByXpubID(ctx, xPub.ID, nil, nil, nil)
		require.NoError(t, err)
		assert.Len(t, utxos, 1)
		assert.True(t, utxos[0].SpendingTxID.Valid)
		assert.Equal(t, "deleted", utxos[0].SpendingTxID.String)
	})
}

func initRevertTransactionData(t *testing.T) (context.Context, ClientInterface, *Transaction, *bip32.ExtendedKey, func()) {
	// this creates an xpub, destination and utxo
	ctx, client, deferMe := initSimpleTestCase(t)

	// we need a draft transaction, otherwise we cannot revert
	draftTransaction := newDraftTransaction(
		testXPub, &TransactionConfig{
			Outputs: []*TransactionOutput{{
				To:       "1A1PjKqjWMNBzTVdcBru27EV1PHcXWc63W", // random address
				Satoshis: 1000,
			}},
			ChangeNumberOfDestinations: 1,
			Sync: &SyncConfig{
				Broadcast:        true,
				BroadcastInstant: false,
				PaymailP2P:       false,
				SyncOnChain:      false,
			},
		},
		append(client.DefaultModelOptions(), New())...,
	)
	// this gets inputs etc.
	err := draftTransaction.Save(ctx)
	require.NoError(t, err)

	var xPriv *bip32.ExtendedKey
	xPriv, err = bip32.NewKeyFromString(testXPriv)
	require.NoError(t, err)

	var hex string
	hex, err = draftTransaction.SignInputs(xPriv)
	require.NoError(t, err)
	assert.NotEmpty(t, hex)

	newOpts := client.DefaultModelOptions(WithXPub(testXPub), New())
	transaction := newTransactionWithDraftID(
		hex, draftTransaction.ID, newOpts...,
	)
	transaction.draftTransaction = draftTransaction
	_hydrateOutgoingWithSync(transaction)
	err = transaction.processUtxos(ctx)
	require.NoError(t, err)

	err = transaction.Save(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, transaction)

	// check transaction was recorded properly
	var tx *Transaction
	tx, err = client.GetTransaction(ctx, testXPubID, transaction.ID)
	require.NoError(t, err)
	assert.Equal(t, transaction.ID, tx.ID)
	assert.Equal(t, testXPubID, tx.XpubInIDs[0])

	// check sync transaction
	var syncTx *SyncTransaction
	syncTx, err = GetSyncTransactionByID(ctx, transaction.ID, client.DefaultModelOptions()...)
	require.NoError(t, err)
	assert.Equal(t, SyncStatusReady, syncTx.BroadcastStatus)

	var utxos []*Utxo
	conditions := &map[string]interface{}{
		xPubIDField: transaction.XPubID,
	}
	utxos, err = client.GetUtxos(ctx, nil, conditions, nil, client.DefaultModelOptions()...)
	require.NoError(t, err)
	assert.Len(t, utxos, 2) // original + new change utxo
	for _, utxo := range utxos {
		if utxo.TransactionID == transaction.ID {
			assert.False(t, utxo.SpendingTxID.Valid)
		} else {
			assert.Equal(t, transaction.ID, utxo.SpendingTxID.String)
		}
	}

	return ctx, client, transaction, xPriv, deferMe
}

// BenchmarkAction_Transaction_recordTransaction will benchmark the method RecordTransaction()
func BenchmarkAction_Transaction_recordTransaction(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx, client, xPub, config, err := initBenchmarkData(b)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}

		var draftTransaction *DraftTransaction
		if draftTransaction, err = client.NewTransaction(ctx, xPub.rawXpubKey, config, client.DefaultModelOptions()...); err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}

		var xPriv *bip32.ExtendedKey
		if xPriv, err = bip32.NewKeyFromString(testXPriv); err != nil {
			return
		}

		var hexString string
		if hexString, err = draftTransaction.SignInputs(xPriv); err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}

		b.StartTimer()
		if _, err = client.RecordTransaction(ctx, xPub.rawXpubKey, hexString, draftTransaction.ID, client.DefaultModelOptions()...); err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}
	}
}

// BenchmarkTransaction_newTransaction will benchmark the method newTransaction()
func BenchmarkAction_Transaction_newTransaction(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx, client, xPub, config, err := initBenchmarkData(b)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}

		b.StartTimer()
		if _, err = client.NewTransaction(ctx, xPub.rawXpubKey, config, client.DefaultModelOptions()...); err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}
	}
}

func initBenchmarkData(b *testing.B) (context.Context, ClientInterface, *Xpub, *TransactionConfig, error) {
	ctx, client, _ := CreateBenchmarkSQLiteClient(b, false, true,
		WithCustomTaskManager(&taskManagerMockBase{}),
		WithFreeCache(),
		WithIUCDisabled(),
	)

	opts := append(client.DefaultModelOptions(), New())
	xPub, err := client.NewXpub(ctx, testXPub, opts...)
	if err != nil {
		b.Fail()
	}
	destination := newDestination(xPub.GetID(), testLockingScript, opts...)
	if err = destination.Save(ctx); err != nil {
		b.Fail()
	}

	utxo := newUtxo(xPub.GetID(), testTxID, testLockingScript, 1, 122500, opts...)
	if err = utxo.Save(ctx); err != nil {
		b.Fail()
	}
	utxo = newUtxo(xPub.GetID(), testTxID, testLockingScript, 2, 122500, opts...)
	if err = utxo.Save(ctx); err != nil {
		b.Fail()
	}
	utxo = newUtxo(xPub.GetID(), testTxID, testLockingScript, 3, 122500, opts...)
	if err = utxo.Save(ctx); err != nil {
		b.Fail()
	}
	utxo = newUtxo(xPub.GetID(), testTxID, testLockingScript, 4, 122500, opts...)
	if err = utxo.Save(ctx); err != nil {
		b.Fail()
	}

	config := &TransactionConfig{
		FeeUnit: &utils.FeeUnit{
			Satoshis: 5,
			Bytes:    100,
		},
		Outputs: []*TransactionOutput{{
			OpReturn: &OpReturn{
				Map: &MapProtocol{
					App:  "getbux.io",
					Type: "blast",
					Keys: map[string]interface{}{
						"bux": "blasting",
					},
				},
			},
		}},
		ChangeDestinationsStrategy: ChangeStrategyRandom,
		ChangeNumberOfDestinations: 2,
	}

	return ctx, client, xPub, config, err
}
