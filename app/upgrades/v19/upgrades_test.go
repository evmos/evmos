package v19_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/evmos/v18/app/upgrades/v19"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	evmostypes "github.com/evmos/evmos/v18/types"
	"github.com/evmos/evmos/v18/utils"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	"github.com/stretchr/testify/require"
)

const (
	// baseAccountAddress is the address of the base account that is set in genesis.
	baseAccountAddress = "evmos1k5c2zrqmmyx4rfkfp3l98qxvrpcr3m4kgxd0dp"
	// erc20ContractHex is the string representation of the ERC-20 token pair address.
	erc20ContractHex = "0x80b5a32E4F032B2a058b4F29EC95EEfEEB87aDcd"
	// SmartContractCode is the hex representation of the smart contract code that is set in genesis.
	//
	// NOTE: This was printed from the working branch and using ExportGenesis -> then printing the genesis account code.
	smartContractCode = "608060405234801561000f575f80fd5b50600436106101d8575f3560e01c80635c975abb11610102578063a217fddf116100a0578063d53913931161006f578063d53913931461057a578063d547741f14610598578063dd62ed3e146105b4578063e63ab1e9146105e4576101d8565b8063a217fddf146104cc578063a457c2d7146104ea578063a9059cbb1461051a578063ca15c8731461054a576101d8565b80638456cb59116100dc5780638456cb59146104445780639010d07c1461044e57806391d148541461047e57806395d89b41146104ae576101d8565b80635c975abb146103da57806370a08231146103f857806379cc679014610428576101d8565b8063282c51f31161017a578063395093511161014957806339509351146103685780633f4ba83a1461039857806340c10f19146103a257806342966c68146103be576101d8565b8063282c51f3146102f45780632f2ff15d14610312578063313ce5671461032e57806336568abe1461034c576101d8565b806318160ddd116101b657806318160ddd1461025a5780631cf2c7e21461027857806323b872dd14610294578063248a9ca3146102c4576101d8565b806301ffc9a7146101dc57806306fdde031461020c578063095ea7b31461022a575b5f80fd5b6101f660048036038101906101f19190612005565b610602565b604051610203919061204a565b60405180910390f35b61021461067b565b60405161022191906120ed565b60405180910390f35b610244600480360381019061023f919061219a565b61070b565b604051610251919061204a565b60405180910390f35b61026261072d565b60405161026f91906121e7565b60405180910390f35b610292600480360381019061028d919061219a565b610736565b005b6102ae60048036038101906102a99190612200565b6107b4565b6040516102bb919061204a565b60405180910390f35b6102de60048036038101906102d99190612283565b6107e2565b6040516102eb91906122bd565b60405180910390f35b6102fc6107fe565b60405161030991906122bd565b60405180910390f35b61032c600480360381019061032791906122d6565b610822565b005b610336610843565b604051610343919061232f565b60405180910390f35b610366600480360381019061036191906122d6565b610859565b005b610382600480360381019061037d919061219a565b6108dc565b60405161038f919061204a565b60405180910390f35b6103a0610912565b005b6103bc60048036038101906103b7919061219a565b61098c565b005b6103d860048036038101906103d39190612348565b610a0a565b005b6103e2610a1e565b6040516103ef919061204a565b60405180910390f35b610412600480360381019061040d9190612373565b610a33565b60405161041f91906121e7565b60405180910390f35b610442600480360381019061043d919061219a565b610a79565b005b61044c610a99565b005b6104686004803603810190610463919061239e565b610b13565b60405161047591906123eb565b60405180910390f35b610498600480360381019061049391906122d6565b610b3f565b6040516104a5919061204a565b60405180910390f35b6104b6610ba2565b6040516104c391906120ed565b60405180910390f35b6104d4610c32565b6040516104e191906122bd565b60405180910390f35b61050460048036038101906104ff919061219a565b610c38565b604051610511919061204a565b60405180910390f35b610534600480360381019061052f919061219a565b610cad565b604051610541919061204a565b60405180910390f35b610564600480360381019061055f9190612283565b610ccf565b60405161057191906121e7565b60405180910390f35b610582610cf0565b60405161058f91906122bd565b60405180910390f35b6105b260048036038101906105ad91906122d6565b610d14565b005b6105ce60048036038101906105c99190612404565b610d35565b6040516105db91906121e7565b60405180910390f35b6105ec610db7565b6040516105f991906122bd565b60405180910390f35b5f7f5a05180f000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19161480610674575061067382610ddb565b5b9050919050565b60606005805461068a9061246f565b80601f01602080910402602001604051908101604052809291908181526020018280546106b69061246f565b80156107015780601f106106d857610100808354040283529160200191610701565b820191905f5260205f20905b8154815290600101906020018083116106e457829003601f168201915b5050505050905090565b5f80610715610e54565b9050610722818585610e5b565b600191505092915050565b5f600454905090565b6107677f3c11d16cbaffd01df69ce1c404f6340ee057498f5f00246190ea54220576a848610762610e54565b610b3f565b6107a6576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161079d9061250f565b60405180910390fd5b6107b0828261101e565b5050565b5f806107be610e54565b90506107cb8582856111e3565b6107d685858561126e565b60019150509392505050565b5f805f8381526020019081526020015f20600101549050919050565b7f3c11d16cbaffd01df69ce1c404f6340ee057498f5f00246190ea54220576a84881565b61082b826107e2565b610834816114dd565b61083e83836114f1565b505050565b5f600760019054906101000a900460ff16905090565b610861610e54565b73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16146108ce576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108c59061259d565b60405180910390fd5b6108d88282611523565b5050565b5f806108e6610e54565b90506109078185856108f88589610d35565b61090291906125e8565b610e5b565b600191505092915050565b6109437f65d7a28e3265b37a6474929f336521b332c1681b933f6cb9f3376673440d862a61093e610e54565b610b3f565b610982576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109799061268b565b60405180910390fd5b61098a611555565b565b6109bd7f9f2df0fed2c77648de5860a4cc508cd0818c85b8b8a1ab4ceeef8d981c8956a66109b8610e54565b610b3f565b6109fc576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109f390612719565b60405180910390fd5b610a0682826115b6565b5050565b610a1b610a15610e54565b8261101e565b50565b5f60075f9054906101000a900460ff16905090565b5f60025f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b610a8b82610a85610e54565b836111e3565b610a95828261101e565b5050565b610aca7f65d7a28e3265b37a6474929f336521b332c1681b933f6cb9f3376673440d862a610ac5610e54565b610b3f565b610b09576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b00906127a7565b60405180910390fd5b610b11611705565b565b5f610b378260015f8681526020019081526020015f2061176790919063ffffffff16565b905092915050565b5f805f8481526020019081526020015f205f015f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff16905092915050565b606060068054610bb19061246f565b80601f0160208091040260200160405190810160405280929190818152602001828054610bdd9061246f565b8015610c285780601f10610bff57610100808354040283529160200191610c28565b820191905f5260205f20905b815481529060010190602001808311610c0b57829003601f168201915b5050505050905090565b5f801b81565b5f80610c42610e54565b90505f610c4f8286610d35565b905083811015610c94576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c8b90612835565b60405180910390fd5b610ca18286868403610e5b565b60019250505092915050565b5f80610cb7610e54565b9050610cc481858561126e565b600191505092915050565b5f610ce960015f8481526020019081526020015f2061177e565b9050919050565b7f9f2df0fed2c77648de5860a4cc508cd0818c85b8b8a1ab4ceeef8d981c8956a681565b610d1d826107e2565b610d26816114dd565b610d308383611523565b505050565b5f60035f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b7f65d7a28e3265b37a6474929f336521b332c1681b933f6cb9f3376673440d862a81565b5f7f7965db0b000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19161480610e4d5750610e4c82611791565b5b9050919050565b5f33905090565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610ec9576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610ec0906128c3565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610f37576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610f2e90612951565b60405180910390fd5b8060035f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9258360405161101191906121e7565b60405180910390a3505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361108c576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611083906129df565b60405180910390fd5b611097825f836117fa565b5f60025f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205490508181101561111b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161111290612a6d565b60405180910390fd5b81810360025f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508160045f82825403925050819055505f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040516111cb91906121e7565b60405180910390a36111de835f8461180a565b505050565b5f6111ee8484610d35565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8114611268578181101561125a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161125190612ad5565b60405180910390fd5b6112678484848403610e5b565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036112dc576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016112d390612b63565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361134a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161134190612bf1565b60405180910390fd5b6113558383836117fa565b5f60025f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050818110156113d9576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016113d090612c7f565b60405180910390fd5b81810360025f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508160025f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040516114c491906121e7565b60405180910390a36114d784848461180a565b50505050565b6114ee816114e9610e54565b61180f565b50565b6114fb8282611893565b61151e8160015f8581526020019081526020015f2061196d90919063ffffffff16565b505050565b61152d828261199a565b6115508160015f8581526020019081526020015f20611a7490919063ffffffff16565b505050565b61155d611aa1565b5f60075f6101000a81548160ff0219169083151502179055507f5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa61159f610e54565b6040516115ac91906123eb565b60405180910390a1565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603611624576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161161b90612ce7565b60405180910390fd5b61162f5f83836117fa565b8060045f82825461164091906125e8565b925050819055508060025f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055508173ffffffffffffffffffffffffffffffffffffffff165f73ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040516116ee91906121e7565b60405180910390a36117015f838361180a565b5050565b61170d611aea565b600160075f6101000a81548160ff0219169083151502179055507f62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258611750610e54565b60405161175d91906123eb565b60405180910390a1565b5f611774835f0183611b34565b5f1c905092915050565b5f61178a825f01611b5b565b9050919050565b5f7f01ffc9a7000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916149050919050565b611805838383611b6a565b505050565b505050565b6118198282610b3f565b61188f5761182681611bc2565b611833835f1c6020611bef565b604051602001611844929190612dd3565b6040516020818303038152906040526040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161188691906120ed565b60405180910390fd5b5050565b61189d8282610b3f565b6119695760015f808481526020019081526020015f205f015f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff02191690831515021790555061190e610e54565b73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16837f2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d60405160405180910390a45b5050565b5f611992835f018373ffffffffffffffffffffffffffffffffffffffff165f1b611e24565b905092915050565b6119a48282610b3f565b15611a70575f805f8481526020019081526020015f205f015f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff021916908315150217905550611a15610e54565b73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16837ff6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b60405160405180910390a45b5050565b5f611a99835f018373ffffffffffffffffffffffffffffffffffffffff165f1b611e8b565b905092915050565b611aa9610a1e565b611ae8576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611adf90612e56565b60405180910390fd5b565b611af2610a1e565b15611b32576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611b2990612ebe565b60405180910390fd5b565b5f825f018281548110611b4a57611b49612edc565b5b905f5260205f200154905092915050565b5f815f01805490509050919050565b611b75838383611f87565b611b7d610a1e565b15611bbd576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611bb490612f79565b60405180910390fd5b505050565b6060611be88273ffffffffffffffffffffffffffffffffffffffff16601460ff16611bef565b9050919050565b60605f6002836002611c019190612f97565b611c0b91906125e8565b67ffffffffffffffff811115611c2457611c23612fd8565b5b6040519080825280601f01601f191660200182016040528015611c565781602001600182028036833780820191505090505b5090507f3000000000000000000000000000000000000000000000000000000000000000815f81518110611c8d57611c8c612edc565b5b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191690815f1a9053507f780000000000000000000000000000000000000000000000000000000000000081600181518110611cf057611cef612edc565b5b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191690815f1a9053505f6001846002611d2e9190612f97565b611d3891906125e8565b90505b6001811115611dd7577f3031323334353637383961626364656600000000000000000000000000000000600f861660108110611d7a57611d79612edc565b5b1a60f81b828281518110611d9157611d90612edc565b5b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191690815f1a905350600485901c945080611dd090613005565b9050611d3b565b505f8414611e1a576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611e1190613076565b60405180910390fd5b8091505092915050565b5f611e2f8383611f8c565b611e8157825f0182908060018154018082558091505060019003905f5260205f20015f9091909190915055825f0180549050836001015f8481526020019081526020015f208190555060019050611e85565b5f90505b92915050565b5f80836001015f8481526020019081526020015f205490505f8114611f7c575f600182611eb89190613094565b90505f6001865f0180549050611ece9190613094565b9050818114611f34575f865f018281548110611eed57611eec612edc565b5b905f5260205f200154905080875f018481548110611f0e57611f0d612edc565b5b905f5260205f20018190555083876001015f8381526020019081526020015f2081905550505b855f01805480611f4757611f466130c7565b5b600190038181905f5260205f20015f90559055856001015f8681526020019081526020015f205f905560019350505050611f81565b5f9150505b92915050565b505050565b5f80836001015f8481526020019081526020015f20541415905092915050565b5f80fd5b5f7fffffffff0000000000000000000000000000000000000000000000000000000082169050919050565b611fe481611fb0565b8114611fee575f80fd5b50565b5f81359050611fff81611fdb565b92915050565b5f6020828403121561201a57612019611fac565b5b5f61202784828501611ff1565b91505092915050565b5f8115159050919050565b61204481612030565b82525050565b5f60208201905061205d5f83018461203b565b92915050565b5f81519050919050565b5f82825260208201905092915050565b5f5b8381101561209a57808201518184015260208101905061207f565b5f8484015250505050565b5f601f19601f8301169050919050565b5f6120bf82612063565b6120c9818561206d565b93506120d981856020860161207d565b6120e2816120a5565b840191505092915050565b5f6020820190508181035f83015261210581846120b5565b905092915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6121368261210d565b9050919050565b6121468161212c565b8114612150575f80fd5b50565b5f813590506121618161213d565b92915050565b5f819050919050565b61217981612167565b8114612183575f80fd5b50565b5f8135905061219481612170565b92915050565b5f80604083850312156121b0576121af611fac565b5b5f6121bd85828601612153565b92505060206121ce85828601612186565b9150509250929050565b6121e181612167565b82525050565b5f6020820190506121fa5f8301846121d8565b92915050565b5f805f6060848603121561221757612216611fac565b5b5f61222486828701612153565b935050602061223586828701612153565b925050604061224686828701612186565b9150509250925092565b5f819050919050565b61226281612250565b811461226c575f80fd5b50565b5f8135905061227d81612259565b92915050565b5f6020828403121561229857612297611fac565b5b5f6122a58482850161226f565b91505092915050565b6122b781612250565b82525050565b5f6020820190506122d05f8301846122ae565b92915050565b5f80604083850312156122ec576122eb611fac565b5b5f6122f98582860161226f565b925050602061230a85828601612153565b9150509250929050565b5f60ff82169050919050565b61232981612314565b82525050565b5f6020820190506123425f830184612320565b92915050565b5f6020828403121561235d5761235c611fac565b5b5f61236a84828501612186565b91505092915050565b5f6020828403121561238857612387611fac565b5b5f61239584828501612153565b91505092915050565b5f80604083850312156123b4576123b3611fac565b5b5f6123c18582860161226f565b92505060206123d285828601612186565b9150509250929050565b6123e58161212c565b82525050565b5f6020820190506123fe5f8301846123dc565b92915050565b5f806040838503121561241a57612419611fac565b5b5f61242785828601612153565b925050602061243885828601612153565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061248657607f821691505b60208210810361249957612498612442565b5b50919050565b7f45524332304d696e7465724275726e6572446563696d616c733a206d757374205f8201527f68617665206275726e657220726f6c6520746f206275726e0000000000000000602082015250565b5f6124f960388361206d565b91506125048261249f565b604082019050919050565b5f6020820190508181035f830152612526816124ed565b9050919050565b7f416363657373436f6e74726f6c3a2063616e206f6e6c792072656e6f756e63655f8201527f20726f6c657320666f722073656c660000000000000000000000000000000000602082015250565b5f612587602f8361206d565b91506125928261252d565b604082019050919050565b5f6020820190508181035f8301526125b48161257b565b9050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f6125f282612167565b91506125fd83612167565b9250828201905080821115612615576126146125bb565b5b92915050565b7f45524332304d696e7465724275726e6572446563696d616c733a206d757374205f8201527f686176652070617573657220726f6c6520746f20756e70617573650000000000602082015250565b5f612675603b8361206d565b91506126808261261b565b604082019050919050565b5f6020820190508181035f8301526126a281612669565b9050919050565b7f45524332304d696e7465724275726e6572446563696d616c733a206d757374205f8201527f68617665206d696e74657220726f6c6520746f206d696e740000000000000000602082015250565b5f61270360388361206d565b915061270e826126a9565b604082019050919050565b5f6020820190508181035f830152612730816126f7565b9050919050565b7f45524332304d696e7465724275726e6572446563696d616c733a206d757374205f8201527f686176652070617573657220726f6c6520746f20706175736500000000000000602082015250565b5f61279160398361206d565b915061279c82612737565b604082019050919050565b5f6020820190508181035f8301526127be81612785565b9050919050565b7f45524332303a2064656372656173656420616c6c6f77616e63652062656c6f775f8201527f207a65726f000000000000000000000000000000000000000000000000000000602082015250565b5f61281f60258361206d565b915061282a826127c5565b604082019050919050565b5f6020820190508181035f83015261284c81612813565b9050919050565b7f45524332303a20617070726f76652066726f6d20746865207a65726f206164645f8201527f7265737300000000000000000000000000000000000000000000000000000000602082015250565b5f6128ad60248361206d565b91506128b882612853565b604082019050919050565b5f6020820190508181035f8301526128da816128a1565b9050919050565b7f45524332303a20617070726f766520746f20746865207a65726f2061646472655f8201527f7373000000000000000000000000000000000000000000000000000000000000602082015250565b5f61293b60228361206d565b9150612946826128e1565b604082019050919050565b5f6020820190508181035f8301526129688161292f565b9050919050565b7f45524332303a206275726e2066726f6d20746865207a65726f206164647265735f8201527f7300000000000000000000000000000000000000000000000000000000000000602082015250565b5f6129c960218361206d565b91506129d48261296f565b604082019050919050565b5f6020820190508181035f8301526129f6816129bd565b9050919050565b7f45524332303a206275726e20616d6f756e7420657863656564732062616c616e5f8201527f6365000000000000000000000000000000000000000000000000000000000000602082015250565b5f612a5760228361206d565b9150612a62826129fd565b604082019050919050565b5f6020820190508181035f830152612a8481612a4b565b9050919050565b7f45524332303a20696e73756666696369656e7420616c6c6f77616e63650000005f82015250565b5f612abf601d8361206d565b9150612aca82612a8b565b602082019050919050565b5f6020820190508181035f830152612aec81612ab3565b9050919050565b7f45524332303a207472616e736665722066726f6d20746865207a65726f2061645f8201527f6472657373000000000000000000000000000000000000000000000000000000602082015250565b5f612b4d60258361206d565b9150612b5882612af3565b604082019050919050565b5f6020820190508181035f830152612b7a81612b41565b9050919050565b7f45524332303a207472616e7366657220746f20746865207a65726f20616464725f8201527f6573730000000000000000000000000000000000000000000000000000000000602082015250565b5f612bdb60238361206d565b9150612be682612b81565b604082019050919050565b5f6020820190508181035f830152612c0881612bcf565b9050919050565b7f45524332303a207472616e7366657220616d6f756e74206578636565647320625f8201527f616c616e63650000000000000000000000000000000000000000000000000000602082015250565b5f612c6960268361206d565b9150612c7482612c0f565b604082019050919050565b5f6020820190508181035f830152612c9681612c5d565b9050919050565b7f45524332303a206d696e7420746f20746865207a65726f2061646472657373005f82015250565b5f612cd1601f8361206d565b9150612cdc82612c9d565b602082019050919050565b5f6020820190508181035f830152612cfe81612cc5565b9050919050565b5f81905092915050565b7f416363657373436f6e74726f6c3a206163636f756e74200000000000000000005f82015250565b5f612d43601783612d05565b9150612d4e82612d0f565b601782019050919050565b5f612d6382612063565b612d6d8185612d05565b9350612d7d81856020860161207d565b80840191505092915050565b7f206973206d697373696e6720726f6c65200000000000000000000000000000005f82015250565b5f612dbd601183612d05565b9150612dc882612d89565b601182019050919050565b5f612ddd82612d37565b9150612de98285612d59565b9150612df482612db1565b9150612e008284612d59565b91508190509392505050565b7f5061757361626c653a206e6f74207061757365640000000000000000000000005f82015250565b5f612e4060148361206d565b9150612e4b82612e0c565b602082019050919050565b5f6020820190508181035f830152612e6d81612e34565b9050919050565b7f5061757361626c653a20706175736564000000000000000000000000000000005f82015250565b5f612ea860108361206d565b9150612eb382612e74565b602082019050919050565b5f6020820190508181035f830152612ed581612e9c565b9050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b7f45524332305061757361626c653a20746f6b656e207472616e736665722077685f8201527f696c652070617573656400000000000000000000000000000000000000000000602082015250565b5f612f63602a8361206d565b9150612f6e82612f09565b604082019050919050565b5f6020820190508181035f830152612f9081612f57565b9050919050565b5f612fa182612167565b9150612fac83612167565b9250828202612fba81612167565b91508282048414831517612fd157612fd06125bb565b5b5092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b5f61300f82612167565b91505f8203613021576130206125bb565b5b600182039050919050565b7f537472696e67733a20686578206c656e67746820696e73756666696369656e745f82015250565b5f61306060208361206d565b915061306b8261302c565b602082019050919050565b5f6020820190508181035f83015261308d81613054565b9050919050565b5f61309e82612167565b91506130a983612167565b92508282039050818111156130c1576130c06125bb565b5b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603160045260245ffdfea26469706673582212208e17db436f847258cc8b148f2ec3689c9d1043000f9acacdf5a893071a6136a964736f6c63430008170033"
)

