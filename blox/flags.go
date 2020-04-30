package blox

import (
	"gopkg.in/urfave/cli.v2"
)

var (
	// BeaconRPCProvider defines a beacon node RPC endpoint.
	BeaconRPCProvider = &cli.StringFlag{
		Name:  "beacon-rpc-provider",
		Usage: "Beacon node RPC provider endpoint",
		Value: "localhost:4000",
	}
	// KeyManager specifies the key manager to use.
	KeyManager = &cli.StringFlag{
		Name:  "keymanager",
		Usage: "The keymanger to use (unencrypted, interop, keystore, wallet)",
		Value: "",
	}
	// KeyManagerOpts specifies the key manager options.
	KeyManagerOpts = &cli.StringFlag{
		Name:  "keymanageropts",
		Usage: "The options for the keymanger, either a JSON string or path to same",
		Value: "",
	}
	// RemoteWalletAccount
	RemoteWalletAccount = &cli.StringFlag{
		Name:  "remote-wallet-account",
		Usage: "Remote wallet account presented on walletd",
		Value: "BloxValidator/Account.*",
	}
	// RemoteWalletLocation
	RemoteWalletLocation = &cli.StringFlag{
		Name:  "remote-wallet-location",
		Usage: "Location of remote wallet (my_remote_wallet.blox.io:443)",
	}
	// RemoteWalletCert defines a flag for the remote wallet's TLS certificate.
	RemoteWalletCert = &cli.StringFlag{
		Name:  "remote-wallet-ca-cert",
		Usage: "Certificate for secure remote wallet communication. Pass this and the remote-wallet-client-cert and remote-wallet-client-key flag in order to use securely communication.",
	}
	// RemoteWalletCert defines a flag for cert of client that communicate with remote wallet.
	RemoteWalletClientCert = &cli.StringFlag{
		Name:  "remote-wallet-client-cert",
		Usage: "Client certificate for secure remote wallet communication.",
	}
	// RemoteWalletClientKey defines a flag for key of client that communicate with remote wallet.
	RemoteWalletClientKey = &cli.StringFlag{
		Name:  "remote-wallet-client-key",
		Usage: "Client key for secure remote wallet communication.",
	}
)
