package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
	"time"

	"gopkg.in/urfave/cli.v2"

	"github.com/prysmaticlabs/prysm/shared/cmd"
	"github.com/prysmaticlabs/prysm/shared/debug"
	"github.com/prysmaticlabs/prysm/shared/featureconfig"
	"github.com/prysmaticlabs/prysm/validator/flags"
	"github.com/prysmaticlabs/prysm/validator/node"
)

func init() {
	appFlags = cmd.WrapFlags(append(appFlags, featureconfig.ValidatorFlags...))
}

func startNode(ctx *cli.Context) error {
	validatorClient, err := node.NewValidatorClient(ctx)
	if err != nil {
		return err
	}
	validatorClient.Start()
	return nil
}

var appFlags = []cli.Flag{
	flags.BeaconRPCProviderFlag,
	flags.CertFlag,
	flags.GraffitiFlag,
	flags.KeystorePathFlag,
	flags.PasswordFlag,
	flags.DisablePenaltyRewardLogFlag,
	flags.UnencryptedKeysFlag,
	flags.InteropStartIndex,
	flags.InteropNumValidators,
	flags.GrpcMaxCallRecvMsgSizeFlag,
	flags.GrpcRetriesFlag,
	flags.GrpcHeadersFlag,
	flags.KeyManager,
	flags.KeyManagerOpts,
	flags.AccountMetricsFlag,
	cmd.VerbosityFlag,
	cmd.DataDirFlag,
	cmd.ClearDB,
	cmd.ForceClearDB,
	cmd.EnableTracingFlag,
	cmd.TracingProcessNameFlag,
	cmd.TracingEndpointFlag,
	cmd.TraceSampleFractionFlag,
	flags.MonitoringPortFlag,
	cmd.LogFormat,
	debug.PProfFlag,
	debug.PProfAddrFlag,
	debug.PProfPortFlag,
	debug.MemProfileRateFlag,
	debug.CPUProfileFlag,
	debug.TraceFlag,
	cmd.LogFileName,
	cmd.ConfigFileFlag,
}

type IParams interface {
	BeaconRPCProvider() string
	KeyManager() string
	KeyManagerOpts() string
}

type Params struct {
	beaconRPCProvider string
	keyManager        string
	keyManagerOpts    string
}

func New(
	beaconRPCProvider string,
	keyManager string,
	keyManagerOpts string,
) *Params {
	return &Params{
		beaconRPCProvider,
		keyManager,
		keyManagerOpts,
	}
}

func (p *Params) BeaconRPCProvider() string {
	return p.beaconRPCProvider
}
func (p *Params) KeyManager() string {
	return p.keyManager
}
func (p *Params) KeyManagerOpts() string {
	return p.keyManagerOpts
}

func Run(params IParams) error {
	flags.BeaconRPCProviderFlag.Value = params.BeaconRPCProvider()
	flags.KeyManager.Value = params.KeyManager()
	flags.KeyManagerOpts.Value = params.KeyManagerOpts()
	app := &cli.App{}
	app.Action = startNode
	app.Flags = appFlags
	err := app.Run(os.Args)
	if err != nil {
		return err
	}
	return nil
}

func Go(location, serverCA, certPEMBlock, keyPEMBlock string, accounts []string) {
	start := time.Now()

	err := setup(location, serverCA, certPEMBlock, keyPEMBlock)
	if err != nil {
		return
	}

	names := listAccounts(accounts)

	fmt.Printf("listAccounts %d ms: %v", int64(time.Since(start)/time.Millisecond), names)
}

var connection *grpc.ClientConn

func setup(location, serverCA, certPEMBlock, keyPEMBlock string) error {
	conn, err := connect(location, []byte(serverCA), []byte(certPEMBlock), []byte(keyPEMBlock))
	if err != nil {
		//load client crt failed
		return err
	}

	connection = conn
	return nil
}

const (
	// maxMessageSize is the largest message that can be received over GRPC.  Set to 8MB, which handles ~128K keys.
	maxMessageSize = 8 * 1024 * 1024
)

func connect(location string, serverCA, certPEMBlock, keyPEMBlock []byte) (*grpc.ClientConn, error) {
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(serverCA) {
		//failed to add server's CA certificate to pool
		return &grpc.ClientConn{}, errors.New("append cert from pem failed")
	}
	clientPair, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		//failed to add client's Key Pair certificate to pool
		return &grpc.ClientConn{}, err
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{clientPair},
		RootCAs:      cp,
	}
	clientCreds := credentials.NewTLS(tlsCfg)
	grpcOpts := []grpc.DialOption{
		// Require TLS with client certificate.
		grpc.WithTransportCredentials(clientCreds),
		// Receive large messages without erroring.
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)),
	}
	conn, err := grpc.Dial(location, grpcOpts...)
	if err != nil {
		//failed to connect
		return &grpc.ClientConn{}, err
	}
	return conn, nil
}

func listAccounts(accountsPatterns []string) []string {
	listerClient := pb.NewListerClient(connection)
	listAccountsReq := &pb.ListAccountsRequest{
		Paths: accountsPatterns,
	}
	resp, err := listerClient.ListAccounts(context.Background(), listAccountsReq)
	if err != nil {
		println(err)
	}

	var names []string
	for i := 0; i < len(resp.Accounts); i++ {
		names = append(names, resp.Accounts[i].Name)
	}
	return names
}
