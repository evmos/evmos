// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package usbwallet

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	gethaccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/evmos/evmos/v20/wallets/accounts"
	usb "github.com/zondax/hid"
)

// Maximum time between wallet health checks to detect USB unplugs.
const heartbeatCycle = time.Second

// driver defines the vendor specific functionality hardware wallets instances
// must implement to allow using them with the wallet lifecycle management.
type driver interface {
	// Status returns a textual status to aid the user in the current state of the
	// wallet. It also returns an error indicating any failure the wallet might have
	// encountered.
	Status() (string, error)

	// Open initializes access to a wallet instance. The passphrase parameter may
	// or may not be used by the implementation of a particular wallet instance.
	Open(device io.ReadWriter, passphrase string) error

	// Close releases any resources held by an open wallet instance.
	Close() error

	// Heartbeat performs a sanity check against the hardware wallet to see if it
	// is still online and healthy.
	Heartbeat() error

	// Derive sends a derivation request to the USB device and returns the Ethereum
	// address located on that path.
	Derive(path gethaccounts.DerivationPath) (common.Address, *ecdsa.PublicKey, error)

	// SignTypedMessage sends the message to the Ledger and waits for the user to sign
	// or deny the transaction.
	SignTypedMessage(path gethaccounts.DerivationPath, messageHash []byte, domainHash []byte) ([]byte, error)
}

// wallet represents the common functionality shared by all USB hardware
// wallets to prevent reimplementing the same complex maintenance mechanisms
// for different vendors.
type wallet struct {
	hub    *Hub              // USB hub scanning
	driver driver            // Hardware implementation of the low level device operations
	url    *gethaccounts.URL // Textual URL uniquely identifying this wallet

	info   usb.DeviceInfo // Known USB device infos about the wallet
	device *usb.Device    // USB device advertising itself as a hardware wallet

	accounts []accounts.Account                             // List of derive accounts pinned on the hardware wallet
	paths    map[common.Address]gethaccounts.DerivationPath // Known derivation paths for signing operations

	healthQuit chan chan error

	// Locking a hardware wallet is a bit special. Since hardware devices are lower
	// performing, any communication with them might take a non-negligible amount of
	// time. Worse still, waiting for user confirmation can take arbitrarily long,
	// but exclusive communication must be upheld during. Locking the entire wallet
	// in the meantime however would stall any parts of the system that don't want
	// to communicate, just read some state (e.g. list the accounts).
	//
	// As such, a hardware wallet needs two locks to function correctly. A state
	// lock can be used to protect the wallet's software-side internal state, which
	// must not be held exclusively during hardware communication. A communication
	// lock can be used to achieve exclusive access to the device itself, this one
	// however should allow "skipping" waiting for operations that might want to
	// use the device, but can live without too (e.g. account self-derivation).
	//
	// Since we have two locks, it's important to know how to properly use them:
	//   - Communication requires the `device` to not change, so obtaining the
	//     commsLock should be done after having a stateLock.
	//   - Communication must not disable read access to the wallet state, so it
	//     must only ever hold a *read* lock to stateLock.
	commsLock chan struct{} // Mutex (buf=1) for the USB comms without keeping the state locked
	stateLock sync.RWMutex  // Protects read and write access to the wallet struct fields
}

// URL implements accounts.Wallet, returning the URL of the USB hardware device.
func (w *wallet) URL() gethaccounts.URL {
	return *w.url // Immutable, no need for a lock
}

// Status implements accounts.Wallet, returning a custom status message from the
// underlying vendor-specific hardware wallet implementation.
func (w *wallet) Status() (string, error) {
	w.stateLock.RLock() // No device communication, state lock is enough
	defer w.stateLock.RUnlock()

	status, failure := w.driver.Status()
	if w.device == nil {
		return "Closed", failure
	}
	return status, failure
}

// Open implements accounts.Wallet, attempting to open a USB connection to the
// hardware wallet.
func (w *wallet) Open(passphrase string) error {
	w.stateLock.Lock() // State lock is enough since there's no connection yet at this point
	defer w.stateLock.Unlock()

	// If the device was already opened once, refuse to try again
	if w.paths != nil {
		return gethaccounts.ErrWalletAlreadyOpen
	}
	// Make sure the actual device connection is done only once
	if w.device == nil {
		device, err := w.info.Open()
		if err != nil {
			return err
		}
		w.device = device
		w.commsLock = make(chan struct{}, 1)
		w.commsLock <- struct{}{} // Enable lock
	}
	// Delegate device initialization to the underlying driver
	if err := w.driver.Open(w.device, passphrase); err != nil {
		return err
	}
	// Connection successful, start life-cycle management
	w.paths = make(map[common.Address]gethaccounts.DerivationPath)

	w.healthQuit = make(chan chan error)

	go w.heartbeat()

	return nil
}

