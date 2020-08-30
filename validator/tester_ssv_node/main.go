package main

import (
	"fmt"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/slotutil"
	"github.com/prysmaticlabs/prysm/shared/version"
	"github.com/prysmaticlabs/prysm/validator/flags"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"
)

type SSVNode struct {
	ctx 			*cli.Context
	ticker 			*slotutil.SlotTicker
	pubKeys 		[][48]byte
}

func NewAndStartNode(ctx *cli.Context) error {
	port := ctx.String(flags.RPCPort.Name)
	serveOn := fmt.Sprintf(":%s", port)

	node := &SSVNode{
		ctx:     ctx,
		ticker:  slotutil.GetSlotTicker(time.Now(), params.BeaconConfig().SecondsPerSlot),
		pubKeys: nil,
	}

	grpcServer := grpc.NewServer()
	ethpb.RegisterSSVServer(grpcServer, node)
	ethpb.RegisterBeaconNodeValidatorServer(grpcServer, node)

	lis, err := net.Listen("tcp", serveOn)
	if err != nil {
		return err
	}

	node.Start()

	log.Printf("starting SSV server: %s", serveOn)
	err = grpcServer.Serve(lis)
	if err !=nil {
		log.Fatalf("could not start server: %s",err.Error())
	}

	return nil
}

var appFlags = []cli.Flag{
	flags.BeaconRPCProviderFlag,
	flags.RPCPort,
	//flags.SSVPubKeysFlag,
}

func main() {
	app := cli.App{}
	app.Name = "validator"
	app.Usage = `launches an ETH2.0 Secret-Shared-Validator (SSV) node`
	app.Version = version.GetVersion()
	app.Action = NewAndStartNode
	app.Flags = appFlags

	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
