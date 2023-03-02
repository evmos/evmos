// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

// These accounts represent the affected accounts during the Claims decay bug

// The decay occurred before planned and the corresponding claimed amounts
// were less than supposed to be

package v12

// Accounts holds the missing claim amount to the corresponding account
var Accounts = [2][2]string{ // TODO this is dummy data, need to update with real values
	{"evmos1009egsf8sk3puq3aynt8eymmcqnneezkkvceav", "1000000000000000000"},
	{"evmos100wr0u56zfmzyfzjqv0zp0wmczy6am7yel2na9", "1000000000000000000"},
}
