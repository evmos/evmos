// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package gov

const (
	// ErrDifferentOrigin is raised when the origin address is not the same as the voter address.
	ErrDifferentOrigin = "tx origin address %s does not match the voter address %s"
	// ErrInvalidVoter is raised when the voter address is not valid.
	ErrInvalidVoter = "invalid voter address: %s"
	// ErrInvalidProposalID invalid proposal id.
	ErrInvalidProposalID = "invalid proposal id %d "
	// ErrInvalidPageRequest invalid page request.
	ErrInvalidPageRequest = "invalid page request"
	// ErrInvalidOption invalid option.
	ErrInvalidOption = "invalid option %s "
	// ErrInvalidMetadata invalid metadata.
	ErrInvalidMetadata = "invalid metadata %s "
)
