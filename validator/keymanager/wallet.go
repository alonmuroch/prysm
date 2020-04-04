package keymanager

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/shared/bls"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/shibukawa/configdir"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type walletOpts struct {
	Location    string   `json:"location"`
	Accounts    []string `json:"accounts"`
	Passphrases []string `json:"passphrases"`
}

// Wallet is a key manager that loads keys from a local Ethereum 2 wallet.
type Wallet struct {
	mu          sync.RWMutex
	accounts    map[[48]byte]e2wtypes.Account
	walletScans map[string]*walletScan
	passphrases []string
}

type walletScan struct {
	wallet  e2wtypes.Wallet
	regexes []*regexp.Regexp
}

var walletOptsHelp = `The wallet key manager stores keys in a local encrypted store.  The options are:
  - location This is the location to look for wallets.  If not supplied it will
    use the standard (operating system-dependent) path.
  - accounts This is a list of account specifiers.  An account specifier is of
    the form <wallet name>/[account name],  where the account name can be a
    regular expression.  If the account specifier is just <wallet name> all
    accounts in that wallet will be used.  Multiple account specifiers can be
    supplied if required.
  - passphrase This is the passphrase used to encrypt the accounts when they
    were created.  Multiple passphrases can be supplied if required.

An sample keymanager options file (with annotations; these should be removed if
using this as a template) is:

  {
    "location":    "/wallets",               // Look for wallets in the directory '/wallets'
    "accounts":    ["Validators/Account.*"], // Use all accounts in the 'Validators' wallet starting with 'Account'
    "passphrases": ["secret1","secret2"]     // Use the passphrases 'secret1' and 'secret2' to decrypt accounts
  }`

// NewWallet creates a key manager populated with the keys from a wallet at the given path.
func NewWallet(input string) (KeyManager, string, error) {
	opts := &walletOpts{}
	err := json.Unmarshal([]byte(input), opts)
	if err != nil {
		return nil, walletOptsHelp, err
	}
	if err := validateWalletOpts(opts); err != nil {
		return nil, walletOptsHelp, err
	}
	var store e2wtypes.Store
	if opts.Location == "" {
		// Use default wallet location.
		configDirs := configdir.New("ethereum2", "wallets")
		opts.Location = configDirs.QueryFolders(configdir.Global)[0].Path
	}
	store = filesystem.New(filesystem.WithLocation(opts.Location))
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.WithError(err).Warn("Failed to set up filsesystem watcher; dynamic updates to validator keys not enabled")
		watcher = nil
	}

	km := &Wallet{
		mu:          sync.RWMutex{},
		accounts:    make(map[[48]byte]e2wtypes.Account),
		walletScans: make(map[string]*walletScan),
		passphrases: opts.Passphrases,
	}
	for _, path := range opts.Accounts {
		wallet, accountRe, err := pathToAccountSpecifier(store, path)
		if err != nil {
			return nil, walletOptsHelp, errors.Wrap(err, "failed to turn path in to account specifier")
		}
		km.addWalletScan(wallet, accountRe)
		for account := range wallet.Accounts() {
			if km.requestedAccount(wallet, account) {
				if km.unlockAccount(account) {
					km.addAccount(account)
				} else {
					log.WithField("path", path).Warn("Failed to unlock account with any supplied passphrase; cannot validate with this key")
				}
			}
		}
		if watcher != nil {
			if err := watcher.Add(filepath.Join(opts.Location, wallet.ID().String())); err != nil {
				log.WithError(err).Warn("Failed to add wallet directory; dynamic updates to validator keys disabled")
				watcher.Close()
				watcher = nil
			}
		}
	}

	// Start listening for changes to the relevant wallets.
	if watcher != nil {
		km.runListener(watcher, opts.Location)
	}

	return km, walletOptsHelp, nil
}

// validateWalletOpts validates the options for this keymanager; it returns an error if there is a problem with them.
func validateWalletOpts(opts *walletOpts) error {
	if len(opts.Accounts) == 0 {
		return errors.New("at least one account specifier is required")
	}

	if len(opts.Passphrases) == 0 {
		return errors.New("at least one passphrase is required to decrypt accounts")
	}

	if strings.Contains(opts.Location, "$") || strings.Contains(opts.Location, "~") || strings.Contains(opts.Location, "%") {
		log.WithField("path", opts.Location).Warn("Keystore path contains unexpanded shell expansion characters")
	}
	return nil
}

