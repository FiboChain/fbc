package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/FiboChain/fbc/x/dex/types"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/types/rest"
	"github.com/FiboChain/fbc/x/common"
	govRest "github.com/FiboChain/fbc/x/gov/client/rest"

	"github.com/gorilla/mux"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/context"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/products", productsHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/dex/deposits", depositsHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/dex/product_rank", matchOrderHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/dexoperator/{address}", operatorHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/dexoperators", operatorsHandler(cliCtx)).Methods("GET")
}

func productsHandler(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerAddress := r.URL.Query().Get("address")
		pageStr := r.URL.Query().Get("page")
		perPageStr := r.URL.Query().Get("per_page")

		page, perPage, err := common.Paginate(pageStr, perPageStr)
		if err != nil {
			common.HandleErrorMsg(w, cliContext, common.CodeInvalidPaginateParam, err.Error())
			return
		}
		params := types.NewQueryDexInfoParams(ownerAddress, page, perPage)
		bz, err := cliContext.Codec.MarshalJSON(&params)

		if err != nil {
			common.HandleErrorMsg(w, cliContext, common.CodeMarshalJSONFailed, err.Error())
			return
		}

		res, _, err := cliContext.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryProducts), bz)
		if err != nil {
			sdkErr := common.ParseSDKError(err.Error())
			common.HandleErrorMsg(w, cliContext, sdkErr.Code, sdkErr.Message)
			return
		}

		rest.PostProcessResponse(w, cliContext, res)
	}

}

func depositsHandler(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		address := r.URL.Query().Get("address")
		baseAsset := r.URL.Query().Get("base_asset")
		quoteAsset := r.URL.Query().Get("quote_asset")
		pageStr := r.URL.Query().Get("page")
		perPageStr := r.URL.Query().Get("per_page")
		if address == "" && baseAsset == "" && quoteAsset == "" {
			common.HandleErrorMsg(w, cliContext, types.CodeAddrAndProductAllRequired, "bad request: address、base_asset and quote_asset could not be empty at the same time")
			return
		}
		page, perPage, err := common.Paginate(pageStr, perPageStr)
		if err != nil {
			common.HandleErrorMsg(w, cliContext, common.CodeInvalidPaginateParam, err.Error())
			return
		}

		params := types.NewQueryDepositParams(address, baseAsset, quoteAsset, page, perPage)
		bz, err := cliContext.Codec.MarshalJSON(&params)
		if err != nil {
			common.HandleErrorMsg(w, cliContext, common.CodeMarshalJSONFailed, err.Error())
			return
		}

		res, _, err := cliContext.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDeposits), bz)
		if err != nil {
			sdkErr := common.ParseSDKError(err.Error())
			common.HandleErrorMsg(w, cliContext, sdkErr.Code, sdkErr.Message)
			return
		}

		rest.PostProcessResponse(w, cliContext, res)
	}

}

func matchOrderHandler(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		perPageStr := r.URL.Query().Get("per_page")

		page, perPage, err := common.Paginate(pageStr, perPageStr)
		if err != nil {
			common.HandleErrorMsg(w, cliContext, common.CodeInvalidPaginateParam, err.Error())
			return
		}

		params := types.NewQueryDexInfoParams("", page, perPage)
		bz, err := cliContext.Codec.MarshalJSON(&params)

		if err != nil {
			common.HandleErrorMsg(w, cliContext, common.CodeMarshalJSONFailed, err.Error())
			return
		}

		res, _, err := cliContext.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryMatchOrder), bz)
		if err != nil {
			sdkErr := common.ParseSDKError(err.Error())
			common.HandleErrorMsg(w, cliContext, sdkErr.Code, sdkErr.Message)
			return
		}

		result := common.GetBaseResponse("hello")
		result2, err2 := json.Marshal(result)
		if err2 != nil {
			common.HandleErrorMsg(w, cliContext, common.CodeMarshalJSONFailed, err2.Error())
			return
		}
		result2 = []byte(strings.Replace(string(result2), "\"hello\"", string(res), 1))
		rest.PostProcessResponse(w, cliContext, result2)
	}

}

func operatorHandler(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		address, err := sdk.AccAddressFromBech32(vars["address"])
		if err != nil {
			common.HandleErrorMsg(w, cliContext, common.CodeCreateAddrFromBech32Failed, err.Error())
			return
		}

		params := types.QueryDexOperatorParams{}
		params.Addr = address
		bz := cliContext.Codec.MustMarshalJSON(&params)
		res, _, err := cliContext.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryOperator), bz)
		if err != nil {
			sdkErr := common.ParseSDKError(err.Error())
			common.HandleErrorMsg(w, cliContext, sdkErr.Code, sdkErr.Message)
			return
		}

		result := common.GetBaseResponse("hello")
		result2, err2 := json.Marshal(result)
		if err2 != nil {
			common.HandleErrorMsg(w, cliContext, common.CodeMarshalJSONFailed, err2.Error())
			return
		}
		result2 = []byte(strings.Replace(string(result2), "\"hello\"", string(res), 1))
		rest.PostProcessResponse(w, cliContext, result2)
	}
}

func operatorsHandler(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := cliContext.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryOperators))
		if err != nil {
			sdkErr := common.ParseSDKError(err.Error())
			common.HandleErrorMsg(w, cliContext, sdkErr.Code, sdkErr.Message)
			return
		}

		result := common.GetBaseResponse("hello")
		result2, err2 := json.Marshal(result)
		if err2 != nil {
			common.HandleErrorMsg(w, cliContext, common.CodeMarshalJSONFailed, err2.Error())
			return
		}
		result2 = []byte(strings.Replace(string(result2), "\"hello\"", string(res), 1))
		rest.PostProcessResponse(w, cliContext, result2)

	}
}

// DelistProposalRESTHandler defines dex proposal handler
func DelistProposalRESTHandler(context.CLIContext) govRest.ProposalRESTHandler {
	return govRest.ProposalRESTHandler{}
}
