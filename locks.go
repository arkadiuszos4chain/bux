package bux

import (
	"context"

	"github.com/mrz1836/go-cachestore"
)

const (
	lockKeyMonitorLockID      = "monitor-lock-id-%s"               // + Lock ID
	lockKeyProcessBroadcastTx = "process-broadcast-transaction-%s" // + Tx ID
	lockKeyProcessIncomingTx  = "process-incoming-transaction-%s"  // + Tx ID
	lockKeyProcessP2PTx       = "process-p2p-transaction-%s"       // + Tx ID
	lockKeyProcessSyncTx      = "process-sync-transaction-task"
	lockKeyProcessXpub        = "action-xpub-id-%s"             // + Xpub ID
	lockKeyRecordBlockHeader  = "action-record-block-header-%s" // + Hash id
	lockKeyRecordTx           = "action-record-transaction-%s"  // + Tx ID
	lockKeyReserveUtxo        = "utxo-reserve-xpub-id-%s"       // + Xpub ID
)

// newWriteLock will take care of creating a lock and defer
func newWriteLock(ctx context.Context, lockKey string, cacheStore cachestore.LockService) (func(), error) {
	secret, err := cacheStore.WriteLock(ctx, lockKey, defaultCacheLockTTL)
	return func() {
		// context is not set, since the req could be canceled, but unlocking should never be stopped
		_, _ = cacheStore.ReleaseLock(context.Background(), lockKey, secret)
	}, err
}

// newWaitWriteLock will take care of creating a lock and defer
func newWaitWriteLock(ctx context.Context, lockKey string, cacheStore cachestore.LockService) (func(), error) {
	secret, err := cacheStore.WaitWriteLock(ctx, lockKey, defaultCacheLockTTL, defaultCacheLockTTW)
	return func() {
		// context is not set, since the req could be canceled, but unlocking should never be stopped
		_, _ = cacheStore.ReleaseLock(context.Background(), lockKey, secret)
	}, err
}
