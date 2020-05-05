package client

import (
	"context"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/slotutil"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"io"
	"time"
)

type SlotTicker struct {
	ticker          *slotutil.SlotTicker
	genesisTime     uint64
	validatorClient ethpb.BeaconNodeValidatorClient
}

func NewSlotTicker(conn *grpc.ClientConn) *SlotTicker {
	validatorClient := ethpb.NewBeaconNodeValidatorClient(conn)

	return &SlotTicker{
		ticker:          nil,
		genesisTime:     0,
		validatorClient: validatorClient,
	}
}

func (ticker *SlotTicker) Start(ctx context.Context) error {
	ctx, span := trace.StartSpan(ctx, "validator.WaitForChainStart")
	defer span.End()

	// First, check if the beacon chain has started.
	stream, err := ticker.validatorClient.WaitForChainStart(ctx, &ptypes.Empty{})
	if err != nil {
		return errors.Wrap(err, "could not setup beacon chain ChainStart streaming client")

	}
	for {
		log.Info("Waiting for beacon chain start log from the ETH 1.0 deposit contract")
		chainStartRes, err := stream.Recv()
		// If the stream is closed, we stop the loop.
		if err == io.EOF {
			break
		}
		// If context is canceled we stop the loop.
		if ctx.Err() == context.Canceled {
			return errors.Wrap(ctx.Err(), "context has been canceled so shutting down the loop")
		}
		if err != nil {
			return errors.Wrap(err, "could not receive ChainStart from stream")
		}
		ticker.genesisTime = chainStartRes.GenesisTime
		break
	}
	// Once the ChainStart log is received, we update the genesis time of the validator client
	// and begin a slot ticker used to track the current slot the beacon node is in.
	ticker.ticker = slotutil.GetSlotTicker(time.Unix(int64(ticker.genesisTime), 0), params.BeaconConfig().SecondsPerSlot)
	log.WithField("genesisTime", time.Unix(int64(ticker.genesisTime), 0)).Info("Beacon chain started")
	return nil
}

func (ticker *SlotTicker) NextSlot() <-chan uint64 {
	return ticker.ticker.C()
}

// SlotDeadline is the start time of the next slot.
func (ticker *SlotTicker) SlotDeadline(slot uint64) time.Time {
	secs := (slot + 1) * params.BeaconConfig().SecondsPerSlot
	return time.Unix(int64(ticker.genesisTime), 0 /*ns*/).Add(time.Duration(secs) * time.Second)
}
