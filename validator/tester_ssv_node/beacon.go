package main

import (
	"context"
	"github.com/gogo/protobuf/types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	log "github.com/sirupsen/logrus"
)

func (n *SSVNode) GetDuties(context.Context, *ethpb.DutiesRequest) (*ethpb.DutiesResponse, error) {
	return nil,nil
}

func (n *SSVNode) StreamDuties(*ethpb.DutiesRequest, ethpb.BeaconNodeValidator_StreamDutiesServer) error {
	return nil
}

func (n *SSVNode) DomainData(context.Context, *ethpb.DomainRequest) (*ethpb.DomainResponse, error) {
	return &ethpb.DomainResponse{
			SignatureDomain: make([]byte,32),
		},nil
}

func (n *SSVNode) WaitForChainStart(*types.Empty, ethpb.BeaconNodeValidator_WaitForChainStartServer) error {
	return nil
}

func (n *SSVNode) WaitForSynced(*types.Empty, ethpb.BeaconNodeValidator_WaitForSyncedServer) error {
	return nil
}

func (n *SSVNode) WaitForActivation(*ethpb.ValidatorActivationRequest, ethpb.BeaconNodeValidator_WaitForActivationServer) error {
	return nil
}

func (n *SSVNode) ValidatorIndex(context.Context, *ethpb.ValidatorIndexRequest) (*ethpb.ValidatorIndexResponse, error) {
	return nil,nil
}

func (n *SSVNode) ValidatorStatus(context.Context, *ethpb.ValidatorStatusRequest) (*ethpb.ValidatorStatusResponse, error) {
	return nil,nil
}

func (n *SSVNode) MultipleValidatorStatus(context.Context, *ethpb.MultipleValidatorStatusRequest) (*ethpb.MultipleValidatorStatusResponse, error) {
	return nil,nil
}

func (n *SSVNode) GetBlock(context.Context, *ethpb.BlockRequest) (*ethpb.BeaconBlock, error) {
	return nil,nil
}

func (n *SSVNode) ProposeBlock(context.Context, *ethpb.SignedBeaconBlock) (*ethpb.ProposeResponse, error) {
	return nil,nil
}

func (n *SSVNode) GetAttestationData(context.Context, *ethpb.AttestationDataRequest) (*ethpb.AttestationData, error) {
	return nil,nil
}

func (n *SSVNode) ProposeAttestation(context.Context, *ethpb.Attestation) (*ethpb.AttestResponse, error) {
	log.Printf("Received partial attestation")
	return &ethpb.AttestResponse{
		AttestationDataRoot:  make([]byte,96),
	},nil
}

func (n *SSVNode) SubmitAggregateSelectionProof(context.Context, *ethpb.AggregateSelectionRequest) (*ethpb.AggregateSelectionResponse, error) {
	return nil,nil
}

func (n *SSVNode) SubmitSignedAggregateSelectionProof(context.Context, *ethpb.SignedAggregateSubmitRequest) (*ethpb.SignedAggregateSubmitResponse, error) {
	return nil,nil
}

func (n *SSVNode) ProposeExit(context.Context, *ethpb.SignedVoluntaryExit) (*ethpb.ProposeExitResponse, error) {
	return nil,nil
}

func (n *SSVNode) SubscribeCommitteeSubnets(context.Context, *ethpb.CommitteeSubnetsSubscribeRequest) (*types.Empty, error) {
	return nil,nil
}

