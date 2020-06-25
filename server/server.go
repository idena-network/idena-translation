package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/idena-network/idena-translation/config"
	"github.com/idena-network/idena-translation/core"
	"github.com/idena-network/idena-translation/docs"
	"github.com/idena-network/idena-translation/types"
	log "github.com/inconshreveable/log15"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"strings"
	"sync"
)

type Server struct {
	port       int
	engine     core.Engine
	mutex      sync.Mutex
	counter    int
	httpServer *http.Server
}

func NewServer(port int, engine core.Engine) *Server {
	return &Server{
		port:   port,
		engine: engine,
	}
}

func (s *Server) Start(swaggerConfig config.SwaggerConfig) {
	router := mux.NewRouter()
	s.initRouter(router)
	if swaggerConfig.Enabled {
		docs.SwaggerInfo.Title = "Idena flip words translation API"
		docs.SwaggerInfo.Version = types.AppVersion
		docs.SwaggerInfo.Host = swaggerConfig.Host
		docs.SwaggerInfo.BasePath = swaggerConfig.BasePath
		router.PathPrefix("/swagger").Handler(httpSwagger.Handler(
			httpSwagger.URL("doc.json"),
		))
	}
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	addr := fmt.Sprintf(":%d", s.port)
	handler := handlers.CORS(originsOk, headersOk, methodsOk)(s.requestFilter(router))
	httpServer := &http.Server{Addr: addr, Handler: handler}
	s.httpServer = httpServer
	log.Info(fmt.Sprintf("Server started at port %v", s.port))
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func (s *Server) Stop() {
	if s.httpServer == nil {
		return
	}
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		panic(err)
	}
}

func (s *Server) requestFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(strings.ToLower(r.URL.Path), "/swagger") {
			reqId := s.generateReqId()
			log.Debug(fmt.Sprintf("Got request %v, url: %v, from: %v", reqId, r.URL, GetIP(r)))
			defer log.Debug(fmt.Sprintf("Completed request %v", reqId))
			err := r.ParseForm()
			if err != nil {
				writeErrResponse(w, reqId, http.StatusBadRequest, err.Error())
				return
			}
			r.URL.Path = strings.ToLower(r.URL.Path)
			ctx := r.Context()
			ctx = context.WithValue(ctx, "reqId", reqId)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) generateReqId() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	id := s.counter
	s.counter++
	return id
}

func GetIP(r *http.Request) string {
	header := r.Header.Get("X-Forwarded-For")
	if len(header) > 0 {
		return strings.Split(header, ", ")[0]
	}
	if strings.Contains(r.RemoteAddr, ":") {
		return strings.Split(r.RemoteAddr, ":")[0]
	}
	return r.RemoteAddr
}

func (s *Server) initRouter(router *mux.Router) {
	router.Path(strings.ToLower("/translation")).HandlerFunc(s.submitTranslation).Methods("POST")
	router.Path(strings.ToLower("/word/{word:[0-9]+}/language/{language}/translations")).
		HandlerFunc(s.getTranslations).Methods("GET")
	router.Path(strings.ToLower("/vote")).HandlerFunc(s.vote).Methods("POST")
	router.Path(strings.ToLower("/word/{word:[0-9]+}/language/{language}/confirmed-translation")).HandlerFunc(s.confirmedTranslation).Methods("GET")
}

func writeErrResponse(w http.ResponseWriter, reqId int, code int, errMessage string) {
	log.Error(fmt.Sprintf("Unable to handle request %v: %v", reqId, errMessage))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	writeResponseBody(w, reqId, types.ErrorResponse{
		Error: errMessage,
	})
}

func writeResponse(w http.ResponseWriter, reqId int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	writeResponseBody(w, reqId, body)
}

func writeResponseBody(w http.ResponseWriter, reqId int, body interface{}) {
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to write response to request %v: %v", reqId, err))
		return
	}
}
