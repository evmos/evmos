package db

import "database/sql"

type MessageIbcRecord struct {
	Height              int64
	TransactionIndex    int16
	MessageIndex        int16
	NestedMessageIndex  int16
	TransactionHash     string
	SequenceNo          string
	Port                string
	Channel             string
	CounterPartyChainId sql.NullString
	CounterPartyPort    sql.NullString
	CounterPartyChannel sql.NullString
	Type                string
	Amount              sql.NullString
	Denom               sql.NullString
	UnsafeBaseDenom     sql.NullString
	PartitionId         int64
}
