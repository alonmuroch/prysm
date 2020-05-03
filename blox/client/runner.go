package client

import (
	"os"

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
