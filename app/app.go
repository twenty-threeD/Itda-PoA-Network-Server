package app

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	dbm "github.com/cosmos/cosmos-db"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtos "github.com/cometbft/cometbft/libs/os"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	authmodule "github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankmodule "github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusmodule "github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment"
	paymentkeeper "github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/keeper"
	paymenttypes "github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/types"
	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa"
	poakeeper "github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/keeper"
	poatypes "github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/types"
)

var (
	DefaultNodeHome string

	maccPerms = map[string][]string{
		authtypes.FeeCollectorName: nil,
	}
)

var _ servertypes.Application = (*ItdaApp)(nil)

type ItdaApp struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	keys map[string]*storetypes.KVStoreKey

	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.BaseKeeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	PoaKeeper             poakeeper.Keeper
	PaymentKeeper         paymentkeeper.Keeper

	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

	configurator module.Configurator
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".itdad")
}

func NewItdaApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *ItdaApp {
	encodingConfig := MakeEncodingConfig()

	appCodec := encodingConfig.Codec
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig

	bApp := baseapp.NewBaseApp(
		Name,
		logger,
		db,
		txConfig.TxDecoder(),
		baseAppOptions...,
	)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		consensusparamtypes.StoreKey,
		poatypes.StoreKey,
		paymenttypes.StoreKey,
	)

	app := &ItdaApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
	}

	moduleAuthority := authtypes.NewModuleAddress(Name).String()

	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		moduleAuthority,
		runtime.EventService{},
	)
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec(Bech32Prefix),
		Bech32Prefix,
		moduleAuthority,
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		BlockedAddresses(),
		moduleAuthority,
		logger,
	)

	app.PoaKeeper = poakeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[poatypes.StoreKey]),
		authcodec.NewBech32Codec(Bech32Prefix),
		logger,
	)

	app.PaymentKeeper = paymentkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[paymenttypes.StoreKey]),
		authcodec.NewBech32Codec(Bech32Prefix),
		logger,
	)

	app.ModuleManager = module.NewManager(
		authmodule.NewAppModule(appCodec, app.AccountKeeper, nil, nil),
		bankmodule.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, nil),
		consensusmodule.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		poa.NewAppModule(appCodec, app.PoaKeeper),
		payment.NewAppModule(appCodec, app.PaymentKeeper),
	)

	app.BasicModuleManager = module.NewBasicManagerFromManager(app.ModuleManager, nil)

	moduleOrder := []string{
		consensusparamtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		paymenttypes.ModuleName,
		poatypes.ModuleName,
	}

	app.ModuleManager.SetOrderBeginBlockers(moduleOrder...)
	app.ModuleManager.SetOrderEndBlockers(moduleOrder...)
	app.ModuleManager.SetOrderInitGenesis(moduleOrder...)
	app.ModuleManager.SetOrderExportGenesis(moduleOrder...)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())

	if err := app.ModuleManager.RegisterServices(app.configurator); err != nil {
		panic(err)
	}

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}

	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	app.MountKVStores(keys)

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.setAnteHandler(txConfig)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			cmtos.Exit(err.Error())
		}
	}

	return app
}

func (app *ItdaApp) setAnteHandler(txConfig client.TxConfig) {
	anteHandler, err := NewAnteHandler(HandlerOptions{
		AccountKeeper:   app.AccountKeeper,
		BankKeeper:      app.BankKeeper,
		SignModeHandler: txConfig.SignModeHandler(),
	})
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
}

func (app *ItdaApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState

	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}

	return app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
}

func (app *ItdaApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

func (app *ItdaApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

func (app *ItdaApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

func (app *ItdaApp) AppCodec() codec.Codec {
	return app.appCodec
}

func (app *ItdaApp) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

func (app *ItdaApp) TxConfig() client.TxConfig {
	return app.txConfig
}

func (app *ItdaApp) SimulationManager() *module.SimulationManager {
	return nil
}

func (app *ItdaApp) RegisterAPIRoutes(apiSvr *api.Server, _ config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

func (app *ItdaApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

func (app *ItdaApp) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

func (app *ItdaApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

func (app *ItdaApp) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)

	for _, m := range app.ModuleManager.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleWithName.Name()] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          authcodec.NewBech32Codec(Bech32Prefix),
		ValidatorAddressCodec: authcodec.NewBech32Codec(Bech32Prefix + "valoper"),
		ConsensusAddressCodec: authcodec.NewBech32Codec(Bech32Prefix + "valcons"),
	}
}

func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)

	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}
