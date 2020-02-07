package sync

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/cache"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
)

func (r *Service) committeeIndexBeaconAttestationSubscriber(ctx context.Context, msg proto.Message) error {
	a, ok := msg.(*eth.Attestation)
	if !ok {
		return fmt.Errorf("message was not type *eth.Attestation, type=%T", msg)
	}

	if exists, _ := r.attPool.HasAggregatedAttestation(a); exists {
		return nil
	}

	return r.attPool.SaveUnaggregatedAttestation(a)
}

func (r *Service) currentCommitteeIndex() []uint64 {
	return cache.TrackedCommitteeIndices.GetIndices(helpers.SlotToEpoch(r.chain.HeadSlot()))
}
