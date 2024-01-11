package db

import "database/sql"

type TransactionRecord struct {
	Hash                  string
	Height                int64
	OverallSuccessStatus  bool
	Messages              string
	Memo                  sql.NullString
	Signatures            []string
	SignerInfos           string
	Fee                   string
	GasWanted             int64
	GasUsed               int64
	RawLog                sql.NullString
	Logs                  sql.NullString
	PartitionId           int64
	Index                 int16
	Epoch                 int64
	TotalValueH           int64
	TotalValueL           int64
	TotalBurntFee         int64
	InvolvedAccountAddrs  []string
	FeePayer              string
	CountFailedMessages   int16
	OriginalSuccessStatus bool
	FailedReason          sql.NullString
	TypeSummary           sql.NullString
}
