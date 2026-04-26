package types

import (
	"testing"

	s "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// init configures the global Bech32 prefix to match the live chain
// (addr_safro). FeeShare validation round-trips every contract,
// deployer, and withdrawer through sdk.AccAddressFromBech32 which
// compares the prefix to the value on sdk.Config; tests in `package
// types` do not run app bootstrap so we set it explicitly here.
func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("addr_safro", "addr_safropub")
}

type GenesisTestSuite struct {
	s.Suite
	address1  string
	address2  string
	contractA string
	contractB string
}

func TestGenesisTestSuite(t *testing.T) {
	s.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) SetupTest() {
	suite.address1 = sdk.AccAddress([]byte("addr_safro1")).String()
	suite.address2 = sdk.AccAddress([]byte("addr_safro2")).String()

	// 32-byte contract addresses bech32-encoded with the addr_safro HRP.
	suite.contractA = "addr_safro1fz7eevr8l8cfpk7a3qv0xgyl4fghhnsf44xs7jly3p48wwge8jfs7epqm2"
	suite.contractB = "addr_safro17nerj8pwku9uwfwct973pjk3ypmxxfwatwvvj9kq3832w4yd700qwkw203"
}

func (suite *GenesisTestSuite) TestValidateGenesis() {
	newGen := NewGenesisState(DefaultParams(), []FeeShare{})
	testCases := []struct {
		name     string
		genState *GenesisState
		expPass  bool
	}{
		{
			name:     "valid genesis constructor",
			genState: &newGen,
			expPass:  true,
		},
		{
			name:     "default",
			genState: DefaultGenesisState(),
			expPass:  true,
		},
		{
			name: "valid genesis",
			genState: &GenesisState{
				Params:   DefaultParams(),
				FeeShare: []FeeShare{},
			},
			expPass: true,
		},
		{
			name: "valid genesis - with fee",
			genState: &GenesisState{
				Params: DefaultParams(),
				FeeShare: []FeeShare{
					{
						ContractAddress:   suite.contractA,
						DeployerAddress:   suite.address1,
						WithdrawerAddress: suite.address1,
					},
					{
						ContractAddress:   suite.contractB,
						DeployerAddress:   suite.address2,
						WithdrawerAddress: suite.address2,
					},
				},
			},
			expPass: true,
		},
		{
			name:     "empty genesis",
			genState: &GenesisState{},
			expPass:  false,
		},
		{
			name: "invalid genesis - duplicated fee",
			genState: &GenesisState{
				Params: DefaultParams(),
				FeeShare: []FeeShare{
					{
						ContractAddress: suite.contractA,
						DeployerAddress: suite.address1,
					},
					{
						ContractAddress: suite.contractA,
						DeployerAddress: suite.address1,
					},
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - duplicated fee with different deployer address",
			genState: &GenesisState{
				Params: DefaultParams(),
				FeeShare: []FeeShare{
					{
						ContractAddress: suite.contractA,
						DeployerAddress: suite.address1,
					},
					{
						ContractAddress: suite.contractA,
						DeployerAddress: suite.address2,
					},
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - invalid withdrawer address",
			genState: &GenesisState{
				Params: DefaultParams(),
				FeeShare: []FeeShare{
					{
						ContractAddress:   suite.contractA,
						DeployerAddress:   suite.address1,
						WithdrawerAddress: "withdraw",
					},
				},
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		err := tc.genState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
