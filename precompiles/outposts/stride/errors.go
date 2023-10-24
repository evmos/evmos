package stride

const (
	// ErrTokenPairNotFound is the error returned when a token pair is not found
	// #nosec G101
	ErrTokenPairNotFound = "token pair not found for %s"
	// ErrUnsupportedToken is the error returned when a token is not supported
	ErrUnsupportedToken = "unsupported token %s. The only supported token contract for Stride Outpost v1 is %s"
)
