package client

import (
	"context"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"go.opencensus.io/trace"
)

// runs the main loop for an SSV client
func runSSVClient(ctx context.Context, v Validator) {
	log.Printf("SSV started")

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
			if task.Topic == eth.StreamTopics_SIGN_ATTESTATION {

			}
			if task.Topic == eth.StreamTopics_SIGN_BLOCK {

			}
			if task.Topic == eth.StreamTopics_SIGN_AGGREGATION {

			}

			span.End()
		}
	}
}
