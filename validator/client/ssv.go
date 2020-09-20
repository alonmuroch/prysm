package client

import (
	"context"
	"encoding/hex"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/params"
	"go.opencensus.io/trace"
	"io"
)

func (v *validator) NextTask(ctx context.Context) (<- chan *ethpb.SSVTask, error) {
	ctx, span := trace.StartSpan(ctx, "validator.SSVTaskStream")
	defer span.End()

	pubkeys, err := v.FetchSignerPubKeys(ctx)
	if err != nil {
		log.Fatalf("Could not fetch keys: %s", err.Error())
	}

	stream, error := v.ssvClient.GetTaskStream(ctx, &ethpb.StreamRequest{
		Topics:               []ethpb.StreamTopics{
			ethpb.StreamTopics_SIGN_BLOCK,
			ethpb.StreamTopics_CHECK_BLOCK,
			ethpb.StreamTopics_SIGN_ATTESTATION,
			ethpb.StreamTopics_CHECK_ATTESTATION,
			ethpb.StreamTopics_SIGN_AGGREGATION,
		},
		PublicKeys: pubkeys,
	})

	for _, pubkey := range pubkeys {
		log.Printf("Connected to SSV node streaming with pubkey: %s", hex.EncodeToString(pubkey))
	}


	if error != nil {
		return nil, error
	}

	ret := make(chan *ethpb.SSVTask)
	go func() {
		for {
			task, err := stream.Recv()
			if err == io.EOF {
				// TODO
			}
			if err != nil {
				return // TODO
			}

			ret <- task
		}
	}()
	return ret, nil
}

func (v *validator) FetchSignerPubKeys(ctx context.Context) ([][]byte,error) {
	keys, err := v.keyManagerV2.FetchValidatingPublicKeys(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([][]byte, len(keys))
	for i := range keys {
		ret[i] = bytesutil.FromBytes48(keys[i])
	}
	return ret, nil
}

// An SSV specific function to sign an attestation as one participant of many
func (v *validator) SignPartialAttestation(ctx context.Context, data *ethpb.AttestationData, pubKey [48]byte) {
	log.Printf("%s", data)
	sig, err := v.signAtt(ctx, pubKey, data)
	if err != nil {
		log.WithError(err).Error("Could not sign partial attestation")
	}

	// get duty to retreive committee
	// TODO - maybe caan get from the SSV node
	req := &ethpb.DutiesRequest{
		Epoch:      data.Slot / params.BeaconConfig().SlotsPerEpoch,
		PublicKeys: [][]byte{pubKey[:]},
	}
	resp, err := v.validatorClient.GetDuties(ctx, req)
	if err != nil {
		log.WithError(err).Error("could not get duty")
	}
	duty := resp.GetDuties()[0] // TODO - should not be hard coded

	// find index in committee
	var indexInCommittee uint64
	var found bool
	for i, vID := range duty.Committee {
		if vID == duty.ValidatorIndex {
			indexInCommittee = uint64(i)
			found = true
			break
		}
	}
	if !found {
		log.Errorf("Validator ID %d not found in committee of %v", duty.ValidatorIndex, duty.Committee)
		return
	}

	// build attestation object
	aggregationBitfield := bitfield.NewBitlist(uint64(len(duty.Committee)))
	aggregationBitfield.SetBitAt(indexInCommittee, true)
	attestation := &ethpb.Attestation{
		Data:            data,
		AggregationBits: aggregationBitfield,
		Signature:       sig,
	}

	_, err = v.validatorClient.ProposeAttestation(ctx, attestation)
	if err != nil {
		log.WithError(err).Error("Could not submit partial attestation to SSV node")
	} else {
		log.Printf("Signed and proposed partial attestation")
	}
}

func (v *validator) SignPartialBlock(ctx context.Context, block *ethpb.BeaconBlock, pubKey [48]byte) {
	epoch := block.Slot / params.BeaconConfig().SlotsPerEpoch
	sig, err := v.signBlock(ctx, pubKey, epoch, block)
	if err != nil {
		log.WithError(err).Error("Could not sign partial block")
	}

	blk := &ethpb.SignedBeaconBlock{
		Block:     block,
		Signature: sig,
	}
	_, err = v.validatorClient.ProposeBlock(ctx, blk)
	if err != nil {
		log.WithError(err).Error("Could not submit partial block to SSV node")
	}
}