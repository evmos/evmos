package db

import "database/sql"

type EvmLogRecord struct {
	TransactionHash    string
	EvmHash            string
	Height             int64
	TransactionIndex   int16
	MessageIndex       int16
	NestedMessageIndex int16
	LogIndex           int16
	Emitter            string
	Topic0             sql.NullString
	Topic1             sql.NullString
	Topic2             sql.NullString
	Topic3             sql.NullString
	Data               sql.NullString
	NftTokenIds        []string
	ValidErcTransfer   sql.NullInt16
	Removed            sql.NullBool
	PartitionId        int64
}
