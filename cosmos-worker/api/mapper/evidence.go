package mapper

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	evidence "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/gogo/protobuf/proto"
)

// EvidenceSubmitEvidenceToSub transforms evidence.MsgSubmitEvidence sdk messages to SubsetEvent
func EvidenceSubmitEvidenceToSub(msg []byte) (se structs.SubsetEvent, er error) {
	mse := &evidence.MsgSubmitEvidence{}
	if err := proto.Unmarshal(msg, mse); err != nil {
		return se, fmt.Errorf("Not a submit_evidence type: %w", err)
	}

	se = structs.SubsetEvent{
		Type:   []string{"submit_evidence"},
		Module: "evidence",
		Node:   map[string][]structs.Account{"submitter": {{ID: mse.Submitter}}},
	}

	ev := mse.GetEvidence()
	if ev == nil {
		return se, errors.New("Evidence is empty")
	}

	se.Additional = map[string][]string{
		"evidence_height": {strconv.FormatInt(ev.GetHeight(), 10)},
	}

	evc := mse.Evidence.GetCachedValue()
	if evc == nil {
		return se, errors.New("Evidence is empty")
	}

	validatorEvi, ok := evc.(exported.ValidatorEvidence)
	if !ok {
		return se, errors.New("Evidence is not ValidatorEvidence type")
	}

	se.Additional["evidence_total_power"] = []string{strconv.FormatInt(validatorEvi.GetTotalPower(), 10)}
	se.Additional["evidence_validator_power"] = []string{strconv.FormatInt(validatorEvi.GetValidatorPower(), 10)}
	se.Additional["evidence_consensus"] = []string{validatorEvi.GetConsensusAddress().String()}
	return se, nil
}