// pathToAccountSpecifier turns a path in to an account specifier.
func pathToAccountSpecifier(store e2wtypes.Store, path string) (e2wtypes.Wallet, *regexp.Regexp, error) {
	parts := strings.Split(path, "/")
	if len(parts) == 0 || len(parts[0]) == 0 {
		return nil, nil, fmt.Errorf("did not understand account specifier %q", path)
	}
	wallet, err := e2wallet.OpenWallet(parts[0], e2wallet.WithStore(store))
	if err != nil {
		return nil, nil, err
	}
	accountSpecifier := "^.*$"
	if len(parts) > 1 && len(parts[1]) > 0 {
		accountSpecifier = fmt.Sprintf("^%s$", parts[1])
	}
	re, err := regexp.Compile(accountSpecifier)
	if err != nil {
		return nil, nil, errors.Wrap(err, fmt.Sprintf("failed to compile account specifier %q", path))
	}
	return wallet, re, nil
}

// FetchValidatingKeys fetches the list of public keys that should be used to validate with.
func (km *Wallet) FetchValidatingKeys() ([][48]byte, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()
	res := make([][48]byte, 0, len(km.accounts))
	for pubKey := range km.accounts {
		res = append(res, pubKey)
	}
	return res, nil
}

// Sign signs a message for the validator to broadcast.
func (km *Wallet) Sign(pubKey [48]byte, root [32]byte) (*bls.Signature, error) {
	km.mu.RLock()
	account, exists := km.accounts[pubKey]
	km.mu.RUnlock()
	if !exists {
		return nil, ErrNoSuchKey
	}
	sig, err := account.Sign(root[:])
	if err != nil {
		return nil, err
	}
	return bls.SignatureFromBytes(sig.Marshal())
}

// addWalletScan builds a new or adds to an existing wallet scan criterion.
func (km *Wallet) addWalletScan(wallet e2wtypes.Wallet, re *regexp.Regexp) {
	scan, exists := km.walletScans[wallet.ID().String()]
	if !exists {
		scan = &walletScan{
			wallet:  wallet,
			regexes: make([]*regexp.Regexp, 0),
		}
		km.walletScans[wallet.ID().String()] = scan
	}
	scan.regexes = append(scan.regexes, re)
}

func (km *Wallet) runListener(watcher *fsnotify.Watcher, base string) {
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					// A file has been written to; add it (handleKeyAtPath() will handle situations where this isn't a key).
					path := strings.TrimPrefix(event.Name, base)
					km.handleKeyAtPath(path)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.WithError(err).Debug("watcher error")
			}
		}
	}()
}

// handleKeyAtPath handles a key that has shown up at a given path.
// Note that it's possible for this to be something other than a key, so we return silently if it doesn't match our expectations.
func (km *Wallet) handleKeyAtPath(path string) {
	walletIDStr, accountIDStr := filepath.Split(path)
	walletIDStr = strings.Trim(walletIDStr, string(filepath.Separator))
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		// Commonly an index, sometimes a backup file; ignore.
		return
	}
	walletScan, exists := km.walletScans[walletIDStr]
	if !exists {
		// Cannot find the wallet to which this refers; ignore.
		return
	}
	account, err := walletScan.wallet.AccountByID(accountID)
	if err != nil {
		// Unrecognised account; ignore.
		return
	}
	if km.requestedAccount(walletScan.wallet, account) {
		if km.unlockAccount(account) {
			km.addAccount(account)
		} else {
			// Valid key but we can't add it; warn.
			log.Warn("Failed to unlock account with any supplied passphrase; cannot validate with this key")
		}
	}
}

// requestedAccount returns true if this account has been requested by the configuration.
func (km *Wallet) requestedAccount(wallet e2wtypes.Wallet, account e2wtypes.Account) bool {
	walletScan, exists := km.walletScans[wallet.ID().String()]
	if !exists {
		// Not a wallet we are interested in.
		return false
	}
	for _, re := range walletScan.regexes {
		if re.Match([]byte(account.Name())) {
			return true
		}
	}
	return false
}

// addAccount adds an account to this keymanager's list.
func (km *Wallet) addAccount(account e2wtypes.Account) {
	pubKey := bytesutil.ToBytes48(account.PublicKey().Marshal())
	km.mu.Lock()
	km.accounts[pubKey] = account
	defer km.mu.Unlock()
}

// unlockAccount attempts to unlock an account with all known passphrases; returns true if successful.
func (km *Wallet) unlockAccount(account e2wtypes.Account) bool {
	for _, passphrase := range km.passphrases {
		if err := account.Unlock([]byte(passphrase)); err != nil {
			log.WithError(err).Trace("Failed to unlock account with one of the supplied passphrases")
			continue
		}
		return true
	}
	return false
}
