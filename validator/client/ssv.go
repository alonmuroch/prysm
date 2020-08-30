package client

import (
	"context"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"go.opencensus.io/trace"
	"io"
)

func (v *validator) NextTask(ctx context.Context)  (<- chan *ethpb.SSVTask, error) {
	ctx, span := trace.StartSpan(ctx, "validator.SSVTaskStream")
	defer span.End()

	stream, error := v.ssvClient.GetTaskStream(ctx, &ethpb.StreamRequest{
		PublicKey:            nil,
		Topics:               []ethpb.StreamTopics{
			ethpb.StreamTopics_SIGN_BLOCK,
			ethpb.StreamTopics_CHECK_BLOCK,
			ethpb.StreamTopics_SIGN_ATTESTATION,
			ethpb.StreamTopics_CHECK_ATTESTATION,
			ethpb.StreamTopics_SIGN_AGGREGATION,
		},
	})

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
