package client

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	originClient "github.com/prysmaticlabs/prysm/validator/client"
	"github.com/prysmaticlabs/prysm/validator/keymanager"
	"google.golang.org/grpc"
)

func NewValidator(keyManager keymanager.KeyManager, conn *grpc.ClientConn) originClient.Validator {
	v := originClient.NewValidatorImplementation(
		ethpb.NewBeaconNodeValidatorClient(conn),
		ethpb.NewBeaconChainClient(conn),
		ethpb.NewNodeClient(conn),
		keyManager,
	)
	return &v
}
