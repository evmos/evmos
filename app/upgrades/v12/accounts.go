// Copyright Tharsis Labs Ltd.(Evmos)
//  SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

// These accounts represent the affected accounts during the Claims decay bug

// The decay occurred before planned and the corresponding claimed amounts
// were less than supposed to be

package v12

// Accounts holds the missing claim amount to the corresponding account
var Accounts = [2][2]string{ // TODO this is dummy data, need to update with real values
	{"evmos1009egsf8sk3puq3aynt8eymmcqnneezkkvceav", "1000000000000000000"},
	{"evmos100wr0u56zfmzyfzjqv0zp0wmczy6am7yel2na9", "1000000000000000000"},
}
