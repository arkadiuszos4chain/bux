package bux

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/mrz1836/go-datastore"
	"github.com/pkg/errors"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/BuxOrg/bux/utils"
)

// DraftTransaction is an object representing the draft BitCoin transaction prior to the final transaction
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type DraftTransaction struct {
	// Base model
	Model `bson:",inline"`

	// Standard transaction model base fields
	TransactionBase `bson:",inline"`

	// Model specific fields
	XpubID        string            `json:"xpub_id" toml:"xpub_id" yaml:"xpub_id" gorm:"<-:create;type:char(64);index;comment:This is the related xPub" bson:"xpub_id"`
	ExpiresAt     time.Time         `json:"expires_at" toml:"expires_at" yaml:"expires_at" gorm:"<-:create;comment:Time when the draft expires" bson:"expires_at"`
	Configuration TransactionConfig `json:"configuration" toml:"configuration" yaml:"configuration" gorm:"<-;type:text;comment:This is the configuration struct in JSON" bson:"configuration"`
	Status        DraftStatus       `json:"status" toml:"status" yaml:"status" gorm:"<-;type:varchar(10);index;comment:This is the status of the draft" bson:"status"`
	FinalTxID     string            `json:"final_tx_id,omitempty" toml:"final_tx_id" yaml:"final_tx_id" gorm:"<-;type:char(64);index;comment:This is the final tx ID" bson:"final_tx_id,omitempty"`
	BUMPs         BUMPs             `json:"bumps,omitempty" toml:"bumps" yaml:"bumps" gorm:"<-;type:text;comment:Slice of BUMPs (BSV Unified Merkle Paths)" bson:"bumps,omitempty"`
}

// newDraftTransaction will start a new draft tx
func newDraftTransaction(rawXpubKey string, config *TransactionConfig, opts ...ModelOps) *DraftTransaction {
	// Random GUID
	id, _ := utils.RandomHex(32)

	// Set the expires time (default)
	expiresAt := time.Now().UTC().Add(defaultDraftTxExpiresIn)
	if config.ExpiresIn > 0 {
		expiresAt = time.Now().UTC().Add(config.ExpiresIn)
	}

	// Start the model
	draft := &DraftTransaction{
		Configuration:   *config,
		ExpiresAt:       expiresAt,
		Status:          DraftStatusDraft,
		TransactionBase: TransactionBase{ID: id},
		XpubID:          utils.Hash(rawXpubKey),
		Model: *NewBaseModel(
			ModelDraftTransaction,
			append(opts, WithXPub(rawXpubKey))...,
		),
	}

	// Set the fee (if not found) (if chainstate is loaded, use the first miner)
	// todo: make this more intelligent or allow the config to dictate the miner selection
	if config.FeeUnit == nil {
		if c := draft.Client(); c != nil {
			draft.Configuration.FeeUnit = c.Chainstate().FeeUnit()
		} else {
			draft.Configuration.FeeUnit = chainstate.DefaultFee
		}
	}
	return draft
}

// getDraftTransactionID will get the draft transaction with the given conditions
func getDraftTransactionID(ctx context.Context, xPubID, id string,
	opts ...ModelOps,
) (*DraftTransaction, error) {
	// Get the record
	config := &TransactionConfig{}
	conditions := map[string]interface{}{
		xPubIDField: xPubID,
		idField:     id,
	}
	if len(xPubID) == 0 {
		conditions = map[string]interface{}{
			idField: id,
		}
	}
	draftTransaction := newDraftTransaction("", config, opts...)
	draftTransaction.ID = "" // newDraftTransaction always sets an ID, need to remove for querying
	if err := Get(ctx, draftTransaction, conditions, false, defaultDatabaseReadTimeout, true); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return draftTransaction, nil
}

// getDraftTransactions will get all the draft transactions with the given conditions
func getDraftTransactions(ctx context.Context, metadata *Metadata, conditions *map[string]interface{},
	queryParams *datastore.QueryParams, opts ...ModelOps,
) ([]*DraftTransaction, error) {
	modelItems := make([]*DraftTransaction, 0)
	if err := getModelsByConditions(ctx, ModelDraftTransaction, &modelItems, metadata, conditions, queryParams, opts...); err != nil {
		return nil, err
	}

	return modelItems, nil
}

