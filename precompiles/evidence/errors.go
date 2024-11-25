// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evidence

const (
	// ErrInvalidEvidenceHash is raised when the evidence hash is invalid.
	ErrInvalidEvidenceHash = "invalid request; hash is empty"
	// ErrExpectedEquivocation is raised when the evidence is not an Equivocation.
	ErrExpectedEquivocation = "invalid evidence type: expected Equivocation"
)
