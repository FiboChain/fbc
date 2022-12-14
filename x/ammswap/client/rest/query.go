package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/context"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/types/rest"
	"github.com/FiboChain/fbc/x/ammswap/types"
	"github.com/FiboChain/fbc/x/common"
	"github.com/gorilla/mux"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r = r.PathPrefix("/swap").Subrouter()
	r.HandleFunc("/token_pair/{name}", querySwapTokenPairHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/params", queryParamsHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/liquidity/add_quote/{token}", swapAddQuoteHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/liquidity/remove_quote/{token_pair}", queryRedeemableAssetsHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/quote/{token}", swapQuoteHandler(cliCtx)).Methods("GET")
}

func querySwapTokenPairHandler(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tokenPairName := vars["name"]
		res, _, err := cliContext.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, types.QuerySwapTokenPair, tokenPairName), nil)
		if err != nil {
			sdkErr := common.ParseSDKError(err.Error())
			common.HandleErrorMsg(w, cliContext, sdkErr.Code, sdkErr.Message)
			return
		}
		rest.PostProcessResponse(w, cliContext, res)
	}

}

func queryParamsHandler(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		res, _, err := cliContext.QueryWithData(fmt.Sprintf("custom/%s/params", types.QuerierRoute), nil)
		if err != nil {
			sdkErr := common.ParseSDKError(err.Error())
			common.HandleErrorMsg(w, cliContext, sdkErr.Code, sdkErr.Message)
			return
		}

		formatAndReturnResult(w, cliContext, res)
	}

}

func queryRedeemableAssetsHandler(cliContext context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tokenPair := vars["token_pair"]
		liquidity := r.URL.Query().Get("liquidity")
		res, _, err := cliContext.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, types.QueryRedeemableAssets, tokenPair, liquidity), nil)
		if err != nil {
			sdkErr := common.ParseSDKError(err.Error())
			common.HandleErrorMsg(w, cliContext, sdkErr.Code, sdkErr.Message)
			return
		}
		formatAndReturnResult(w, cliContext, res)
	}

}

func formatAndReturnResult(w http.ResponseWriter, cliContext context.CLIContext, data []byte) {
	replaceStr := "replaceHere"
	result := common.GetBaseResponse(replaceStr)
	resultJson, err := json.Marshal(result)
	if err != nil {
		common.HandleErrorMsg(w, cliContext, common.CodeMarshalJSONFailed, err.Error())
		return
	}
	resultJson = []byte(strings.Replace(string(resultJson), "\""+replaceStr+"\"", string(data), 1))

	rest.PostProcessResponse(w, cliContext, resultJson)
}

func swapQuoteHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		buyToken := vars["token"]
		sellTokenAmount := r.URL.Query().Get("sell_token_amount")

		params := types.NewQuerySwapBuyInfoParams(sellTokenAmount, buyToken)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			common.HandleErrorMsg(w, cliCtx, common.CodeMarshalJSONFailed, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QuerySwapQuoteInfo), bz)
		if err != nil {
			sdkErr := common.ParseSDKError(err.Error())
			common.HandleErrorMsg(w, cliCtx, sdkErr.Code, sdkErr.Message)
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func swapAddQuoteHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		baseToken := vars["token"]
		quoteTokenAmount := r.URL.Query().Get("quote_token_amount")

		params := types.NewQuerySwapAddInfoParams(quoteTokenAmount, baseToken)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			common.HandleErrorMsg(w, cliCtx, common.CodeMarshalJSONFailed, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QuerySwapAddLiquidityQuote), bz)
		if err != nil {
			sdkErr := common.ParseSDKError(err.Error())
			common.HandleErrorMsg(w, cliCtx, sdkErr.Code, sdkErr.Message)
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
