package blox

import (
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v2"

	"github.com/prysmaticlabs/prysm/shared/cmd"
	"github.com/prysmaticlabs/prysm/shared/featureconfig"
	"github.com/prysmaticlabs/prysm/validator/node"
)

var log = logrus.WithField("prefix", "Run")

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
	BeaconRPCProvider,
	KeyManager,
	KeyManagerOpts,
	RemoteWalletAccount,
	RemoteWalletLocation,
	RemoteWalletCert,
	RemoteWalletClientCert,
	RemoteWalletClientKey,
}

type IParams interface {
	BeaconRPCProvider() string
	KeyManager() string
	KeyManagerOpts() string
	RemoteWalletAccount() string
	RemoteWalletLocation() string
	RemoteWalletCert() string
	RemoteWalletClientCert() string
	RemoteWalletClientKey() string
}

type Params struct {
	beaconRPCProvider      string
	keyManager             string
	keyManagerOpts         string
	remoteWalletAccount    string
	remoteWalletLocation   string
	remoteWalletCert       string
	remoteWalletClientCert string
	remoteWalletClientKey  string
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
func (p *Params) RemoteWalletAccount() string {
	return p.remoteWalletAccount
}
func (p *Params) RemoteWalletLocation() string {
	return p.remoteWalletLocation
}
func (p *Params) RemoteWalletCert() string {
	return p.remoteWalletCert
}
func (p *Params) RemoteWalletClientCert() string {
	return p.remoteWalletClientCert
}
func (p *Params) RemoteWalletClientKey() string {
	return p.remoteWalletClientKey
}

func initApp(params IParams) *cli.App {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  BeaconRPCProvider.Name,
				Value: params.BeaconRPCProvider(),
				Usage: BeaconRPCProvider.Usage,
			},
			&cli.StringFlag{
				Name:  KeyManager.Name,
				Value: params.KeyManager(),
				Usage: KeyManager.Usage,
			},
			&cli.StringFlag{
				Name:  KeyManagerOpts.Name,
				Value: params.KeyManagerOpts(),
				Usage: KeyManagerOpts.Usage,
			},
			&cli.StringFlag{
				Name:  RemoteWalletAccount.Name,
				Value: params.RemoteWalletAccount(),
				Usage: RemoteWalletAccount.Usage,
			},
			&cli.StringFlag{
				Name:  RemoteWalletLocation.Name,
				Value: params.RemoteWalletLocation(),
				Usage: RemoteWalletLocation.Usage,
			},
			&cli.StringFlag{
				Name:  RemoteWalletCert.Name,
				Value: params.RemoteWalletCert(),
				Usage: RemoteWalletCert.Usage,
			},
			&cli.StringFlag{
				Name:  RemoteWalletClientCert.Name,
				Value: params.RemoteWalletClientCert(),
				Usage: RemoteWalletClientCert.Usage,
			},
			&cli.StringFlag{
				Name:  RemoteWalletClientKey.Name,
				Value: params.RemoteWalletClientKey(),
				Usage: RemoteWalletClientKey.Usage,
			},
		},
	}
	return app
}

func Run(params IParams) error {
	app := initApp(params)
	app.Action = startNode
	app.Flags = appFlags
	err := app.Run(os.Args)
	if err != nil {
		return err
	}
	return nil
}