var (
	// code is the bytecode of the contract.
	//
	// These conversions are taken from the code hash check in InitGenesis:
	// https://github.com/evmos/evmos/blob/ca0f3e4d3e8407fd983b42aa595b052a67c1598b/x/evm/genesis.go#L48-L64
	code = common.Hex2Bytes(smartContractCode)
	// codeHash is the hash of the contract code.
	codeHash = crypto.Keccak256Hash(code)

	// storage is the genesis storage for the deployed ERC-20 contract
	//
	// NOTE: This is generated the same way described above with ExportGenesis and iterating of the genesis accounts on the working example
	storage = evmtypes.Storage{
		{Key: "0x0000000000000000000000000000000000000000000000000000000000000005", Value: "0x786d706c00000000000000000000000000000000000000000000000000000008"}, //gitleaks:allow
		{Key: "0x0000000000000000000000000000000000000000000000000000000000000006", Value: "0x786d706c00000000000000000000000000000000000000000000000000000008"}, //gitleaks:allow
		{Key: "0x0000000000000000000000000000000000000000000000000000000000000007", Value: "0x0000000000000000000000000000000000000000000000000000000000000600"}, //gitleaks:allow
		{Key: "0x0eb5be412f275a18f6e4d622aee4ff40b21467c926224771b782d4c095d1444b", Value: "0x00000000000000000000000047eeb2eac350e1923b8cbdfa4396a077b36e62a0"}, //gitleaks:allow
		{Key: "0x0ef55e0cc676cf0b0dcbdc7d53c3b797e88e44b10ecee85d45144bc7392574c7", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
		{Key: "0x26bde78f605f19d1d853933fa781096670ea82ad96c9a3fb49f407e9600b316a", Value: "0x00000000000000000000000047eeb2eac350e1923b8cbdfa4396a077b36e62a0"}, //gitleaks:allow
		{Key: "0x28bcb2563cf7895ce732c75018c5f73c44037088fc4505201fc28c3147d1d4a0", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
		{Key: "0x2e681a43037a582fb0535432b520e9dc97303bf239f82648091f9be82f4b677e", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
		{Key: "0x42aa905ad67e072b45e9dae81c8fa8bc705cc25014d5e9f10fa5114ae9c0dcf1", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
		{Key: "0x4796a5437e25bdc491b74d328cf6b437c8587e216f52049c7df56421f51ae30f", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
		{Key: "0x64e21244e91af723e1b962171ed4828dcecc0d7b89872e516a5db8266da80000", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
		{Key: "0x6d2487ab6e76634bbe98b9d5b39803d625a8f9249da9e03e70c638bf76d9e29b", Value: "0x00000000000000000000000047eeb2eac350e1923b8cbdfa4396a077b36e62a0"}, //gitleaks:allow
		{Key: "0x97c12224ac75d13d2d9a9e30dd25212b81a4e2c19743b211209b4bca3db99142", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
		{Key: "0xa6eef7e35abe7026729641147f7915573c7e97b47efa546f5f6e3230263bcb49", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
		{Key: "0xb009fbc347bffd144efd545cc4b15a37592e1dd7063753564d9ecc6fea764b6f", Value: "0x00000000000000000000000047eeb2eac350e1923b8cbdfa4396a077b36e62a0"}, //gitleaks:allow
		{Key: "0xb9cbbae02fe941283ec0eefd7b121e3bc7f89fae077b27bdd75a7fd4cf1543a8", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
		{Key: "0xc5724e8640ef1f7915e4839c81ad4b592af3c601230608793acd429a848553e9", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
		{Key: "0xfac4953099c6f6272238a038333d99c9cd0475cb85c72b761909fadae4b6cbcd", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
		{Key: "0xfd061ffb53f8d83182630fadee503ca39e0bae885f162932fe84d25caddbc888", Value: "0x0000000000000000000000000000000000000000000000000000000000000001"}, //gitleaks:allow
	}

	// erc20Addr is the address of the ERC-20 contract.
	erc20Addr = common.HexToAddress(erc20ContractHex)

	// smartContractAddress is the bech32 address of the Ethereum smart contract account that is set in genesis.
	smartContractAddress = utils.EthToCosmosAddr(erc20Addr).String()
)

// createGenesisWithERC20 creates a genesis state that contains the state containing an EthAccount that is a smart contract.
func createGenesisWithERC20(keyring testkeyring.Keyring) testnetwork.CustomGenesisState {
	genesisAccounts := []authtypes.AccountI{
		&authtypes.BaseAccount{
			Address:       baseAccountAddress,
			PubKey:        nil,
			AccountNumber: 0,
			Sequence:      1,
		},
		&evmostypes.EthAccount{
			BaseAccount: &authtypes.BaseAccount{
				Address:       smartContractAddress,
				PubKey:        nil,
				AccountNumber: 9,
				Sequence:      1,
			},
			CodeHash: codeHash.String(),
		},
	}

	// Add all keys from the keyring to the genesis accounts as well.
	//
	// NOTE: This is necessary to enable the account to send EVM transactions,
	// because the Mono ante handler checks the account balance by querying the
	// account from the account keeper first. If these accounts are not in the genesis
	// state, the ante handler finds a zero balance because of the missing account.
	for i, addr := range keyring.GetAllAccAddrs() {
		genesisAccounts = append(genesisAccounts, &authtypes.BaseAccount{
			Address:       addr.String(),
			PubKey:        nil,
			AccountNumber: uint64(i + 1),
			Sequence:      1,
		})
	}

	accGenesisState := authtypes.DefaultGenesisState()
	for _, genesisAccount := range genesisAccounts {
		// NOTE: This type requires to be packed into a *types.Any as seen on SDK tests,
		// e.g. https://github.com/evmos/cosmos-sdk/blob/v0.47.5-evmos.2/x/auth/keeper/keeper_test.go#L193-L223
		accGenesisState.Accounts = append(accGenesisState.Accounts, codectypes.UnsafePackAny(genesisAccount))
	}

	// Add the smart contracts to the EVM genesis
	evmGenesisState := evmtypes.DefaultGenesisState()
	evmGenesisState.Accounts = append(evmGenesisState.Accounts, evmtypes.GenesisAccount{
		Address: erc20ContractHex,
		Code:    smartContractCode,
		Storage: storage,
	})

	// Combine module genesis states
	return testnetwork.CustomGenesisState{
		authtypes.ModuleName: accGenesisState,
		evmtypes.ModuleName:  evmGenesisState,
	}
}

func TestMigrateEthAccountsToBaseAccounts(t *testing.T) {
	keyring := testkeyring.New(1)
	network := testnetwork.NewUnitTestNetwork(
		// NOTE: This genesis was created using pre-EthAccount removal code
		// so that the accounts stored are actual EthAccounts.
		testnetwork.WithCustomGenesis(createGenesisWithERC20(keyring)),
	)

	require.NoError(t, network.NextBlock(), "failed to advance block")

	// Check the contract is an EthAccount before migration
	erc20AddrAsBech32 := utils.EthToCosmosAddr(erc20Addr)
	acc := network.App.AccountKeeper.GetAccount(network.GetContext(), erc20AddrAsBech32)
	require.NotNil(t, acc, "account not found")
	require.IsType(t, &evmostypes.EthAccount{}, acc, "expected account to be an EthAccount")

	// Migrate the accounts
	v19.MigrateEthAccountsToBaseAccounts(network.GetContext(), network.App.AccountKeeper, network.App.EvmKeeper)

	// Check the contract is a BaseAccount after migration
	acc = network.App.AccountKeeper.GetAccount(network.GetContext(), erc20AddrAsBech32)
	require.NotNil(t, acc, "account not found")
	require.IsType(t, &authtypes.BaseAccount{}, acc, "account should be a base account")

	// Check that the keeper has the new store entry
	require.Equal(t,
		codeHash.String(),
		network.App.EvmKeeper.GetCodeHash(network.GetContext(), erc20Addr).String(),
		"expected different code hash",
	)

	require.Equal(t,
		code,
		network.App.EvmKeeper.GetCode(network.GetContext(), codeHash),
		"expected different code",
	)
}
