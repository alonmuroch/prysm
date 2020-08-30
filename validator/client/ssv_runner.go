package client

import (
	"context"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"go.opencensus.io/trace"
)

type SSVValidator interface {
	NextTask(ctx context.Context) (<- chan *ethpb.SSVTask, error)
	FetchSignerPubKeys(ctx context.Context) ([][]byte,error)
	SignPartialAttestation(ctx context.Context, data *ethpb.AttestationData, pubKey [48]byte)
	SignPartialBlock(ctx context.Context, block *ethpb.BeaconBlock, pubKey [48]byte)
}

// runs the main loop for an SSV client
func runSSVClient(ctx context.Context, v SSVValidator) {
	log.Printf("SSV client started")

	taskStreams,err := v.NextTask(ctx)
	if err != nil {
		log.Fatalf("Could not fetch SSV task stream: %v", err)
	}

	for {
		ctx, span := trace.StartSpan(ctx, "validator.processSSVTask")

		select {
		case <-ctx.Done():
			log.Info("Context canceled, stopping validator")
			return // Exit if context is canceled.
		case task := <- taskStreams:
			span.AddAttributes(trace.StringAttribute("task", task.Topic.String()))
			if task.Topic == ethpb.StreamTopics_SIGN_ATTESTATION {
				go v.SignPartialAttestation(ctx, task.GetAttestation(), bytesutil.ToBytes48(task.GetPublicKey()))
			}
			if task.Topic == ethpb.StreamTopics_SIGN_BLOCK {
				go v.SignPartialBlock(ctx, task.GetBlock(), bytesutil.ToBytes48(task.GetPublicKey()))
			}
			if task.Topic == ethpb.StreamTopics_SIGN_AGGREGATION {

			}

			span.End()
		}
	}
}
