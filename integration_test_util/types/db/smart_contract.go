package db

import "database/sql"

type SmartContractRecord struct {
	Height                       int64
	EvmTxHash                    string
	CreatorEvmAddr               string
	OriginalInput                string
	OriginalInputHash            string
	DeployedBytecode             string
	DeployedBytecodeHash         string
	SourceCode                   sql.NullString
	SourceCodeHash               sql.NullString
	AbiHash                      sql.NullString
	Name                         sql.NullString
	Symbol                       sql.NullString
	LogoUrl                      sql.NullString
	Erc                          sql.NullString
	FetchNameErr                 sql.NullString
	FetchSymbolErr               sql.NullString
	FetchDecimalsErr             sql.NullString
	ProxyCurImplAddr             sql.NullString
	ProxyPrevImplAddr            sql.NullString
	MinimalProxyImplAddr         sql.NullString
	VerifiedContractName         sql.NullString
	VerifiedContractFileName     sql.NullString
	VerifiedCompilerVersion      sql.NullString
	VerifiedEvmVersion           sql.NullString
	VerifiedConstructorArguments sql.NullString
	VerifiedLicenseType          sql.NullString
	VerifyId                     sql.NullString
	PossibleErc20                bool
	PossibleErc721               bool
	PossibleErc1155              bool
	Tracking                     bool
	Decimals                     int16
	DataVersion                  int16
	VerifiedEpoch                sql.NullInt64
	ProxyCurImplVer              sql.NullInt64
	VerifiedOptimizerRuns        sql.NullInt64
	ResyncAfterTracking          sql.NullBool
	VerifiedOptimizerEnabled     sql.NullBool
	IsInternalCreated            sql.NullBool
	ImplDetectMethod             sql.NullInt16
	VerifiedMatchType            sql.NullInt16
}
