package client

import (
	"context"
	"encoding/hex"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
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
	sig, err := v.signAtt(ctx, pubKey, data)
	if err != nil {
		log.WithError(err).Error("Could not sign partial attestation")
	}

	attestation := &ethpb.Attestation{
		Data:            data,
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