// heartbeat is a health check loop for the USB wallets to periodically verify
// whether they are still present or if they malfunctioned.
func (w *wallet) heartbeat() {
	// Execute heartbeat checks until termination or error
	var (
		errc chan error
		err  error
	)
	for errc == nil && err == nil {
		// Wait until termination is requested or the heartbeat cycle arrives
		select {
		case errc = <-w.healthQuit:
			// Termination requested
			continue
		case <-time.After(heartbeatCycle):
			// Heartbeat time
		}
		// Execute a tiny data exchange to see responsiveness
		w.stateLock.RLock()
		if w.device == nil {
			// Terminated while waiting for the lock
			w.stateLock.RUnlock()
			continue
		}
		<-w.commsLock // Don't lock state while resolving version
		err = w.driver.Heartbeat()
		w.commsLock <- struct{}{}
		w.stateLock.RUnlock()

		if err != nil {
			w.stateLock.Lock() // Lock state to tear the wallet down
			//#nosec G703 -- ignoring the returned error on purpose here
			_ = w.close()
			w.stateLock.Unlock()
		}
		// Ignore non hardware related errors
		err = nil
	}
	// In case of error, wait for termination
	if err != nil {
		errc = <-w.healthQuit
	}
	errc <- err
}

// Close implements accounts.Wallet, closing the USB connection to the device.
func (w *wallet) Close() error {
	// Ensure the wallet was opened
	w.stateLock.RLock()
	hQuit := w.healthQuit
	w.stateLock.RUnlock()

	// Terminate the health checks
	var herr error
	if hQuit != nil {
		errc := make(chan error)
		hQuit <- errc
		herr = <-errc // Save for later, we *must* close the USB
	}

	// Terminate the device connection
	w.stateLock.Lock()
	defer w.stateLock.Unlock()

	w.healthQuit = nil

	if err := w.close(); err != nil {
		return err
	}
	if herr != nil {
		return herr
	}
	return nil
}

// close is the internal wallet closer that terminates the USB connection and
// resets all the fields to their defaults.
//
// Note, close assumes the state lock is held!
func (w *wallet) close() error {
	// Allow duplicate closes, especially for health-check failures
	if w.device == nil {
		return nil
	}
	// Close the device, clear everything, then return
	//#nosec G703 -- ignoring the returned error on purpose here
	_ = w.device.Close()
	w.device = nil

	w.accounts, w.paths = nil, nil
	return w.driver.Close()
}

// Accounts implements accounts.Wallet, returning the list of accounts pinned to
// the USB hardware wallet. If self-derivation was enabled, the account list is
// periodically expanded based on current chain state.
func (w *wallet) Accounts() []accounts.Account {
	// Return current account list
	w.stateLock.RLock()
	defer w.stateLock.RUnlock()

	cpy := make([]accounts.Account, len(w.accounts))
	copy(cpy, w.accounts)
	return cpy
}

// Contains implements accounts.Wallet, returning whether a particular account is
// or is not pinned into this wallet instance. Although we could attempt to resolve
// unpinned accounts, that would be a non-negligible hardware operation.
func (w *wallet) Contains(account accounts.Account) bool {
	w.stateLock.RLock()
	defer w.stateLock.RUnlock()

	_, exists := w.paths[account.Address]
	return exists
}

