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

package vm

import (
	"sync"
	"testing"
)

func TestNewStack(t *testing.T) {
	stackPool = sync.Pool{
		New: func() interface{} {
			return &Stack{}
		},
	}

	stack, err := NewStack()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if stack == nil {
		t.Errorf("Expected a non-nil Stack, got nil")
	}

	stackPool = sync.Pool{
		New: func() interface{} {
			return "not a Stack"
		},
	}

	stack, err = NewStack()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err.Error() != "type assertion failure: cannot get Stack pointer from stackPool" {
		t.Errorf("Expected 'type assertion failure' error, got %v", err)
	}
	if stack != nil {
		t.Errorf("Expected nil Stack, got %v", stack)
	}
}

func TestNewReturnStack(t *testing.T) {
	rStackPool = sync.Pool{
		New: func() interface{} {
			return &ReturnStack{}
		},
	}

	rStack, err := NewReturnStack()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rStack == nil {
		t.Errorf("Expected a non-nil ReturnStack, got nil")
	}

	rStackPool = sync.Pool{
		New: func() interface{} {
			return "not a ReturnStack"
		},
	}

	rStack, err = NewReturnStack()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err.Error() != "type assertion failure: cannot get ReturnStack pointer from rStackPool" {
		t.Errorf("Expected 'type assertion failure' error, got %v", err)
	}
	if rStack != nil {
		t.Errorf("Expected nil ReturnStack, got %v", rStack)
	}
}
