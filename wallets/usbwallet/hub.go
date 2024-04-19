// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package usbwallet

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	// runtime is listed as a potential source for non-determinism, but we use it only for checking the OS
	// #nosec
	"runtime"

	gethaccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/evmos/evmos/v17/wallets/accounts"
	usb "github.com/zondax/hid"
)

const (
	// LedgerScheme is the protocol scheme prefixing account and wallet URLs.
	LedgerScheme = "ledger"

	// onLinux is a boolean value to check if the operating system is Linux-based.
	onLinux = runtime.GOOS == "linux"

	// refreshThrottling is the minimum time between wallet refreshes to avoid USB
	// trashing.
	refreshThrottling = 500 * time.Millisecond
)

var _ accounts.Backend = &Hub{}

// Hub is an accounts.Backend that can find and handle generic USB hardware wallets.
type Hub struct {
	scheme     string        // Protocol scheme prefixing account and wallet URLs.
	vendorID   uint16        // USB vendor identifier used for device discovery
	productIDs []uint16      // USB product identifiers used for device discovery
	usageID    uint16        // USB usage page identifier used for macOS device discovery
	endpointID int           // USB endpoint identifier used for non-macOS device discovery
	makeDriver func() driver // Factory method to construct a vendor specific driver

	refreshed time.Time         // Time instance when the list of wallets was last refreshed
	wallets   []accounts.Wallet // List of USB wallet devices currently tracking

	quit chan chan error

	stateLock sync.RWMutex // Protects the internals of the hub from racey access

	// TODO(karalabe): remove if hotplug lands on Windows
	commsPend int        // Number of operations blocking enumeration
	commsLock sync.Mutex // Lock protecting the pending counter and enumeration
	enumFails uint32     // Number of times enumeration has failed
}

// NewLedgerHub creates a new hardware wallet manager for Ledger devices.
func NewLedgerHub() (*Hub, error) {
	return newHub(LedgerScheme, 0x2c97, []uint16{
		// Device definitions taken from
		// https://github.com/LedgerHQ/ledger-live/blob/38012bc8899e0f07149ea9cfe7e64b2c146bc92b/libs/ledgerjs/packages/devices/src/index.ts

		// Original product IDs
		0x0000, /* Ledger Blue */
		0x0001, /* Ledger Nano S */
		0x0004, /* Ledger Nano X */
		0x0005, /* Ledger Nano S Plus */
		0x0006, /* Ledger Nano FTS */

		0x0015, /* HID + U2F + WebUSB Ledger Blue */
		0x1015, /* HID + U2F + WebUSB Ledger Nano S */
		0x4015, /* HID + U2F + WebUSB Ledger Nano X */
		0x5015, /* HID + U2F + WebUSB Ledger Nano S Plus */
		0x6015, /* HID + U2F + WebUSB Ledger Nano FTS */

		0x0011, /* HID + WebUSB Ledger Blue */
		0x1011, /* HID + WebUSB Ledger Nano S */
		0x4011, /* HID + WebUSB Ledger Nano X */
		0x5011, /* HID + WebUSB Ledger Nano S Plus */
		0x6011, /* HID + WebUSB Ledger Nano FTS */
	}, 0xffa0, 0, newLedgerDriver)
}

// newHub creates a new hardware wallet manager for generic USB devices.
func newHub(scheme string, vendorID uint16, productIDs []uint16, usageID uint16, endpointID int, makeDriver func() driver) (*Hub, error) {
	if !usb.Supported() {
		return nil, errors.New("unsupported platform")
	}
	hub := &Hub{
		scheme:     scheme,
		vendorID:   vendorID,
		productIDs: productIDs,
		usageID:    usageID,
		endpointID: endpointID,
		makeDriver: makeDriver,
		quit:       make(chan chan error),
	}
	hub.refreshWallets()
	return hub, nil
}

// Wallets implements accounts.Backend, returning all the currently tracked USB
// devices that appear to be hardware wallets.
func (hub *Hub) Wallets() []accounts.Wallet {
	// Make sure the list of wallets is up-to-date
	hub.refreshWallets()

	hub.stateLock.RLock()
	defer hub.stateLock.RUnlock()

	cpy := make([]accounts.Wallet, len(hub.wallets))
	copy(cpy, hub.wallets)
	return cpy
}

// refreshWallets scans the USB devices attached to the machine and updates the
// list of wallets based on the found devices.
func (hub *Hub) refreshWallets() {
	// Don't scan the USB like crazy it the user fetches wallets in a loop
	hub.stateLock.RLock()
	elapsed := time.Since(hub.refreshed)
	hub.stateLock.RUnlock()

	if elapsed < refreshThrottling {
		return
	}

	// If USB enumeration is continually failing, don't keep trying indefinitely
	if atomic.LoadUint32(&hub.enumFails) > 2 {
		return
	}

	// Retrieve the current list of USB wallet devices
	var devices []usb.DeviceInfo

	if onLinux {
		// hidapi on Linux opens the device during enumeration to retrieve some infos,
		// breaking the Ledger protocol if that is waiting for user confirmation. This
		// is a bug acknowledged at Ledger, but it won't be fixed on old devices, so we
		// need to prevent concurrent comms ourselves. The more elegant solution would
		// be to ditch enumeration in favor of hotplug events, but that don't work yet
		// on Windows so if we need to hack it anyway, this is more elegant for now.
		hub.commsLock.Lock()
		if hub.commsPend > 0 { // A confirmation is pending, don't refresh
			hub.commsLock.Unlock()
			return
		}
	}
	infos := usb.Enumerate(hub.vendorID, 0)
	if infos == nil {
		if onLinux {
			// See rationale before the enumeration why this is needed and only on Linux.
			hub.commsLock.Unlock()
		}
		return
	}
	atomic.StoreUint32(&hub.enumFails, 0)

	for _, info := range infos {
		for _, id := range hub.productIDs {
			// Windows and macOS use UsageID matching, Linux uses Interface matching
			if info.ProductID == id && (info.UsagePage == hub.usageID || info.Interface == hub.endpointID) {
				devices = append(devices, info)
				break
			}
		}
	}

	if onLinux {
		// See rationale before the enumeration why this is needed and only on Linux.
		hub.commsLock.Unlock()
	}

	// Transform the current list of wallets into the new one
	hub.stateLock.Lock()

	wallets := make([]accounts.Wallet, 0, len(devices))

	for _, device := range devices {
		url := gethaccounts.URL{
			Scheme: hub.scheme,
			Path:   device.Path,
		}

		// Drop wallets in front of the next device or those that failed for some reason
		for len(hub.wallets) > 0 {
			// Abort if we're past the current device and found an operational one
			_, err := hub.wallets[0].Status()
			if hub.wallets[0].URL().Cmp(url) >= 0 || err == nil {
				break
			}
			// Drop the stale and failed devices
			hub.wallets = hub.wallets[1:]
		}

		// If there are no more wallets or the device is before the next, wrap new wallet
		if len(hub.wallets) == 0 || hub.wallets[0].URL().Cmp(url) > 0 {
			wallet := &wallet{
				hub:    hub,
				driver: hub.makeDriver(),
				url:    &url,
				info:   device,
			}

			wallets = append(wallets, wallet)
			continue
		}
		// If the device is the same as the first wallet, keep it
		if hub.wallets[0].URL().Cmp(url) == 0 {
			wallets = append(wallets, hub.wallets[0])
			hub.wallets = hub.wallets[1:]
			continue
		}
	}

	hub.refreshed = time.Now().UTC()
	hub.wallets = wallets
	hub.stateLock.Unlock()
}
