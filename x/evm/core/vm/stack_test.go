// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm_test

import (
	"sync"
	"testing"
)

// Helper function to run table-driven tests
func runTest(t *testing.T, name string, newFunc func() interface{}, expectErr bool, errMsg string, testFunc interface{}, pool *sync.Pool) {
	t.Run(name, func(t *testing.T) {
		originalPool := *pool
		defer func() { *pool = originalPool }() // Cleanup to reset the pool after the test

		pool.New = newFunc

		switch tf := testFunc.(type) {
		case func() (*Stack, error):
			stack, err := tf()
			if expectErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if err.Error() != errMsg {
					t.Errorf("Expected error message %q, got %q", errMsg, err.Error())
				}
				if stack != nil {
					t.Errorf("Expected nil Stack, got %v", stack)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if stack == nil {
					t.Errorf("Expected a non-nil Stack, got nil")
				}
			}
		case func() (*ReturnStack, error):
			rStack, err := tf()
			if expectErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if err.Error() != errMsg {
					t.Errorf("Expected error message %q, got %q", errMsg, err.Error())
				}
				if rStack != nil {
					t.Errorf("Expected nil ReturnStack, got %v", rStack)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if rStack == nil {
					t.Errorf("Expected a non-nil ReturnStack, got nil")
				}
			}
		default:
			t.Fatalf("Unexpected test function signature")
		}
	})
}

func TestNewStack(t *testing.T) {
	tests := []struct {
		name      string
		newFunc   func() interface{}
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid Stack allocation",
			newFunc:   func() interface{} { return &Stack{} },
			expectErr: false,
		},
		{
			name:      "Invalid Stack allocation",
			newFunc:   func() interface{} { return "not a Stack" },
			expectErr: true,
			errMsg:    "type assertion failure: cannot get Stack pointer from stackPool",
		},
	}

	for _, tt := range tests {
		runTest(t, tt.name, tt.newFunc, tt.expectErr, tt.errMsg, NewStack, &stackPool)
	}
}

func TestNewReturnStack(t *testing.T) {
	tests := []struct {
		name      string
		newFunc   func() interface{}
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid ReturnStack allocation",
			newFunc:   func() interface{} { return &ReturnStack{} },
			expectErr: false,
		},
		{
			name:      "Invalid ReturnStack allocation",
			newFunc:   func() interface{} { return "not a ReturnStack" },
			expectErr: true,
			errMsg:    "type assertion failure: cannot get ReturnStack pointer from rStackPool",
		},
	}

	for _, tt := range tests {
		runTest(t, tt.name, tt.newFunc, tt.expectErr, tt.errMsg, NewReturnStack, &rStackPool)
	}
}
