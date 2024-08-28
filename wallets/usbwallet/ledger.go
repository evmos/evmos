// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package usbwallet

import (
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	gethaccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ledgerOpcode is an enumeration encoding the supported Ledger opcodes.
type ledgerOpcode byte

// ledgerParam1 is an enumeration encoding the supported Ledger parameters for
// specific opcodes. The same parameter values may be reused between opcodes.
type ledgerParam1 byte

// ledgerParam2 is an enumeration encoding the supported Ledger parameters for
// specific opcodes. The same parameter values may be reused between opcodes.
type ledgerParam2 byte

const (
	ledgerOpRetrieveAddress  ledgerOpcode = 0x02 // Returns the public key and Ethereum address for a given BIP 32 path
	ledgerOpGetConfiguration ledgerOpcode = 0x06 // Returns specific wallet application configuration
	ledgerOpSignTypedMessage ledgerOpcode = 0x0c // Signs an Ethereum message following the EIP 712 specification

	ledgerP1DirectlyFetchAddress    ledgerParam1 = 0x00 // Return address directly from the wallet
	ledgerP1InitTypedMessageData    ledgerParam1 = 0x00 // First chunk of Typed Message data
	ledgerP2DiscardAddressChainCode ledgerParam2 = 0x00 // Do not return the chain code along with the address
)

// errLedgerReplyInvalidHeader is the error message returned by a Ledger data exchange
// if the device replies with a mismatching header. This usually means the device
// is in browser mode.
var errLedgerReplyInvalidHeader = errors.New("ledger: invalid reply header")

// errLedgerInvalidVersionReply is the error message returned by a Ledger version retrieval
// when a response does arrive, but it does not contain the expected data.
var errLedgerInvalidVersionReply = errors.New("ledger: invalid version reply")

// ledgerDriver implements the communication with a Ledger hardware wallet.
type ledgerDriver struct {
	device  io.ReadWriter // USB device connection to communicate through
	version [3]byte       // Current version of the Ledger firmware (zero if app is offline)
	browser bool          // Flag whether the Ledger is in browser mode (reply channel mismatch)
	failure error         // Any failure that would make the device unusable
}

// newLedgerDriver creates a new instance of a Ledger USB protocol driver.
func newLedgerDriver() driver {
	return &ledgerDriver{}
}

// Status implements usbwallet.driver, returning various states the Ledger can
// currently be in.
func (w *ledgerDriver) Status() (string, error) {
	if w.failure != nil {
		return fmt.Sprintf("Failed: %v", w.failure), w.failure
	}
	if w.browser {
		return "Ethereum app in browser mode", w.failure
	}
	if w.offline() {
		return "Ethereum app offline", w.failure
	}
	return fmt.Sprintf("Ethereum app v%d.%d.%d online", w.version[0], w.version[1], w.version[2]), w.failure
}

// offline returns whether the wallet and the Ethereum app is offline or not.
//
// The method assumes that the state lock is held!
func (w *ledgerDriver) offline() bool {
	return w.version == [3]byte{0, 0, 0}
}

// Open implements usbwallet.driver, attempting to initialize the connection to the
// Ledger hardware wallet. The Ledger does not require a user passphrase, so that
// parameter is silently discarded.
func (w *ledgerDriver) Open(device io.ReadWriter, _ string) error {
	w.device, w.failure = device, nil

	_, _, err := w.ledgerDerive(gethaccounts.DefaultBaseDerivationPath)
	if err != nil {
		// Ethereum app is not running or in browser mode, nothing more to do, return
		if err == errLedgerReplyInvalidHeader {
			w.browser = true
		}
		return nil
	}
	// Try to resolve the Ethereum app's version, will fail prior to v1.0.2
	version, err := w.ledgerVersion()
	if err == nil {
		w.version = version
	} else {
		w.version = [3]byte{1, 0, 0} // Assume worst case, can't verify if v1.0.0 or v1.0.1
	}
	return nil
}

// Close implements usbwallet.driver, cleaning up and metadata maintained within
// the Ledger driver.
func (w *ledgerDriver) Close() error {
	w.browser, w.version = false, [3]byte{}
	return nil
}

// Heartbeat implements usbwallet.driver, performing a sanity check against the
// Ledger to see if it's still online.
func (w *ledgerDriver) Heartbeat() error {
	if _, err := w.ledgerVersion(); err != nil && err != errLedgerInvalidVersionReply {
		w.failure = err
		return err
	}
	return nil
}

// Derive implements usbwallet.driver, sending a derivation request to the Ledger
// and returning the Ethereum address located on that derivation path.
func (w *ledgerDriver) Derive(path gethaccounts.DerivationPath) (common.Address, *ecdsa.PublicKey, error) {
	return w.ledgerDerive(path)
}

// SignTypedMessage implements usbwallet.driver, sending the message to the Ledger and
// waiting for the user to sign or deny the transaction.
//
// Note: this was introduced in the ledger 1.5.0 firmware
func (w *ledgerDriver) SignTypedMessage(path gethaccounts.DerivationPath, domainHash, messageHash []byte) ([]byte, error) {
	// If the Ethereum app doesn't run, abort
	if w.offline() {
		return nil, gethaccounts.ErrWalletClosed
	}
	// Ensure the wallet is capable of signing the given transaction
	if w.version[0] < 1 && w.version[1] < 5 {
		//nolint:stylecheck // ST1005 requires error strings to be lowercase but Ledger as a brand name should start with a capital letter
		return nil, fmt.Errorf("Ledger version >= 1.5.0 required for EIP-712 signing (found version v%d.%d.%d)", w.version[0], w.version[1], w.version[2])
	}
	// All infos gathered and metadata checks out, request signing
	return w.ledgerSignTypedMessage(path, domainHash, messageHash)
}

// ledgerVersion retrieves the current version of the Ethereum wallet app running
// on the Ledger wallet.
//
// The version retrieval protocol is defined as follows:
//
//	CLA | INS | P1 | P2 | Lc | Le
//	----+-----+----+----+----+---
//	 E0 | 06  | 00 | 00 | 00 | 04
//
// With no input data, and the output data being:
//
//	Description                                        | Length
//	---------------------------------------------------+--------
//	Flags 01: arbitrary data signature enabled by user | 1 byte
//	Application major version                          | 1 byte
//	Application minor version                          | 1 byte
//	Application patch version                          | 1 byte
func (w *ledgerDriver) ledgerVersion() ([3]byte, error) {
	// Send the request and wait for the response
	reply, err := w.ledgerExchange(ledgerOpGetConfiguration, 0, 0, nil)
	if err != nil {
		return [3]byte{}, err
	}
	if len(reply) != 4 {
		return [3]byte{}, errLedgerInvalidVersionReply
	}
	// Cache the version for future reference
	var version [3]byte
	copy(version[:], reply[1:])
	return version, nil
}

// ledgerDerive retrieves the currently active Ethereum address from a Ledger
// wallet at the specified derivation path.
//
// The address derivation protocol is defined as follows:
//
//	CLA | INS | P1 | P2 | Lc  | Le
//	----+-----+----+----+-----+---
//	 E0 | 02  | 00 return address
//	            01 display address and confirm before returning
//	               | 00: do not return the chain code
//	               | 01: return the chain code
//	                    | var | 00
//
// Where the input data is:
//
//	Description                                      | Length
//	-------------------------------------------------+--------
//	Number of BIP 32 derivations to perform (max 10) | 1 byte
//	First derivation index (big endian)              | 4 bytes
//	...                                              | 4 bytes
//	Last derivation index (big endian)               | 4 bytes
//
// And the output data is:
//
//	Description             | Length
//	------------------------+-------------------
//	Public Key length       | 1 byte
//	Uncompressed Public Key | arbitrary
//	Ethereum address length | 1 byte
//	Ethereum address        | 40 bytes hex ascii
//	Chain code if requested | 32 bytes
func (w *ledgerDriver) ledgerDerive(derivationPath gethaccounts.DerivationPath) (common.Address, *ecdsa.PublicKey, error) {
	// Flatten the derivation path into the Ledger request
	path := make([]byte, 1+4*len(derivationPath))
	path[0] = byte(len(derivationPath))
	for i, component := range derivationPath {
		binary.BigEndian.PutUint32(path[1+4*i:], component)
	}

	// Send the request and wait for the response
	reply, err := w.ledgerExchange(ledgerOpRetrieveAddress, ledgerP1DirectlyFetchAddress, ledgerP2DiscardAddressChainCode, path)
	if err != nil {
		return common.Address{}, nil, err
	}

	// Verify public key was returned
	// #nosec G701 -- gosec will raise a warning on this integer conversion for potential overflow
	if len(reply) < 1 || len(reply) < 1+int(reply[0]) {
		return common.Address{}, nil, errors.New("reply lacks public key entry")
	}

	// #nosec G701 -- gosec will raise a warning on this integer conversion for potential overflow
	replyFirstByteAsInt := int(reply[0])

	pubkeyBz := reply[1 : 1+replyFirstByteAsInt]

	publicKey, err := crypto.UnmarshalPubkey(pubkeyBz)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	// Discard pubkey after fetching
	reply = reply[1+replyFirstByteAsInt:]

	// Extract the Ethereum hex address string
	// #nosec G701 -- gosec will raise a warning on this integer conversion for potential overflow
	if len(reply) < 1 || len(reply) < 1+int(reply[0]) {
		return common.Address{}, nil, errors.New("reply lacks address entry")
	}

	// Reset first byte after discarding pubkey from response
	// #nosec G701 -- gosec will raise a warning on this integer conversion for potential overflow
	replyFirstByteAsInt = int(reply[0])

	hexStr := reply[1 : 1+replyFirstByteAsInt]

	// Decode the hex string into an Ethereum address and return
	var address common.Address
	if _, err = hex.Decode(address[:], hexStr); err != nil {
		return common.Address{}, nil, err
	}

	derivedAddr := crypto.PubkeyToAddress(*publicKey)
	if derivedAddr != address {
		return common.Address{}, nil, fmt.Errorf("address mismatch, expected %s, got %s", derivedAddr, address)
	}

	return address, publicKey, nil
}

// ledgerSignTypedMessage sends the transaction to the Ledger wallet, and waits for the user
// to confirm or deny the transaction.
//
// The signing protocol is defined as follows:
//
//	CLA | INS | P1 | P2                          | Lc  | Le
//	----+-----+----+-----------------------------+-----+---
//	 E0 | 0C  | 00 | implementation version : 00 | variable | variable
//
// Where the input is:
//
//	Description                                      | Length
//	-------------------------------------------------+----------
//	Number of BIP 32 derivations to perform (max 10) | 1 byte
//	First derivation index (big endian)              | 4 bytes
//	...                                              | 4 bytes
//	Last derivation index (big endian)               | 4 bytes
//	domain hash                                      | 32 bytes
//	message hash                                     | 32 bytes
//
// And the output data is:
//
//	Description | Length
//	------------+---------
//	signature V | 1 byte
//	signature R | 32 bytes
//	signature S | 32 bytes
func (w *ledgerDriver) ledgerSignTypedMessage(derivationPath gethaccounts.DerivationPath, domainHash, messageHash []byte) ([]byte, error) {
	// Flatten the derivation path into the Ledger request
	path := make([]byte, 1+4*len(derivationPath))
	path[0] = byte(len(derivationPath))
	for i, component := range derivationPath {
		binary.BigEndian.PutUint32(path[1+4*i:], component)
	}
	// Create the 712 message
	var payload []byte
	payload = append(payload, path...)
	payload = append(payload, domainHash...)
	payload = append(payload, messageHash...)

	// Send the request and wait for the response
	var (
		op    = ledgerP1InitTypedMessageData
		reply []byte
		err   error
	)

	// Send the message over, ensuring it's processed correctly
	reply, err = w.ledgerExchange(ledgerOpSignTypedMessage, op, 0, payload)
	if err != nil {
		return nil, err
	}

	// Extract the Ethereum signature and do a sanity validation
	if len(reply) != crypto.SignatureLength {
		return nil, errors.New("reply lacks signature")
	}

	var signature []byte
	signature = append(signature, reply[1:]...)
	signature = append(signature, reply[0])

	return signature, nil
}

// ledgerExchange performs a data exchange with the Ledger wallet, sending it a
// message and retrieving the response.
//
// The common transport header is defined as follows:
//
//	Description                           | Length
//	--------------------------------------+----------
//	Communication channel ID (big endian) | 2 bytes
//	Command tag                           | 1 byte
//	Packet sequence index (big endian)    | 2 bytes
//	Payload                               | arbitrary
//
// The Communication channel ID allows commands multiplexing over the same
// physical link. It is not used for the time being, and should be set to 0101
// to avoid compatibility issues with implementations ignoring a leading 00 byte.
//
// The Command tag describes the message content. Use TAG_APDU (0x05) for standard
// APDU payloads, or TAG_PING (0x02) for a simple link test.
//
// The Packet sequence index describes the current sequence for fragmented payloads.
// The first fragment index is 0x00.
//
// APDU Command payloads are encoded as follows:
//
//	Description              | Length
//	-----------------------------------
//	APDU length (big endian) | 2 bytes
//	APDU CLA                 | 1 byte
//	APDU INS                 | 1 byte
//	APDU P1                  | 1 byte
//	APDU P2                  | 1 byte
//	APDU length              | 1 byte
//	Optional APDU data       | arbitrary
func (w *ledgerDriver) ledgerExchange(opcode ledgerOpcode, p1 ledgerParam1, p2 ledgerParam2, data []byte) ([]byte, error) {
	// Construct the message payload, possibly split into multiple chunks
	apdu := make([]byte, 2, 7+len(data))

	//#nosec G701 -- gosec will raise a warning on this integer conversion for potential overflow
	binary.BigEndian.PutUint16(apdu, uint16(5+len(data))) //nolint:gosec
	apdu = append(apdu, []byte{0xe0, byte(opcode), byte(p1), byte(p2), byte(len(data))}...)
	apdu = append(apdu, data...)

	// Stream all the chunks to the device
	header := []byte{0x01, 0x01, 0x05, 0x00, 0x00} // Channel ID and command tag appended
	chunk := make([]byte, 64)
	space := len(chunk) - len(header)

	for i := 0; len(apdu) > 0; i++ {
		// Construct the new message to stream
		chunk = append(chunk[:0], header...)
		//#nosec G701 -- gosec will raise a warning on this integer conversion for potential overflow
		binary.BigEndian.PutUint16(chunk[3:], uint16(i)) //nolint:gosec

		if len(apdu) > space {
			chunk = append(chunk, apdu[:space]...)
			apdu = apdu[space:]
		} else {
			chunk = append(chunk, apdu...)
			apdu = nil
		}
		// Send over to the device
		if _, err := w.device.Write(chunk); err != nil {
			return nil, err
		}
	}
	// Stream the reply back from the wallet in 64 byte chunks
	var reply []byte
	chunk = chunk[:64] // Yeah, we surely have enough space
	for {
		// Read the next chunk from the Ledger wallet
		if _, err := io.ReadFull(w.device, chunk); err != nil {
			return nil, err
		}

		// Make sure the transport header matches
		if chunk[0] != 0x01 || chunk[1] != 0x01 || chunk[2] != 0x05 {
			return nil, errLedgerReplyInvalidHeader
		}
		// If it's the first chunk, retrieve the total message length
		var payload []byte

		if chunk[3] == 0x00 && chunk[4] == 0x00 {
			//#nosec G701 -- gosec will raise a warning on this integer conversion for potential overflow
			reply = make([]byte, 0, int(binary.BigEndian.Uint16(chunk[5:7])))
			payload = chunk[7:]
		} else {
			payload = chunk[5:]
		}
		// Append to the reply and stop when filled up
		if left := cap(reply) - len(reply); left > len(payload) {
			reply = append(reply, payload...)
		} else {
			reply = append(reply, payload[:left]...)
			break
		}
	}
	return reply[:len(reply)-2], nil
}