// getDraftTransactionsCount will get a count of all the access keys with the given conditions
func getDraftTransactionsCount(ctx context.Context, metadata *Metadata, conditions *map[string]interface{},
	opts ...ModelOps,
) (int64, error) {
	return getModelCountByConditions(ctx, ModelDraftTransaction, DraftTransaction{}, metadata, conditions, opts...)
}

// GetModelName will get the name of the current model
func (m *DraftTransaction) GetModelName() string {
	return ModelDraftTransaction.String()
}

// GetModelTableName will get the db table name of the current model
func (m *DraftTransaction) GetModelTableName() string {
	return tableDraftTransactions
}

// Save will save the model into the Datastore
func (m *DraftTransaction) Save(ctx context.Context) (err error) {
	if err = Save(ctx, m); err != nil {

		m.DebugLog("save tx error: " + err.Error())

		// todo: run in a go routine?
		// un-reserve the utxos
		if utxoErr := unReserveUtxos(
			ctx, m.XpubID, m.ID, m.GetOptions(false)...,
		); utxoErr != nil {
			err = errors.Wrap(err, utxoErr.Error())
		}
	}
	return
}

// GetID will get the model ID
func (m *DraftTransaction) GetID() string {
	return m.ID
}

// processConfigOutputs will process all the outputs,
// doing any lookups and creating locking scripts
func (m *DraftTransaction) processConfigOutputs(ctx context.Context) error {
	// Get the client
	c := m.Client()
	// Get sender's paymail
	paymailFrom := c.GetPaymailConfig().DefaultFromPaymail
	conditions := map[string]interface{}{
		xPubIDField: m.XpubID,
	}
	paymails, err := c.GetPaymailAddressesByXPubID(ctx, m.XpubID, nil, &conditions, nil)
	if err == nil && len(paymails) != 0 {
		paymailFrom = fmt.Sprintf("%s@%s", paymails[0].Alias, paymails[0].Domain)
	}
	// Special case where we are sending all funds to a single (address, paymail, handle)
	if m.Configuration.SendAllTo != nil {
		outputs := m.Configuration.Outputs

		m.Configuration.SendAllTo.UseForChange = true
		m.Configuration.SendAllTo.Satoshis = 0
		m.Configuration.Outputs = []*TransactionOutput{m.Configuration.SendAllTo}

		if err := m.Configuration.Outputs[0].processOutput(
			ctx, c.Cachestore(),
			c.PaymailClient(),
			paymailFrom,
			c.GetPaymailConfig().DefaultNote,
			false,
		); err != nil {
			return err
		}

		// re-add the other outputs we had before
		for _, output := range outputs {
			output.UseForChange = false // make sure we do not add change to this output
			if err := output.processOutput(
				ctx, c.Cachestore(),
				c.PaymailClient(),
				paymailFrom,
				c.GetPaymailConfig().DefaultNote,
				true,
			); err != nil {
				return err
			}
			m.Configuration.Outputs = append(m.Configuration.Outputs, output)
		}
	} else {
		// Loop all outputs and process
		for index := range m.Configuration.Outputs {

			// Start the output script slice
			if m.Configuration.Outputs[index].Scripts == nil {
				m.Configuration.Outputs[index].Scripts = make([]*ScriptOutput, 0)
			}

			// Process the outputs
			if err := m.Configuration.Outputs[index].processOutput(
				ctx, c.Cachestore(),
				c.PaymailClient(),
				paymailFrom,
				c.GetPaymailConfig().DefaultNote,
				true,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

// createTransactionHex will create the transaction with the given inputs and outputs
func (m *DraftTransaction) createTransactionHex(ctx context.Context) (err error) {
	// Check that we have outputs
	if len(m.Configuration.Outputs) == 0 && m.Configuration.SendAllTo == nil {
		return ErrMissingTransactionOutputs
	}

	// Get the total satoshis needed to make this transaction
	satoshisNeeded := m.getTotalSatoshis()

	// Set opts
	opts := m.GetOptions(false)

	// Process the outputs first
	// if an error occurs in processing the outputs, we have at least not made any reservations yet
	if err = m.processConfigOutputs(ctx); err != nil {
		return
	}

	var inputUtxos *[]*bt.UTXO
	var satoshisReserved uint64

	if m.Configuration.SendAllTo != nil { // Send TO ALL

		var spendableUtxos []*Utxo
		// todo should all utxos be sent to the SendAllTo address, not only the p2pkhs?
		if spendableUtxos, err = getSpendableUtxos(
			ctx, m.XpubID, utils.ScriptTypePubKeyHash, nil, m.Configuration.FromUtxos, opts...,
		); err != nil {
			return err
		}
		for _, utxo := range spendableUtxos {
			// Reserve the utxos
			utxo.DraftID.Valid = true
			utxo.DraftID.String = m.ID
			utxo.ReservedAt.Valid = true
			utxo.ReservedAt.Time = time.Now().UTC()

			// Save the UTXO
			if err = utxo.Save(ctx); err != nil {
				return err
			}

			m.Configuration.Outputs[0].Satoshis += utxo.Satoshis
		}

		// Get the inputUtxos (in bt.UTXO format) and the total amount of satoshis from the utxos
		if inputUtxos, satoshisReserved, err = m.getInputsFromUtxos(
			spendableUtxos,
		); err != nil {
			return
		}

		if err = m.processUtxos(
			ctx, spendableUtxos,
		); err != nil {
			return err
		}
	} else {

		// we can only include separate utxos (like tokens) when not using SendAllTo
		var includeUtxoSatoshis uint64
		if m.Configuration.IncludeUtxos != nil {
			includeUtxoSatoshis, err = m.addIncludeUtxos(ctx)
			if err != nil {
				return err
			}
		}

		// Reserve and Get utxos for the transaction
		var reservedUtxos []*Utxo
		feePerByte := float64(m.Configuration.FeeUnit.Satoshis / m.Configuration.FeeUnit.Bytes)

		reserveSatoshis := satoshisNeeded + m.estimateFee(m.Configuration.FeeUnit, 0)
		if reserveSatoshis <= dustLimit && !m.containsOpReturn() {
			m.client.Logger().Error(ctx, "amount of satoshis to send less than the dust limit")
			return ErrOutputValueTooLow
		}
		if reservedUtxos, err = reserveUtxos(
			ctx, m.XpubID, m.ID, reserveSatoshis, feePerByte, m.Configuration.FromUtxos, opts...,
		); err != nil {
			return
		}

		// Get the inputUtxos (in bt.UTXO format) and the total amount of satoshis from the utxos
		if inputUtxos, satoshisReserved, err = m.getInputsFromUtxos(
			reservedUtxos,
		); err != nil {
			return
		}

		// add the satoshis from the utxos we forcibly included to the total input sats
		satoshisReserved += includeUtxoSatoshis

		// Reserve the utxos
		if err = m.processUtxos(
			ctx, reservedUtxos,
		); err != nil {
			return err
		}
	}

	// Start a new transaction from the reservedUtxos
	tx := bt.NewTx()
	if err = tx.FromUTXOs(*inputUtxos...); err != nil {
		return
	}

	// Estimate the fee for the transaction
	fee := m.estimateFee(m.Configuration.FeeUnit, 0)
	if m.Configuration.SendAllTo != nil {
		if m.Configuration.Outputs[0].Satoshis <= dustLimit {
			return ErrOutputValueTooLow
		}

		m.Configuration.Fee = fee
		m.Configuration.Outputs[0].Satoshis -= fee

		// subtract all the satoshis sent in other outputs
		for _, output := range m.Configuration.Outputs {
			if !output.UseForChange { // only normal outputs
				m.Configuration.Outputs[0].Satoshis -= output.Satoshis
			}
		}

		m.Configuration.Outputs[0].Scripts[0].Satoshis = m.Configuration.Outputs[0].Satoshis
	} else {
		if satoshisReserved < satoshisNeeded+fee {
			return ErrNotEnoughUtxos
		}

		// if we have a remainder, add that to an output to our own wallet address
		satoshisChange := satoshisReserved - satoshisNeeded - fee
		m.Configuration.Fee = fee
		if satoshisChange > 0 {
			var newFee uint64
			newFee, err = m.setChangeDestination(
				ctx, satoshisChange, fee,
			)
			if err != nil {
				return
			}
			m.Configuration.Fee = newFee
		}
	}

	// Add the outputs to the bt transaction
	if err = m.addOutputsToTx(tx); err != nil {
		return
	}

	// final sanity check
	inputValue := uint64(0)
	usedUtxos := make([]string, 0)
	bumps := make(map[uint64][]BUMP)
	for _, input := range m.Configuration.Inputs {
		// check whether an utxo was used twice, this is not valid
		if utils.StringInSlice(input.Utxo.ID, usedUtxos) {
			return ErrDuplicateUTXOs
		}
		usedUtxos = append(usedUtxos, input.Utxo.ID)
		inputValue += input.Satoshis
		tx, err := m.client.GetTransactionByID(ctx, input.UtxoPointer.TransactionID)
		if err != nil {
			return err
		}
		if len(tx.BUMP.Path) != 0 {
			bumps[tx.BlockHeight] = append(bumps[tx.BlockHeight], tx.BUMP)
		}
	}
	outputValue := uint64(0)
	for _, output := range m.Configuration.Outputs {
		outputValue += output.Satoshis
	}

	if inputValue < outputValue {
		return ErrOutputValueTooHigh
	}
	if m.Configuration.Fee <= 0 {
		return ErrTransactionFeeInvalid
	}
	if inputValue-outputValue != m.Configuration.Fee {
		return ErrTransactionFeeInvalid
	}
	for _, b := range bumps {
		bump, err := CalculateMergedBUMP(b)
		if err != nil {
			return fmt.Errorf("Error while calculating Merged BUMP: %s", err.Error())
		}
		if bump == nil {
			continue
		}
		m.BUMPs = append(m.BUMPs, bump)
	}
	// Create the final hex (without signatures)
	m.Hex = tx.String()

	return
}

// addIncludeUtxos will add the included utxos
func (m *DraftTransaction) addIncludeUtxos(ctx context.Context) (uint64, error) {
	// Whatever utxos are selected, the IncludeUtxos should be added to the transaction
	// This can be used to add for instance tokens where fees need to be paid from other utxos
	// The satoshis of these inputs are not added to the reserved satoshis. If these inputs contain satoshis
	// that will be added to the total inputs and handled with the change addresses.
	includeUtxos := make([]*Utxo, 0)
	opts := m.GetOptions(false)
	var includeUtxoSatoshis uint64
	for _, utxo := range m.Configuration.IncludeUtxos {
		utxoModel, err := getUtxo(ctx, utxo.TransactionID, utxo.OutputIndex, opts...)
		if err != nil {
			return 0, err
		} else if utxoModel == nil {
			return 0, ErrMissingUtxo
		}
		includeUtxos = append(includeUtxos, utxoModel)
		includeUtxoSatoshis += utxoModel.Satoshis
	}
	return includeUtxoSatoshis, m.processUtxos(ctx, includeUtxos)
}

// processUtxos will process the utxos
func (m *DraftTransaction) processUtxos(ctx context.Context, utxos []*Utxo) error {
	// Get destinations
	opts := m.GetOptions(false)
	for _, utxo := range utxos {
		lockingScript := utils.GetDestinationLockingScript(utxo.ScriptPubKey)
		destination, err := getDestinationWithCache(
			ctx, m.Client(), "", "", lockingScript, opts...,
		)
		if err != nil {
			return err
		}
		if destination == nil {
			return ErrMissingDestination
		}
		m.Configuration.Inputs = append(
			m.Configuration.Inputs, &TransactionInput{
				Utxo:        *utxo,
				Destination: *destination,
			})
	}

	return nil
}

// estimateSize will loop the inputs and outputs and estimate the size of the transaction
func (m *DraftTransaction) estimateSize() uint64 {
	size := defaultOverheadSize // version + nLockTime

	inputSize := bt.VarInt(len(m.Configuration.Inputs))
	size += uint64(inputSize.Length())

	for _, input := range m.Configuration.Inputs {
		size += utils.GetInputSizeForType(input.Type)
	}

	outputSize := bt.VarInt(len(m.Configuration.Outputs))
	size += uint64(outputSize.Length())

	for _, output := range m.Configuration.Outputs {
		for _, s := range output.Scripts {
			size += utils.GetOutputSize(s.Script)
		}
	}

	return size
}

// estimateFee will loop the inputs and outputs and estimate the required fee
func (m *DraftTransaction) estimateFee(unit *utils.FeeUnit, addToSize uint64) uint64 {
	size := m.estimateSize() + addToSize
	feeEstimate := float64(size) * (float64(unit.Satoshis) / float64(unit.Bytes))
	return uint64(math.Ceil(feeEstimate))
}

// addOutputs will add the given outputs to the bt.Tx
func (m *DraftTransaction) addOutputsToTx(tx *bt.Tx) (err error) {
	var s *bscript.Script
	for _, output := range m.Configuration.Outputs {
		for _, sc := range output.Scripts {
			if s, err = bscript.NewFromHexString(
				sc.Script,
			); err != nil {
				return
			}

			scriptType := sc.ScriptType
			if scriptType == "" {
				scriptType = utils.GetDestinationType(sc.Script)
			}

			if scriptType == utils.ScriptTypeNullData {
				// op_return output - only one allowed to have 0 satoshi value ???
				if sc.Satoshis > 0 {
					return ErrInvalidOpReturnOutput
				}

				tx.AddOutput(&bt.Output{
					LockingScript: s,
					Satoshis:      0,
				})
			} else if scriptType == utils.ScriptTypePubKeyHash {
				// sending to a p2pkh
				if sc.Satoshis == 0 {
					return ErrOutputValueTooLow
				}

				if err = tx.AddP2PKHOutputFromScript(
					s, sc.Satoshis,
				); err != nil {
					return
				}
			} else {
				// add non-standard output script
				tx.AddOutput(&bt.Output{
					LockingScript: s,
					Satoshis:      sc.Satoshis,
				})
			}
		}
	}
	return
}

// setChangeDestination will make a new change destination
func (m *DraftTransaction) setChangeDestination(ctx context.Context, satoshisChange uint64, fee uint64) (uint64, error) {
	m.Configuration.ChangeSatoshis = satoshisChange

	useExistingOutputsForChange := make([]int, 0)
	for index := range m.Configuration.Outputs {
		if m.Configuration.Outputs[index].UseForChange {
			useExistingOutputsForChange = append(useExistingOutputsForChange, index)
		}
	}

	newFee := fee
	if len(useExistingOutputsForChange) > 0 {
		// reset destinations if set
		m.Configuration.ChangeDestinationsStrategy = ChangeStrategyDefault
		m.Configuration.ChangeDestinations = nil

		numberOfExistingOutputs := uint64(len(useExistingOutputsForChange))
		changePerOutput := uint64(float64(satoshisChange) / float64(numberOfExistingOutputs))
		remainderOutput := satoshisChange - (changePerOutput * numberOfExistingOutputs)
		for _, outputIndex := range useExistingOutputsForChange {
			m.Configuration.Outputs[outputIndex].Satoshis += changePerOutput + remainderOutput
			remainderOutput = 0 // reset remainder to 0 for other outputs
		}
	} else {
		numberOfDestinations := m.Configuration.ChangeNumberOfDestinations
		if numberOfDestinations <= 0 {
			numberOfDestinations = 1 // todo get from config
		}
		minimumSatoshis := m.Configuration.ChangeMinimumSatoshis
		if minimumSatoshis <= 0 { // todo: protect against un-spendable amount? less than fee to miner for min tx?
			minimumSatoshis = 1250 // todo get from config
		}

		if float64(satoshisChange)/float64(numberOfDestinations) < float64(minimumSatoshis) {
			// we cannot split our change to the number of destinations given, re-calc
			numberOfDestinations = 1
		}

		newFee = m.estimateFee(m.Configuration.FeeUnit, uint64(numberOfDestinations)*changeOutputSize)
		satoshisChange -= newFee - fee
		m.Configuration.ChangeSatoshis = satoshisChange

		if m.Configuration.ChangeDestinations == nil {
			if err := m.setChangeDestinations(
				ctx, numberOfDestinations,
			); err != nil {
				return fee, err
			}
		}

		changeSatoshis, err := m.getChangeSatoshis(satoshisChange)
		if err != nil {
			return fee, err
		}

		for _, destination := range m.Configuration.ChangeDestinations {
			m.Configuration.Outputs = append(m.Configuration.Outputs, &TransactionOutput{
				To: destination.Address,
				Scripts: []*ScriptOutput{{
					Address:    destination.Address,
					Satoshis:   changeSatoshis[destination.LockingScript],
					Script:     destination.LockingScript,
					ScriptType: utils.ScriptTypePubKeyHash,
				}},
				Satoshis: changeSatoshis[destination.LockingScript],
			})
		}
	}

	return newFee, nil
}

// split the change satoshis amongst the change destinations according to the strategy given in config
func (m *DraftTransaction) getChangeSatoshis(satoshisChange uint64) (changeSatoshis map[string]uint64, err error) {
	changeSatoshis = make(map[string]uint64)
	var lastDestination string
	changeUsed := uint64(0)

	if m.Configuration.ChangeDestinationsStrategy == ChangeStrategyNominations {
		return nil, ErrChangeStrategyNotImplemented
	} else if m.Configuration.ChangeDestinationsStrategy == ChangeStrategyRandom {
		nDestinations := float64(len(m.Configuration.ChangeDestinations))
		var a *big.Int
		for _, destination := range m.Configuration.ChangeDestinations {
			if a, err = rand.Int(
				rand.Reader, big.NewInt(math.MaxInt64),
			); err != nil {
				return
			}
			randomChange := (((float64(a.Int64()) / (1 << 63)) * 50) + 75) / 100
			changeForDestination := uint64(randomChange * float64(satoshisChange) / nDestinations)

			changeSatoshis[destination.LockingScript] = changeForDestination
			lastDestination = destination.LockingScript
			changeUsed += changeForDestination
		}
	} else {
		// default
		changePerDestination := uint64(float64(satoshisChange) / float64(len(m.Configuration.ChangeDestinations)))
		for _, destination := range m.Configuration.ChangeDestinations {
			changeSatoshis[destination.LockingScript] = changePerDestination
			lastDestination = destination.LockingScript
			changeUsed += changePerDestination
		}
	}

	// handle remainder
	changeSatoshis[lastDestination] += satoshisChange - changeUsed

	return
}

// setChangeDestinations will set the change destinations based on the number
func (m *DraftTransaction) setChangeDestinations(ctx context.Context, numberOfDestinations int) error {
	// Set the options
	opts := m.GetOptions(false)
	optsNew := append(opts, New())
	c := m.Client()

	var err error
	var xPub *Xpub
	var num uint32

	// Loop for each destination
	for i := 0; i < numberOfDestinations; i++ {
		if xPub, err = getXpubWithCache(
			ctx, c, m.rawXpubKey, "", opts...,
		); err != nil {
			return err
		} else if xPub == nil {
			return ErrMissingXpub
		}

		if num, err = xPub.incrementNextNum(
			ctx, utils.ChainInternal,
		); err != nil {
			return err
		}

		var destination *Destination
		if destination, err = newAddress(
			m.rawXpubKey, utils.ChainInternal, num, optsNew...,
		); err != nil {
			return err
		}

		destination.DraftID = m.ID
		if err = destination.Save(ctx); err != nil {
			return err
		}

		m.Configuration.ChangeDestinations = append(m.Configuration.ChangeDestinations, destination)
	}

	return nil
}

// getInputsFromUtxos this function transforms bux utxos to bt.UTXOs
func (m *DraftTransaction) getInputsFromUtxos(reservedUtxos []*Utxo) (*[]*bt.UTXO, uint64, error) {
	// transform to bt.utxo and check if we have enough
	inputUtxos := new([]*bt.UTXO)
	satoshisReserved := uint64(0)
	var lockingScript *bscript.Script
	var err error
	for _, utxo := range reservedUtxos {

		if lockingScript, err = bscript.NewFromHexString(
			utxo.ScriptPubKey,
		); err != nil {
			return nil, 0, errors.Wrap(ErrInvalidLockingScript, err.Error())
		}

		var txIDBytes []byte
		if txIDBytes, err = hex.DecodeString(
			utxo.TransactionID,
		); err != nil {
			return nil, 0, errors.Wrap(ErrInvalidTransactionID, err.Error())
		}

		*inputUtxos = append(*inputUtxos, &bt.UTXO{
			TxID:           txIDBytes,
			Vout:           utxo.OutputIndex,
			Satoshis:       utxo.Satoshis,
			LockingScript:  lockingScript,
			SequenceNumber: bt.DefaultSequenceNumber,
		})
		satoshisReserved += utxo.Satoshis
	}

	return inputUtxos, satoshisReserved, nil
}

// getTotalSatoshis calculate the total satoshis of all outputs
func (m *DraftTransaction) getTotalSatoshis() (satoshis uint64) {
	for _, output := range m.Configuration.Outputs {
		satoshis += output.Satoshis
	}
	return
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *DraftTransaction) BeforeCreating(ctx context.Context) (err error) {
	m.DebugLog("starting: " + m.Name() + " BeforeCreating hook...")

	// Prepare the transaction
	if err = m.createTransactionHex(ctx); err != nil {
		return
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return
}

// AfterUpdated will fire after a successful update into the Datastore
func (m *DraftTransaction) AfterUpdated(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterUpdated hook...")

	// todo: run these in go routines?

	// remove reservation from all utxos related to this draft transaction
	if m.Status == DraftStatusCanceled || m.Status == DraftStatusExpired {
		utxos, err := getUtxosByDraftID(
			ctx, m.ID,
			nil,
			m.GetOptions(false)...,
		)
		if err != nil {
			return err
		}
		for index := range utxos {
			utxos[index].DraftID.String = ""
			utxos[index].DraftID.Valid = false
			utxos[index].ReservedAt.Time = time.Time{}
			utxos[index].ReservedAt.Valid = false
			if err = utxos[index].Save(ctx); err != nil {
				return err
			}
		}
	}

	m.DebugLog("end: " + m.Name() + " AfterUpdated hook")
	return nil
}

// RegisterTasks will register the model specific tasks on client initialization
func (m *DraftTransaction) RegisterTasks() error {
	// No task manager loaded?
	tm := m.Client().Taskmanager()
	if tm == nil {
		return nil
	}

	// Register the task locally (cron task - set the defaults)
	cleanUpTask := m.Name() + "_clean_up"
	ctx := context.Background()

	// Register the task
	if err := tm.RegisterTask(&taskmanager.Task{
		Name:       cleanUpTask,
		RetryLimit: 1,
		Handler: func(client ClientInterface) error {
			if taskErr := taskCleanupDraftTransactions(ctx, client.Logger(), WithClient(client)); taskErr != nil {
				client.Logger().Error(ctx, "error running "+cleanUpTask+" task: "+taskErr.Error())
			}
			return nil
		},
	}); err != nil {
		return err
	}

	// Run the task periodically
	return tm.RunTask(ctx, &taskmanager.TaskOptions{
		Arguments:      []interface{}{m.Client()},
		RunEveryPeriod: m.Client().GetTaskPeriod(cleanUpTask),
		TaskName:       cleanUpTask,
	})
}

// Migrate model specific migration on startup
func (m *DraftTransaction) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableDraftTransactions), metadataField)
}

// SignInputsWithKey will sign all the inputs using a key (string) (helper method)
func (m *DraftTransaction) SignInputsWithKey(xPrivKey string) (signedHex string, err error) {
	// Decode the xPriv using the key
	var xPriv *bip32.ExtendedKey
	if xPriv, err = bip32.NewKeyFromString(xPrivKey); err != nil {
		return
	}

	return m.SignInputs(xPriv)
}

// SignInputs will sign all the inputs using the given xPriv key
func (m *DraftTransaction) SignInputs(xPriv *bip32.ExtendedKey) (signedHex string, err error) {
	// Start a bt draft transaction
	var txDraft *bt.Tx
	if txDraft, err = bt.NewTxFromString(m.Hex); err != nil {
		return
	}

	// Sign the inputs
	for index, input := range m.Configuration.Inputs {

		// Get the locking script
		var ls *bscript.Script
		if ls, err = bscript.NewFromHexString(
			input.Destination.LockingScript,
		); err != nil {
			return
		}
		txDraft.Inputs[index].PreviousTxScript = ls
		txDraft.Inputs[index].PreviousTxSatoshis = input.Satoshis

		// Derive the child key (chain)
		var chainKey *bip32.ExtendedKey
		if chainKey, err = xPriv.Child(
			input.Destination.Chain,
		); err != nil {
			return
		}

		// Derive the child key (num)
		var numKey *bip32.ExtendedKey
		if numKey, err = chainKey.Child(
			input.Destination.Num,
		); err != nil {
			return
		}

		// Get the private key
		var privateKey *bec.PrivateKey
		if privateKey, err = bitcoin.GetPrivateKeyFromHDKey(
			numKey,
		); err != nil {
			return
		}

		// Get the unlocking script
		var s *bscript.Script
		if s, err = utils.GetUnlockingScript(
			txDraft, uint32(index), privateKey,
		); err != nil {
			return
		}

		// Insert the locking script
		if err = txDraft.InsertInputUnlockingScript(
			uint32(index), s,
		); err != nil {
			return
		}
	}

	// Return the signed hex
	signedHex = txDraft.String()
	return
}

func (m *DraftTransaction) containsOpReturn() bool {
	for _, output := range m.Configuration.Outputs {
		if output.OpReturn != nil {
			return true
		}
	}
	return false
}
