package db

import "database/sql"

type MessageRecord struct {
	TransactionHash                string
	Index                          int16
	Type                           string
	InvolvedAccountAddrs           []string
	PartitionId                    int64
	Height                         int64
	Content                        string
	TransactionIndex               int16
	NestedIndex                    int16
	NestedType                     sql.NullString
	NumNested                      int16
	InnerContent                   sql.NullString
	IbcDenom                       sql.NullString
	EvmTxHash                      sql.NullString
	EvmFrom                        sql.NullString
	EvmTo                          sql.NullString
	EvmNonce                       sql.NullInt64
	EvmTxType                      sql.NullInt16
	EvmCallType                    sql.NullString
	EvmInputData                   sql.NullString
	EvmLogBloom                    sql.NullString
	EvmValueH                      int64
	EvmValueL                      int64
	EvmGasPrice                    string
	EvmBurntFee                    int64
	EvmInternalTxBalanceAdjustment sql.NullString
	EvmSmartContracts              []string
	EvmInternalTxs                 sql.NullInt16
	EvmFailedReason                sql.NullString
	EvmMethod                      sql.NullString
	EvmRevenueAddr                 sql.NullString
	SortRef                        string
	TotalValueH                    int64
	TotalValueL                    int64
	Success                        bool
}
