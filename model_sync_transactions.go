package bux

import (
	"context"
	"fmt"

	"github.com/BuxOrg/bux/taskmanager"
	"github.com/mrz1836/go-datastore"
	customTypes "github.com/mrz1836/go-datastore/custom_types"
)

// SyncTransaction is an object representing the chain-state sync configuration and results for a given transaction
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type SyncTransaction struct {
	// Base model
	Model `bson:",inline"`

	// Model specific fields
	ID              string               `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the unique transaction id" bson:"_id"`
	Configuration   SyncConfig           `json:"configuration" toml:"configuration" yaml:"configuration" gorm:"<-;type:text;comment:This is the configuration struct in JSON" bson:"configuration"`
	LastAttempt     customTypes.NullTime `json:"last_attempt" toml:"last_attempt" yaml:"last_attempt" gorm:"<-;comment:When the last broadcast occurred" bson:"last_attempt,omitempty"`
	Results         SyncResults          `json:"results" toml:"results" yaml:"results" gorm:"<-;type:text;comment:This is the results struct in JSON" bson:"results"`
	BroadcastStatus SyncStatus           `json:"broadcast_status" toml:"broadcast_status" yaml:"broadcast_status" gorm:"<-;type:varchar(10);index;comment:This is the status of the broadcast" bson:"broadcast_status"`
	P2PStatus       SyncStatus           `json:"p2p_status" toml:"p2p_status" yaml:"p2p_status" gorm:"<-;column:p2p_status;type:varchar(10);index;comment:This is the status of the p2p paymail requests" bson:"p2p_status"`
	SyncStatus      SyncStatus           `json:"sync_status" toml:"sync_status" yaml:"sync_status" gorm:"<-;type:varchar(10);index;comment:This is the status of the on-chain sync" bson:"sync_status"`

	// internal fields
	transaction *Transaction
}

// newSyncTransaction will start a new model (config is required)
func newSyncTransaction(txID string, config *SyncConfig, opts ...ModelOps) *SyncTransaction {
	// Do not allow making a model without the configuration
	if config == nil {
		return nil
	}

	// Broadcasting
	bs := SyncStatusReady
	if !config.Broadcast {
		bs = SyncStatusSkipped
	}

	// Notify Paymail P2P
	ps := SyncStatusPending
	if !config.PaymailP2P {
		ps = SyncStatusSkipped
	}

	// Sync
	ss := SyncStatusReady
	if !config.SyncOnChain {
		ss = SyncStatusSkipped
	}

	return &SyncTransaction{
		BroadcastStatus: bs,
		Configuration:   *config,
		ID:              txID,
		Model:           *NewBaseModel(ModelSyncTransaction, opts...),
		P2PStatus:       ps,
		SyncStatus:      ss,
	}
}

// GetID will get the ID
func (m *SyncTransaction) GetID() string {
	return m.ID
}

// isSkipped will return true if Broadcasting, P2P and SyncOnChain are all skipped
func (m *SyncTransaction) isSkipped() bool {
	return m.BroadcastStatus == SyncStatusSkipped &&
		m.SyncStatus == SyncStatusSkipped &&
		m.P2PStatus == SyncStatusSkipped
}

// GetModelName will get the name of the current model
func (m *SyncTransaction) GetModelName() string {
	return ModelSyncTransaction.String()
}

// GetModelTableName will get the db table name of the current model
func (m *SyncTransaction) GetModelTableName() string {
	return tableSyncTransactions
}

// Save will save the model into the Datastore
func (m *SyncTransaction) Save(ctx context.Context) error {
	return Save(ctx, m)
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *SyncTransaction) BeforeCreating(_ context.Context) error {
	m.DebugLog("starting: [" + m.name.String() + "] BeforeCreating hook...")

	// Make sure ID is valid
	if len(m.ID) == 0 {
		return ErrMissingFieldID
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return nil
}

// AfterCreated will fire after the model is created in the Datastore
func (m *SyncTransaction) AfterCreated(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterCreated hook...")

	m.DebugLog("end: " + m.Name() + " AfterCreated hook")
	return nil
}

func (m *SyncTransaction) BeforeUpdating(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " BeforeUpdate hook...")

	// Trim the results to the last 20
	maxResultsLength := 20

	ln := len(m.Results.Results)
	if ln > maxResultsLength {
		m.Client().Logger().
			Warn(ctx, fmt.Sprintf("trimming syncTx.Results, TxID: %s", m.ID))

		m.Results.Results = m.Results.Results[ln-maxResultsLength:]
	}

	m.DebugLog("end: " + m.Name() + " BeforeUpdate hook")
	return nil
}

// Migrate model specific migration on startup
func (m *SyncTransaction) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableSyncTransactions), metadataField)
}

// RegisterTasks will register the model specific tasks on client initialization
func (m *SyncTransaction) RegisterTasks() error {
	// No task manager loaded?
	tm := m.Client().Taskmanager()
	if tm == nil {
		return nil
	}

	// Sync with chain - task
	// Register the task locally (cron task - set the defaults)
	syncTask := m.Name() + "_" + syncActionSync
	ctx := context.Background()

	// Register the task
	if err := tm.RegisterTask(&taskmanager.Task{
		Name:       syncTask,
		RetryLimit: 1,
		Handler: func(client ClientInterface) error {
			if taskErr := taskSyncTransactions(ctx, client, WithClient(client)); taskErr != nil {
				client.Logger().Error(ctx, "error running "+syncTask+" task: "+taskErr.Error())
			}
			return nil
		},
	}); err != nil {
		return err
	}

	// Run the task periodically
	err := tm.RunTask(ctx, &taskmanager.TaskOptions{
		Arguments:      []interface{}{m.Client()},
		RunEveryPeriod: m.Client().GetTaskPeriod(syncTask),
		TaskName:       syncTask,
	})
	if err != nil {
		return err
	}

	// Broadcast - task
	// Register the task locally (cron task - set the defaults)
	broadcastTask := m.Name() + "_" + syncActionBroadcast

	// Register the task
	if err = tm.RegisterTask(&taskmanager.Task{
		Name:       broadcastTask,
		RetryLimit: 1,
		Handler: func(client ClientInterface) error {
			if taskErr := taskBroadcastTransactions(ctx, client.Logger(), WithClient(client)); taskErr != nil {
				client.Logger().Error(ctx, "error running "+broadcastTask+" task: "+taskErr.Error())
			}
			return nil
		},
	}); err != nil {
		return err
	}

	// Run the task periodically
	if err = tm.RunTask(ctx, &taskmanager.TaskOptions{
		Arguments:      []interface{}{m.Client()},
		RunEveryPeriod: m.Client().GetTaskPeriod(broadcastTask),
		TaskName:       broadcastTask,
	}); err != nil {
		return err
	}

	return nil
}
