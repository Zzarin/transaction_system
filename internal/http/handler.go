package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Zzarin/transaction_system/internal"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Handler struct {
	handler *http.ServeMux
}

type incRequest struct {
	ClientName string  `json:"clientName"`
	OrderID    int     `json:"orderID"`
	TypeTXN    string  `json:"typeTXN"`
	Amount     float64 `json:"amount"`
}

func incReqToIncStruct(incData incRequest, clientID string) (*internal.IncStruct, error) {
	clientIDInt, err := strconv.Atoi(clientID)
	if err != nil {
		return nil, fmt.Errorf("converting clientID to string %v", zap.Error(err))
	}

	return &internal.IncStruct{
		ClientID:   clientIDInt,
		ClientName: incData.ClientName,
		OrderID:    incData.OrderID,
		TypeTXN:    incData.TypeTXN,
		Amount:     incData.Amount,
	}, nil
}

func InitEndpoints(ctx context.Context, account internal.AccountService, logger *zap.Logger) *http.ServeMux {
	newHandler := &Handler{
		handler: http.NewServeMux(),
	}
	newHandler.handler.Handle("/domain/pay", Todo(ctx, account, logger))
	return newHandler.handler
}

func Todo(ctx context.Context, account internal.AccountService, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxRequest, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		urlString, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			logger.Error("query string parsing", zap.Error(err))
			http.Error(w, "query string parsing", http.StatusBadRequest)
			return
		}

		respInBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error("reading request body", zap.Error(err))
			http.Error(w, "reading request body", http.StatusBadRequest)
			return
		}

		var incData incRequest
		err = json.Unmarshal(respInBytes, &incData)
		if err != nil {
			logger.Error("unmarshalling JSON", zap.Error(err))
			http.Error(w, "unmarshalling JSON error", http.StatusBadRequest)
			return
		}
		domainStruct, err := incReqToIncStruct(incData, urlString["client"][0])
		if err != nil {
			logger.Error("converting to domain struct", zap.Error(err))
			http.Error(w, "converting to domain struct", http.StatusBadRequest)
			return
		}

		result, err := account.BalanceSufficient(ctxRequest, domainStruct)
		if err != nil {
			logger.Error("checking balance", zap.Error(err))
			http.Error(w, "check your balance", http.StatusBadGateway)
			return
		}
		if result == false {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusOK)
		return

	})
}
