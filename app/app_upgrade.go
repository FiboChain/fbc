package app

import (
	"sort"

	sdkerrors "github.com/FiboChain/fbc/libs/cosmos-sdk/types/errors"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/client/utils"

	cliContext "github.com/FiboChain/fbc/libs/cosmos-sdk/client/context"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/store/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/types/module"
	upgradetypes "github.com/FiboChain/fbc/libs/cosmos-sdk/types/upgrade"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/params"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/params/subspace"
)

func (app *FBChainApp) RegisterTxService(clientCtx cliContext.CLIContext) {
	utils.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.grpcSimulate, clientCtx.InterfaceRegistry)
}
func (app *FBChainApp) grpcSimulate(txBytes []byte) (sdk.GasInfo, *sdk.Result, error) {
	tx, err := app.GetTxDecoder()(txBytes)
	if err != nil {
		return sdk.GasInfo{}, nil, sdkerrors.Wrap(err, "failed to decode tx")
	}
	return app.Simulate(txBytes, tx, 0, nil)
}

func (app *FBChainApp) setupUpgradeModules() {
	heightTasks, paramMap, cf, pf, vf := app.CollectUpgradeModules(app.mm)

	app.heightTasks = heightTasks

	app.GetCMS().AppendCommitFilters(cf)
	app.GetCMS().AppendPruneFilters(pf)
	app.GetCMS().AppendVersionFilters(vf)

	vs := app.subspaces
	for k, vv := range paramMap {
		supace, exist := vs[k]
		if !exist {
			continue
		}
		vs[k] = supace.LazyWithKeyTable(subspace.NewKeyTable(vv.ParamSetPairs()...))
	}
}

func (o *FBChainApp) CollectUpgradeModules(m *module.Manager) (map[int64]*upgradetypes.HeightTasks,
	map[string]params.ParamSet, []types.StoreFilter, []types.StoreFilter, []types.VersionFilter) {
	hm := make(map[int64]*upgradetypes.HeightTasks)
	paramsRet := make(map[string]params.ParamSet)
	commitFiltreMap := make(map[*types.StoreFilter]struct{})
	pruneFilterMap := make(map[*types.StoreFilter]struct{})
	versionFilterMap := make(map[*types.VersionFilter]struct{})

	for _, mm := range m.Modules {
		if ada, ok := mm.(upgradetypes.UpgradeModule); ok {
			set := ada.RegisterParam()
			if set != nil {
				if _, exist := paramsRet[ada.ModuleName()]; !exist {
					paramsRet[ada.ModuleName()] = set
				}
			}
			h := ada.UpgradeHeight()
			if h > 0 {
				h++
			}

			cf := ada.CommitFilter()
			if cf != nil {
				if _, exist := commitFiltreMap[cf]; !exist {
					commitFiltreMap[cf] = struct{}{}
				}
			}
			pf := ada.PruneFilter()
			if pf != nil {
				if _, exist := pruneFilterMap[pf]; !exist {
					pruneFilterMap[pf] = struct{}{}
				}
			}
			vf := ada.VersionFilter()
			if vf != nil {
				if _, exist := versionFilterMap[vf]; !exist {
					versionFilterMap[vf] = struct{}{}
				}
			}

			t := ada.RegisterTask()
			if t == nil {
				continue
			}
			if err := t.ValidateBasic(); nil != err {
				panic(err)
			}
			taskList := hm[h]
			if taskList == nil {
				v := make(upgradetypes.HeightTasks, 0)
				taskList = &v
				hm[h] = taskList
			}
			*taskList = append(*taskList, t)
		}
	}

	for _, v := range hm {
		sort.Sort(*v)
	}

	commitFilters := make([]types.StoreFilter, 0)
	pruneFilters := make([]types.StoreFilter, 0)
	versionFilters := make([]types.VersionFilter, 0)
	for pointerFilter, _ := range commitFiltreMap {
		commitFilters = append(commitFilters, *pointerFilter)
	}
	for pointerFilter, _ := range pruneFilterMap {
		pruneFilters = append(pruneFilters, *pointerFilter)
	}
	for pointerFilter, _ := range versionFilterMap {
		versionFilters = append(versionFilters, *pointerFilter)
	}

	return hm, paramsRet, commitFilters, pruneFilters, versionFilters
}
