// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evidence

const (
	// ErrEvidenceNotFound is raised when the evidence is not found.
	ErrEvidenceNotFound = "evidence not found"
	// ErrInvalidEvidenceHash is raised when the evidence hash is invalid.
	ErrInvalidEvidenceHash = "invalid request; hash is empty"
	// ErrOriginDifferentFromSubmitter is raised when the origin address is different from the submitter address.
	ErrOriginDifferentFromSubmitter = "tx origin address %s does not match the submitter address %s"
	// ErrExpectedEquivocation is raised when the evidence is not an Equivocation.
	ErrExpectedEquivocation = "invalid evidence type: expected Equivocation"
)
