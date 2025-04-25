package http

import (
	"ChaikaReports/internal/handler/http/decoder"
	"ChaikaReports/internal/handler/http/encoder"
	"ChaikaReports/internal/service"
	"encoding/json"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

const (
	apiPrefix = "/api"
	v1Prefix  = apiPrefix + "/v1/report"
)

func NewHTTPHandler(svc service.SalesService, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc(apiPrefix, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"versions":      []string{"v1"},
			"documentation": v1Prefix + "/docs/index.html",
		})
		if err != nil {
			_ = level.Error(logger).Log("msg", "failed to encode API info", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("GET")

	registerV1Routes(logger, r, svc)

	return r
}

func registerV1Routes(logger log.Logger, router *mux.Router, svc service.SalesService) {
	v1 := router.PathPrefix(v1Prefix).Subrouter()

	v1.PathPrefix("/docs/").Handler(httpSwagger.Handler(
		httpSwagger.URL(v1Prefix+"/docs/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	))

	v1.Methods("POST").Path("/sale").Handler(kitHttp.NewServer(
		MakeInsertSalesEndpoint(svc),
		decoder.DecodeInsertSalesRequest,
		encoder.EncodeResponse,
		kitHttp.ServerErrorEncoder(encoder.EncodeError(logger)),
	))

	v1.Methods("GET").Path("/trip/cart/employee").Handler(kitHttp.NewServer(
		MakeGetEmployeeCartsInTripEndpoint(svc),
		decoder.DecodeGetEmployeeCartsInTripRequest,
		encoder.EncodeResponse,
		kitHttp.ServerErrorEncoder(encoder.EncodeError(logger)),
	))

	v1.Methods("GET").Path("/trip/employee_id").Handler(kitHttp.NewServer(
		MakeGetEmployeeIDsByTripEndpoint(svc),
		decoder.DecodeGetEmployeeIDsByTripRequest,
		encoder.EncodeResponse,
		kitHttp.ServerErrorEncoder(encoder.EncodeError(logger)),
	))

	v1.Methods("GET").Path("/trip/employee_trip").Handler(kitHttp.NewServer(
		MakeGetEmployeeTripsEndpoint(svc),
		decoder.DecodeGetEmployeeTripsRequest,
		encoder.EncodeResponse,
		kitHttp.ServerErrorEncoder(encoder.EncodeError(logger)),
	))

	v1.Methods("PUT").Path("/trip/cart/item/quantity").Handler(kitHttp.NewServer(
		MakeUpdateItemQuantityEndpoint(svc),
		decoder.DecodeUpdateItemQuantityRequest,
		encoder.EncodeResponse,
		kitHttp.ServerErrorEncoder(encoder.EncodeError(logger)),
	))

	v1.Methods("DELETE").Path("/trip/cart/item").Handler(kitHttp.NewServer(
		MakeDeleteItemFromCartEndpoint(svc),
		decoder.DecodeDeleteItemFromCartRequest,
		encoder.EncodeResponse,
		kitHttp.ServerErrorEncoder(encoder.EncodeError(logger)),
	))
}
