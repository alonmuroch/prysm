package client

import (
	"context"
	"fmt"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	originClient "github.com/prysmaticlabs/prysm/validator/client"
	"github.com/prysmaticlabs/prysm/validator/keymanager"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/urfave/cli.v2"
	"os"
	"sync"

	"github.com/prysmaticlabs/prysm/shared/cmd"
	"github.com/prysmaticlabs/prysm/shared/debug"
	"github.com/prysmaticlabs/prysm/shared/featureconfig"
	"github.com/prysmaticlabs/prysm/validator/flags"
)

func init() {
	appFlags = cmd.WrapFlags(append(appFlags, featureconfig.ValidatorFlags...))
}

var appFlags = []cli.Flag{
	flags.KeyManagerLocation,
	flags.KeyManagerCACert,
	flags.KeyManagerClientCert,
	flags.KeyManagerClientKey,
	flags.KeyManagerAccountPath,
	flags.BeaconRPCProviderFlag,
	flags.CertFlag,
	flags.GraffitiFlag,
	flags.KeystorePathFlag,
	flags.PasswordFlag,
	flags.DisablePenaltyRewardLogFlag,
	flags.UnencryptedKeysFlag,
	flags.InteropStartIndex,
	flags.InteropNumValidators,
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
	cmd.ChainConfigFileFlag,
	cmd.GrpcMaxCallRecvMsgSizeFlag,
}

type IParams interface {
	BeaconRPCProvider() string
	KeyManager() string
	KeyManagerLocation() string
	KeyManagerCACert() string
	KeyManagerClientCert() string
	KeyManagerClientKey() string
	KeyManagerAccountPath() string
}

type Params struct {
	beaconRPCProvider     string
	keyManager            string
	keyManagerLocation    string
	keyManagerCACert      string
	keyManagerClientCert  string
	keyManagerClientKey   string
	keyManagerAccountPath string
}

func NewParams(
	beaconRPCProvider string,
	keyManager string,
	keyManagerLocation string,
	keyManagerCACert string,
	keyManagerClientCert string,
	keyManagerClientKey string,
	keyManagerAccountPath string,
) *Params {
	return &Params{
		beaconRPCProvider,
		keyManager,
		keyManagerLocation,
		keyManagerCACert,
		keyManagerClientCert,
		keyManagerClientKey,
		keyManagerAccountPath,
	}
}

func (p *Params) BeaconRPCProvider() string {
	return p.beaconRPCProvider
}
func (p *Params) KeyManager() string {
	return p.keyManager
}
func (p *Params) KeyManagerLocation() string {
	return p.keyManagerLocation
}
func (p *Params) KeyManagerCACert() string {
	return p.keyManagerCACert
}
func (p *Params) KeyManagerClientCert() string {
	return p.keyManagerClientCert
}
func (p *Params) KeyManagerClientKey() string {
	return p.keyManagerClientKey
}
func (p *Params) KeyManagerAccountPath() string {
	return p.keyManagerAccountPath
}

type validatorRole int8

const (
	roleUnknown = iota
	roleAttester
	roleProposer
	roleAggregator
)

func run(ctx context.Context, v originClient.Validator) error {
	//ticker initialization
	if err := v.WaitForChainStart(ctx); err != nil {
		log.Fatalf("Could not determine if beacon chain started: %v", err)
		return nil
	}
	//get next slot
	slot := <-v.NextSlot()

	//deadline to deal with slot
	deadline := v.SlotDeadline(slot)
	slotCtx, cancel := context.WithDeadline(ctx, deadline)

	// this is necessary to set genesis time and the slot timer
	if err := v.UpdateDuties(ctx, slot); err != nil {
		cancel()
		log.WithError(err).Error("Update duties failed")
		return nil
	}

	var wg sync.WaitGroup

	allRoles, err := v.RolesAt(ctx, slot)
	if err != nil {
		log.WithError(err).Error("Could not get validator roles")
		return nil
	}
	for id, roles := range allRoles {
		wg.Add(len(roles))
		for _, role := range roles {
			go func(role validatorRole, id [48]byte) {
				defer wg.Done()
				switch role {
				case roleAttester:
					log.WithField("pubKey", fmt.Sprintf("%#x", bytesutil.Trunc(id[:]))).WithField("slot", slot).Info("SubmitAttestation")
					v.SubmitAttestation(slotCtx, slot, id)
				case roleProposer:
					log.WithField("pubKey", fmt.Sprintf("%#x", bytesutil.Trunc(id[:]))).WithField("slot", slot).Info("ProposeBlock")
					v.ProposeBlock(slotCtx, slot, id)
				case roleAggregator:
					log.WithField("pubKey", fmt.Sprintf("%#x", bytesutil.Trunc(id[:]))).WithField("slot", slot).Info("SubmitAggregateAndProof")
					v.SubmitAggregateAndProof(slotCtx, slot, id)
				case roleUnknown:
					log.WithField("pubKey", fmt.Sprintf("%#x", bytesutil.Trunc(id[:]))).WithField("slot", slot).Info("No active roles, doing nothing")
				default:
					log.WithField("pubKey", fmt.Sprintf("%#x", bytesutil.Trunc(id[:]))).WithField("slot", slot).Warnf("Unhandled role %v", role)
				}
			}(validatorRole(role), id)
		}
	}
	// Wait for all processes to complete, then iteration complete.
	go func() { wg.Wait() }()

	return nil
}

var grpcConnection *grpc.ClientConn

func startValidator(ctx *cli.Context) error {
	l := ctx.String(flags.KeyManagerLocation.Name)
	caC := ctx.String(flags.KeyManagerCACert.Name)
	cC := ctx.String(flags.KeyManagerClientCert.Name)
	cK := ctx.String(flags.KeyManagerClientKey.Name)
	aP := ctx.String(flags.KeyManagerAccountPath.Name)
	km, err := keymanager.NewRemoteWalletd(l, caC, cC, cK, aP)
	if err != nil {
		return err
	}
	if grpcConnection == nil {
		conn, err := NewGRPCConnection(ctx)
		if err != nil {
			return err
		}
		grpcConnection = conn
	}
	validator := NewValidator(km, grpcConnection)
	return run(ctx, validator)
}

func Run(params IParams) error {
	flags.BeaconRPCProviderFlag.Value = params.BeaconRPCProvider()
	flags.KeyManager.Value = params.KeyManager()
	flags.KeyManagerLocation.Value = params.KeyManagerLocation()
	flags.KeyManagerCACert.Value = params.KeyManagerCACert()
	flags.KeyManagerClientCert.Value = params.KeyManagerClientCert()
	flags.KeyManagerClientKey.Value = params.KeyManagerClientKey()
	flags.KeyManagerAccountPath.Value = params.KeyManagerAccountPath()
	app := &cli.App{}
	app.Action = startValidator
	app.Flags = appFlags
	err := app.Run(os.Args)
	if err != nil {
		return err
	}
	return nil
}
