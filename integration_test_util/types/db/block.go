package db

import "database/sql"

type BlockRecord struct {
	Height                   int64
	Hash                     string
	NumTxs                   int16
	NumEtx                   int16
	NumEtxCreateContract     int16
	ProposerAddress          sql.NullString
	Epoch                    int64
	ParentHash               string
	StateRoot                string
	Size                     int
	TotalValueH              int64
	TotalValueL              int64
	GasUsed                  int64
	GasLimit                 int64
	BaseFee                  int64
	CoinChanges              sql.NullString
	Erc20InvolvedAddresses   []string
	Erc721InvolvedAddresses  []string
	Erc1155InvolvedAddresses []string
	JustInvolvedAddresses    []string
	AnyBalanceError          sql.NullBool
	Extra                    sql.NullString
	EsMode                   sql.NullBool
	DataVersion              int16
	OutDated                 bool
	BurntFee                 int64
}
