package bux

import (
	"strings"

	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoin-sv/go-paymail"
	"github.com/bitcoin-sv/go-paymail/server"
)

// processOutput will inspect the output to determine how to process
func (t *TransactionOutput) _processOutput(defaultFromSender, defaultNote string, checkSatoshis bool, destination *Destination) error {

	// Convert known handle formats ($handcash or 1relayx)
	if strings.Contains(t.To, handleHandcashPrefix) ||
		(len(t.To) < handleMaxLength && len(t.To) > 1 && t.To[:1] == handleRelayPrefix) {

		// Convert the handle and check if it's changed (becomes a paymail address)
		if p := paymail.ConvertHandle(t.To, false); p != t.To {
			t.To = p
		}
	}

	// Check for Paymail, Bitcoin Address or OP Return
	if len(t.To) > 0 && strings.Contains(t.To, "@") { // Paymail output
		if checkSatoshis && t.Satoshis <= 0 {
			return ErrOutputValueTooLow
		}
		return t._processPaymailOutput(defaultFromSender, defaultNote, destination)
	} else {
		panic("invalid paymail")
	}
}

// processPaymailOutput will detect how to process the Paymail output given
func (t *TransactionOutput) _processPaymailOutput(fromPaymail, defaultNote string, destination *Destination) error {

	// Standardize the paymail address (break into parts)
	alias, domain, paymailAddress := paymail.SanitizePaymail(t.To)
	if len(paymailAddress) == 0 {
		return ErrPaymailAddressIsInvalid
	}

	// Set the sanitized version of the paymail address provided
	t.To = paymailAddress

	// Start setting the Paymail information (nil check might not be needed)
	if t.PaymailP4 == nil {
		t.PaymailP4 = &PaymailP4{
			Alias:  alias,
			Domain: domain,
		}
	} else {
		t.PaymailP4.Alias = alias
		t.PaymailP4.Domain = domain
	}

	// Does the provider support P2P?

	return t._processPaymailViaP2P(fromPaymail, BeefPaymailPayloadFormat, defaultNote, destination)
}

func (t *TransactionOutput) _processPaymailViaP2P(fromPaymail string, format PaymailPayloadFormat, note string, destination *Destination) error {

	// todo: this is a hack since paymail providers will complain if satoshis are empty (SendToAll has 0 satoshi)
	satoshis := t.Satoshis
	if satoshis <= 0 {
		satoshis = 100
	}

	// Get the outputs and destination information from the Paymail provider
	// destinationInfo, err := startP2PTransaction(
	// 	client, t.PaymailP4.Alias, t.PaymailP4.Domain,
	// 	p2pDestinationURL, satoshis,
	// )
	md := &server.RequestMetadata{
		Alias:      t.PaymailP4.Alias,
		Domain:     t.PaymailP4.Domain,
		IPAddress:  "0.0.0.0",
		Note:       note,
		RequestURI: "http://uri",
		UserAgent:  "beef-test-4chain",
	}
	destinationInfo, err := _createP2PDestinationResponse(md.Alias, md.Domain, t.Satoshis, md, destination)

	if err != nil {
		return err
	}

	// split the total output satoshis across all the paymail outputs given
	outputValues, err := utils.SplitOutputValues(satoshis, len(destinationInfo.Outputs))
	if err != nil {
		return err
	}

	// Loop all received P2P outputs and build scripts
	for index, out := range destinationInfo.Outputs {
		t.Scripts = append(
			t.Scripts,
			&ScriptOutput{
				Address:    out.Address,
				Satoshis:   outputValues[index],
				Script:     out.Script,
				ScriptType: utils.ScriptTypePubKeyHash,
			},
		)
	}

	// Set the remaining P2P information
	//t.PaymailP4.ReceiveEndpoint = p2pSubmitTxURL
	t.PaymailP4.ReferenceID = destinationInfo.Reference
	t.PaymailP4.ResolutionType = ResolutionTypeP2P
	t.PaymailP4.FromPaymail = fromPaymail
	t.PaymailP4.Format = format

	return nil
}

// CreateP2PDestinationResponse will create a p2p destination response
func _createP2PDestinationResponse(alias, domain string,
	satoshis uint64, requestMetadata *server.RequestMetadata, destination *Destination) (*paymail.PaymentDestinationPayload, error) {

	referenceID, err := utils.RandomHex(16)
	if err != nil {
		return nil, err
	}

	metadata := _createMetadata(requestMetadata, "CreateP2PDestinationResponse")
	metadata[ReferenceIDField] = referenceID
	metadata[satoshisField] = satoshis

	// Append the output(s)
	var outputs []*paymail.PaymentOutput
	outputs = append(outputs, &paymail.PaymentOutput{
		Address:  destination.Address,
		Satoshis: satoshis,
		Script:   destination.LockingScript,
	})

	return &paymail.PaymentDestinationPayload{
		Outputs:   outputs,
		Reference: referenceID,
	}, nil
}

func _createMetadata(serverMetaData *server.RequestMetadata, request string) (metadata Metadata) {
	metadata = make(Metadata)
	metadata["paymail_request"] = request

	if serverMetaData != nil {
		if serverMetaData.UserAgent != "" {
			metadata["user_agent"] = serverMetaData.UserAgent
		}
		if serverMetaData.Note != "" {
			metadata["note"] = serverMetaData.Note
		}
		if serverMetaData.Domain != "" {
			metadata[domainField] = serverMetaData.Domain
		}
		if serverMetaData.IPAddress != "" {
			metadata["ip_address"] = serverMetaData.IPAddress
		}
	}
	return
}
