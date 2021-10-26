package types

type ERC20Data struct {
	Name     string
	Symbol   string
	Decimals uint8
}

type ERC20StringResponse struct {
	Name string
}

type ERC20Uint8Response struct {
	Value uint8
}

func NewERC20Data(name string, symbol string, decimals uint8) ERC20Data {
	return ERC20Data{
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}
}

func NewERC20StringResponse() ERC20StringResponse {
	return ERC20StringResponse{}
}

func NewERC20Uint8Response() ERC20Uint8Response {
	return ERC20Uint8Response{}
}
