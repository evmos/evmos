package types

// ERC20Data represents the ERC20 token details used to map
// the token to a Cosmos Coin
type ERC20Data struct {
	Name     string
	Symbol   string
	Decimals uint8
}

// ERC20StringResponse defines the string value from the call response
type ERC20StringResponse struct {
	Value string
}

// ERC20StringResponse defines the string value from the call response
type ERC20Uint8Response struct {
	Value uint8
}

// NewERC20Data creates a new ERC20Data instance
func NewERC20Data(name, symbol string, decimals uint8) ERC20Data {
	return ERC20Data{
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}
}