// Derive implements accounts.Wallet, deriving a new account at the specific
// derivation path. If pin is set to true, the account will be added to the list
// of tracked accounts.
func (w *wallet) Derive(path gethaccounts.DerivationPath, pin bool) (accounts.Account, error) {
	formatPathIfNeeded(path)

	// Try to derive the actual account and update its URL if successful
	w.stateLock.RLock() // Avoid device disappearing during derivation

	if w.device == nil {
		w.stateLock.RUnlock()
		return accounts.Account{}, gethaccounts.ErrWalletClosed
	}
	<-w.commsLock // Avoid concurrent hardware access
	address, publicKey, err := w.driver.Derive(path)
	w.commsLock <- struct{}{}

	w.stateLock.RUnlock()

	// If an error occurred or no pinning was requested, return
	if err != nil {
		return accounts.Account{}, err
	}

	account := accounts.Account{
		Address:   address,
		PublicKey: publicKey,
	}
	if !pin {
		return account, nil
	}
	// Pinning needs to modify the state
	w.stateLock.Lock()
	defer w.stateLock.Unlock()

	if _, ok := w.paths[address]; !ok {
		w.accounts = append(w.accounts, account)
		w.paths[address] = make(gethaccounts.DerivationPath, len(path))
		copy(w.paths[address], path)
	}
	return account, nil
}

// Format the hd path to harden the first three values (purpose, coinType, account)
// if needed, modifying the array in-place.
func formatPathIfNeeded(path gethaccounts.DerivationPath) {
	for i := 0; i < 3; i++ {
		if path[i] < 0x80000000 {
			path[i] += 0x80000000
		}
	}
}

// signHash implements accounts.Wallet, however signing arbitrary data is not
// supported for hardware wallets, so this method will always return an error.
func (w *wallet) signHash(_ accounts.Account, _ []byte) ([]byte, error) {
	return nil, gethaccounts.ErrNotSupported
}

// SignData signs keccak256(data). The mimetype parameter describes the type of data being signed
func (w *wallet) signData(account accounts.Account, mimeType string, data []byte) ([]byte, error) {
	// Unless we are doing 712 signing, simply dispatch to signHash
	if !(mimeType == gethaccounts.MimetypeTypedData && len(data) == 66 && data[0] == 0x19 && data[1] == 0x01) {
		return w.signHash(account, crypto.Keccak256(data))
	}

	// dispatch to 712 signing if the mimetype is TypedData and the format matches
	w.stateLock.RLock() // Comms have own mutex, this is for the state fields
	defer w.stateLock.RUnlock()

	// If the wallet is closed, abort
	if w.device == nil {
		return nil, gethaccounts.ErrWalletClosed
	}
	// Make sure the requested account is contained within
	path, ok := w.paths[account.Address]
	if !ok {
		return nil, gethaccounts.ErrUnknownAccount
	}
	// All infos gathered and metadata checks out, request signing
	<-w.commsLock
	defer func() { w.commsLock <- struct{}{} }()

	// Ensure the device isn't screwed with while user confirmation is pending
	// TODO(karalabe): remove if hotplug lands on Windows
	w.hub.commsLock.Lock()
	w.hub.commsPend++
	w.hub.commsLock.Unlock()

	defer func() {
		w.hub.commsLock.Lock()
		w.hub.commsPend--
		w.hub.commsLock.Unlock()
	}()
	// Sign the transaction
	signature, err := w.driver.SignTypedMessage(path, data[2:34], data[34:66])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (w *wallet) verifyTypedDataSignature(account accounts.Account, rawData []byte, signature []byte) error {
	if len(signature) != crypto.SignatureLength {
		return fmt.Errorf("invalid signature length: %d", len(signature))
	}

	// Copy signature as it would otherwise be modified
	sigCopy := make([]byte, len(signature))
	copy(sigCopy, signature)

	// Subtract 27 to match ECDSA standard
	sigCopy[crypto.RecoveryIDOffset] -= 27

	hash := crypto.Keccak256(rawData)

	derivedPubkey, err := crypto.Ecrecover(hash, sigCopy)
	if err != nil {
		return err
	}

	accountPK := crypto.FromECDSAPub(account.PublicKey)

	if !bytes.Equal(derivedPubkey, accountPK) {
		return errors.New("unauthorized: invalid signature verification")
	}

	return nil
}

// SignTypedData signs a TypedData in EIP-712 format. This method is a wrapper
// to call SignData after hashing and encoding the TypedData input
func (w *wallet) SignTypedData(account accounts.Account, typedData apitypes.TypedData) ([]byte, error) {
	_, rawData, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	rawDataBz := []byte(rawData)

	sigBytes, err := w.signData(account, "data/typed", rawDataBz)
	if err != nil {
		return nil, err
	}

	// Verify recovered public key matches expected value
	if err = w.verifyTypedDataSignature(account, rawDataBz, sigBytes); err != nil {
		return nil, err
	}

	return sigBytes, nil
}
