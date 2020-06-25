package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/idena-network/idena-translation/types"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

// @Tags Translation
// @Id submitTranslation
// @Summary Create or update translation
// @Param translation body types.SubmitTranslationRequest true "translation details"
// @Success 200 {object} types.SubmitTranslationResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /translation [post]
func (s *Server) submitTranslation(w http.ResponseWriter, r *http.Request) {
	reqId, _ := r.Context().Value("reqId").(int)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeErrResponse(w, reqId, http.StatusInternalServerError, err.Error())
		return
	}
	request := types.SubmitTranslationRequest{}
	if err := json.Unmarshal(body, &request); err != nil {
		writeErrResponse(w, reqId, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.engine.SubmitTranslation(request)
	if err != nil {
		if _, ok := err.(*types.BadRequestError); ok {
			writeErrResponse(w, reqId, http.StatusBadRequest, err.Error())
			return
		}
		writeErrResponse(w, reqId, http.StatusInternalServerError, err.Error())
		return
	}
	writeResponse(w, reqId, response)
}

// @Tags Translation
// @Id getTranslations
// @Summary Get translations sorted by rating
// @Param word path integer true "word id"
// @Param language path string true "language"
// @Param continuation-token header string false "continuation token to get next translations"
// @Success 200 {object} types.GetTranslationsResponse
// @Header 200 {string} continuation-token "continuation token"
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /word/{word}/language/{language}/translations [get]
func (s *Server) getTranslations(w http.ResponseWriter, r *http.Request) {
	reqId, _ := r.Context().Value("reqId").(int)
	vars := mux.Vars(r)
	wordId, err := toUint(vars, "word")
	if err != nil {
		writeErrResponse(w, reqId, http.StatusBadRequest, err.Error())
		return
	}
	response, continuationToken, err := s.engine.GetTranslations(uint32(wordId), mux.Vars(r)["language"], r.Header.Get("continuation-token"))
	if err != nil {
		if _, ok := err.(*types.BadRequestError); ok {
			writeErrResponse(w, reqId, http.StatusBadRequest, err.Error())
			return
		}
		writeErrResponse(w, reqId, http.StatusInternalServerError, err.Error())
		return
	}
	if len(continuationToken) > 0 {
		w.Header().Set("continuation-token", continuationToken)
	}
	writeResponse(w, reqId, response)
}

func toUint(vars map[string]string, name string) (uint64, error) {
	value, err := strconv.ParseUint(vars[name], 10, 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("invalid value '%v'", name))
	}
	return value, nil
}

// @Tags Translation
// @Id vote
// @Summary Vote for or against translation
// @Param vote body types.VoteRequest true "vote details"
// @Success 200 {object} types.VoteResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /vote [post]
func (s *Server) vote(w http.ResponseWriter, r *http.Request) {
	reqId, _ := r.Context().Value("reqId").(int)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeErrResponse(w, reqId, http.StatusInternalServerError, err.Error())
		return
	}
	request := types.VoteRequest{}
	if err := json.Unmarshal(body, &request); err != nil {
		writeErrResponse(w, reqId, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.engine.Vote(request)
	if err != nil {
		if _, ok := err.(*types.BadRequestError); ok {
			writeErrResponse(w, reqId, http.StatusBadRequest, err.Error())
			return
		}
		writeErrResponse(w, reqId, http.StatusInternalServerError, err.Error())
		return
	}
	writeResponse(w, reqId, response)
}

// @Tags Translation
// @Id getConfirmedTranslation
// @Summary Get confirmed translation
// @Param word path integer true "word id"
// @Param language path string true "language"
// @Success 200 {object} types.GetConfirmedTranslationResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /word/{word}/language/{language}/confirmed-translation [get]
func (s *Server) confirmedTranslation(w http.ResponseWriter, r *http.Request) {
	reqId, _ := r.Context().Value("reqId").(int)
	vars := mux.Vars(r)
	wordId, err := toUint(vars, "word")
	if err != nil {
		writeErrResponse(w, reqId, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.engine.GetConfirmedTranslation(uint32(wordId), mux.Vars(r)["language"])
	if err != nil {
		writeErrResponse(w, reqId, http.StatusInternalServerError, err.Error())
		return
	}
	writeResponse(w, reqId, response)
}
