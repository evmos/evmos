// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package vm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

<<<<<<< HEAD
func TestExtendActivators(t *testing.T) {
	eips_snapshot := GetActivatorsEipNumbers()

	testCases := []struct {
		name           string
		new_activators map[int]func(*JumpTable)
		expPass        bool
		errContains    string
		postCheck      func()
=======
func TestValidEIPName(t *testing.T) {
	testCases := []struct {
		name        string
		eipName     string
		expPass     bool
		errContains string
	}{
		{
			"fail - invalid number",
			"os_OS",
			false,
			"eip number should be convertible to int",
		},
		{
			"fail - invalid structure, only chain name",
			"os",
			false,
			"eip name does not conform to structure 'chainName_Number'",
		},
		{
			"fail - invalid structure, only number",
			"0000",
			false,
			"eip name does not conform to structure 'chainName_Number'",
		},
		{
			"fail - invalid structure, only delimiter",
			"_",
			false,
			"eip number should be convertible to int",
		},
		{
			"success - valid eip name",
			"os_0000",
			true,
			"",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEIPName(tc.eipName)
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}

func TestExtendActivators(t *testing.T) {
	eips_snapshot := GetActivatorsEipNames()

	testCases := []struct {
		name          string
		newActivators map[string]func(*JumpTable)
		expPass       bool
		errContains   string
		postCheck     func()
>>>>>>> main
	}{
		{
			"success - nil new activators",
			nil,
			true,
			"",
			func() {
<<<<<<< HEAD
				eips := GetActivatorsEipNumbers()
=======
				eips := GetActivatorsEipNames()
>>>>>>> main
				require.ElementsMatch(t, eips_snapshot, eips, "expected eips number to be equal")
			},
		},
		{
			"success - single new activator",
<<<<<<< HEAD
			map[int]func(*JumpTable){
				0o000: func(jt *JumpTable) {},
=======
			map[string]func(*JumpTable){
				"evmos_0": func(jt *JumpTable) {},
>>>>>>> main
			},
			true,
			"",
			func() {
<<<<<<< HEAD
				eips := GetActivatorsEipNumbers()
				require.ElementsMatch(t, append(eips_snapshot, 0o000), eips, "expected eips number to be equal")
=======
				eips := GetActivatorsEipNames()
				require.ElementsMatch(t, append(eips_snapshot, "evmos_0"), eips, "expected eips number to be equal")
>>>>>>> main
			},
		},
		{
			"success - multiple new activators",
<<<<<<< HEAD
			map[int]func(*JumpTable){
				0o001: func(jt *JumpTable) {},
				0o002: func(jt *JumpTable) {},
=======
			map[string]func(*JumpTable){
				"evmos_1": func(jt *JumpTable) {},
				"evmos_2": func(jt *JumpTable) {},
>>>>>>> main
			},
			true,
			"",
			func() {
<<<<<<< HEAD
				eips := GetActivatorsEipNumbers()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, 0o000, 0o001, 0o002), eips, "expected eips number to be equal")
=======
				eips := GetActivatorsEipNames()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, "evmos_0", "evmos_1", "evmos_2"), eips, "expected eips number to be equal")
>>>>>>> main
			},
		},
		{
			"fail - repeated activator",
<<<<<<< HEAD
			map[int]func(*JumpTable){
				3855: func(jt *JumpTable) {},
=======
			map[string]func(*JumpTable){
				"ethereum_3855": func(jt *JumpTable) {},
>>>>>>> main
			},
			false,
			"",
			func() {
<<<<<<< HEAD
				eips := GetActivatorsEipNumbers()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, 0o000, 0o001, 0o002), eips, "expected eips number to be equal")
=======
				eips := GetActivatorsEipNames()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, "evmos_0", "evmos_1", "evmos_2"), eips, "expected eips number to be equal")
>>>>>>> main
			},
		},
		{
			"fail - valid activator is not stored if a repeated is present",
<<<<<<< HEAD
			map[int]func(*JumpTable){
				0o003: func(jt *JumpTable) {},
				3855:  func(jt *JumpTable) {},
=======
			map[string]func(*JumpTable){
				"evmos_3":       func(jt *JumpTable) {},
				"ethereum_3855": func(jt *JumpTable) {},
>>>>>>> main
			},
			false,
			"",
			func() {
<<<<<<< HEAD
				eips := GetActivatorsEipNumbers()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, 0o000, 0o001, 0o002), eips, "expected eips number to be equal")
=======
				eips := GetActivatorsEipNames()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, "evmos_0", "evmos_1", "evmos_2"), eips, "expected eips number to be equal")
>>>>>>> main
			},
		},
	}

	for _, tc := range testCases {
<<<<<<< HEAD
		err := ExtendActivators(tc.new_activators)
=======
		err := ExtendActivators(tc.newActivators)
>>>>>>> main
		if tc.expPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errContains, "expected different error")
		}

		tc.postCheck()
	}
}
<<<<<<< HEAD
=======

func TestAddOperation(t *testing.T) {
	// Functions used to create an operation.
	customExecute := func(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
		// no - op
		return nil, nil
	}
	customDynamicGas := func(evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
		// no-op
		return 0, nil
	}
	customMemorySize := func(stack *Stack) (uint64, bool) {
		// no-op
		return 0, false
	}

	const (
		EXISTENT OpCode = STOP
		NEW      OpCode = 0xf
	)

	testCases := []struct {
		name        string
		opName      string
		opNumber    OpCode
		expPass     bool
		errContains string
		postCheck   func()
	}{
		{
			"fail - operation with same number already exists",
			"TEST",
			EXISTENT,
			false,
			"already exists",
			func() {
				name := EXISTENT.String()
				require.Equal(t, "STOP", name)
			},
		},
		{
			"fail - operation with same name already exists",
			"CREATE",
			NEW,
			false,
			"already exists",
			func() {
				name := NEW.String()
				require.Contains(t, name, "not defined")
			},
		},
		{
			"fail - operation with same name of STOP",
			"STOP",
			NEW,
			false,
			"already exists",
			func() {
				name := NEW.String()
				require.Contains(t, name, "not defined")
			},
		},
		{
			"pass - new operation added to the list",
			"TEST",
			NEW,
			true,
			"",
			func() {
				name := NEW.String()
				require.Equal(t, "TEST", name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opInfo := OpCodeInfo{
				Number: tc.opNumber,
				Name:   tc.opName,
			}
			_, err := ExtendOperations(opInfo, customExecute, 0, customDynamicGas, 0, 0, customMemorySize)

			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}

			tc.postCheck()
		})
	}
}
>>>>>>> main
