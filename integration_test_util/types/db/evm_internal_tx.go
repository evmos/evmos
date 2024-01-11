package db

type EvmInternalTxRecord struct {
	TransactionHash    string
	EvmHash            string
	Height             int64
	TransactionIndex   int16
	MessageIndex       int16
	NestedMessageIndex int16
	Identifier         string
	Type               string
	From               string
	To                 string
	ValueH             int64
	ValueL             int64
	Gas                int32
	Order              int16
	PartitionId        int64
}
