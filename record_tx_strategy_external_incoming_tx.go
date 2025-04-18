package bux

import (
	"context"
	"fmt"

	"github.com/libsv/go-bt/v2"
	zLogger "github.com/mrz1836/go-logger"
)

type externalIncomingTx struct {
	Hex                  string
	broadcastNow         bool // e.g. BEEF must be broadcasted now
	allowBroadcastErrors bool // only BEEF cannot allow for broadcast errors
}

func (strategy *externalIncomingTx) Execute(ctx context.Context, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	logger := c.Logger()

	// process
	if !strategy.broadcastNow && c.IsITCEnabled() { // do not save transaction to database now, save IncomingTransaction instead and let task manager handle and process it
		return _addTxToCheck(ctx, strategy, c, opts)
	}

	transaction, err := _createExternalTxToRecord(ctx, strategy, c, opts)
	if err != nil {
		return nil, fmt.Errorf("ExternalIncomingTx.Execute(): creation of external incoming tx failed. Reason: %w", err)
	}

	logger.Info(ctx, fmt.Sprintf("ExternalIncomingTx.Execute(): start without ITC, TxID: %s", transaction.ID))

	if strategy.broadcastNow || transaction.syncTransaction.BroadcastStatus == SyncStatusReady {

		err := _externalIncomingBroadcast(ctx, logger, transaction, strategy.allowBroadcastErrors)
		if err != nil {
			logger.
				Error(ctx, fmt.Sprintf("ExternalIncomingTx.Execute(): broadcasting failed, transaction rejected! Reason: %s, TxID: %s", err, transaction.ID))

			return nil, fmt.Errorf("ExternalIncomingTx.Execute(): broadcasting failed, transaction rejected! Reason: %w", err)
		}
	}

	// record
	if err = transaction.Save(ctx); err != nil {
		return nil, fmt.Errorf("ExternalIncomingTx.Execute(): saving of Transaction failed. Reason: %w", err)
	}

	logger.Info(ctx, fmt.Sprintf("ExternalIncomingTx.Execute(): complete, TxID: %s", transaction.ID))
	return transaction, nil
}

func (strategy *externalIncomingTx) Validate() error {
	if strategy.Hex == "" {
		return ErrMissingFieldHex
	}

	return nil // is valid
}

func (strategy *externalIncomingTx) TxID() string {
	btTx, _ := bt.NewTxFromString(strategy.Hex)
	return btTx.TxID()
}

func (strategy *externalIncomingTx) LockKey() string {
	return fmt.Sprintf("incoming-%s", strategy.TxID())
}

func (strategy *externalIncomingTx) ForceBroadcast(force bool) {
	strategy.broadcastNow = force
}

func (strategy *externalIncomingTx) FailOnBroadcastError(forceFail bool) {
	strategy.allowBroadcastErrors = !forceFail
}

func _addTxToCheck(ctx context.Context, tx *externalIncomingTx, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	logger := c.Logger()

	incomingTx := newIncomingTransaction(tx.Hex, c.DefaultModelOptions(append(opts, New())...)...)

	logger.Info(ctx, fmt.Sprintf("ExternalIncomingTx.Execute(): start ITC, TxID: %s", incomingTx.ID))

	if err := incomingTx.Save(ctx); err != nil {
		return nil, fmt.Errorf("ExternalIncomingTx.Execute(): addind new IncomingTx to check queue failed. Reason: %w", err)
	}

	result := incomingTx.toTransactionDto()
	result.Status = statusProcessing

	logger.Info(ctx, fmt.Sprintf("ExternalIncomingTx.Execute(): complete ITC, TxID: %s", incomingTx.ID))
	return result, nil
}

func _createExternalTxToRecord(ctx context.Context, eTx *externalIncomingTx, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	// Create NEW tx model
	tx := newTransaction(eTx.Hex, c.DefaultModelOptions(append(opts, New())...)...)
	_hydrateExternalWithSync(tx)

	if !tx.TransactionBase.hasOneKnownDestination(ctx, c, tx.GetOptions(false)...) {
		return nil, ErrNoMatchingOutputs
	}

	if err := tx.processUtxos(ctx); err != nil {
		return nil, err
	}

	tx.TotalValue, tx.Fee = tx.getValues()
	if tx.TransactionBase.parsedTx != nil {
		tx.NumberOfInputs = uint32(len(tx.TransactionBase.parsedTx.Inputs))
		tx.NumberOfOutputs = uint32(len(tx.TransactionBase.parsedTx.Outputs))
	}

	return tx, nil
}

func _hydrateExternalWithSync(tx *Transaction) {
	sync := newSyncTransaction(
		tx.ID,
		tx.Client().DefaultSyncConfig(),
		tx.GetOptions(true)...,
	)

	// to simplfy: broadcast every external incoming txs
	sync.BroadcastStatus = SyncStatusReady

	sync.P2PStatus = SyncStatusSkipped  // the sender of the Tx should have already notified us
	sync.SyncStatus = SyncStatusPending // wait until transaciton will be broadcasted

	// Use the same metadata
	sync.Metadata = tx.Metadata
	sync.transaction = tx
	tx.syncTransaction = sync
}

func _externalIncomingBroadcast(ctx context.Context, logger zLogger.GormLoggerInterface, tx *Transaction, allowErrors bool) error {
	logger.Info(ctx, fmt.Sprintf("ExternalIncomingTx.Execute(): start broadcast, TxID: %s", tx.ID))

	err := broadcastSyncTransaction(ctx, tx.syncTransaction)

	if err == nil {
		logger.
			Info(ctx, fmt.Sprintf("ExternalIncomingTx.Execute(): broadcast complete, TxID: %s", tx.ID))

		return nil
	}

	if allowErrors {
		logger.
			Warn(ctx, fmt.Sprintf("ExternalIncomingTx.Execute(): broadcasting failed, next try will be handled by task manager. Reason: %s, TxID: %s", err, tx.ID))

		// ignore error, transaction will be broadcaset in a cron task
		return nil
	}

	return err
}
