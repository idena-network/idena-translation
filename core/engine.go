package core

import (
	"fmt"
	"github.com/idena-network/idena-translation/db"
	"github.com/idena-network/idena-translation/node"
	"github.com/idena-network/idena-translation/types"
	"strings"
	"time"
)

type Engine interface {
	SubmitTranslation(request types.SubmitTranslationRequest) (types.SubmitTranslationResponse, error)
	GetTranslations(wordId uint32, language string, continuationToken string) (types.GetTranslationsResponse, string, error)
	Vote(request types.VoteRequest) (types.VoteResponse, error)
	GetConfirmedTranslation(wordId uint32, language string) (types.GetConfirmedTranslationResponse, error)
}

func NewEngine(dbAccessor db.Accessor, nodeClient node.Client, itemsLimit, confirmedRate uint8) Engine {
	return &engineImpl{
		dbAccessor:    dbAccessor,
		nodeClient:    nodeClient,
		itemsLimit:    itemsLimit,
		confirmedRate: confirmedRate,
	}
}

type engineImpl struct {
	nodeClient    node.Client
	dbAccessor    db.Accessor
	itemsLimit    uint8
	confirmedRate uint8
}

func (engine *engineImpl) SubmitTranslation(request types.SubmitTranslationRequest) (types.SubmitTranslationResponse, error) {
	if err := request.Validate(); err != nil {
		return types.SubmitTranslationResponse{}, &types.BadRequestError{
			Message: err.Error(),
		}
	}
	address, err := engine.nodeClient.GetSignatureAddress(getTranslationSignedValue(request), request.Signature)
	if err != nil {
		return types.SubmitTranslationResponse{}, err
	}
	isIdentity, err := engine.nodeClient.IsIdentity(address)
	if err != nil {
		return types.SubmitTranslationResponse{}, err
	}
	if !isIdentity {
		return types.SubmitTranslationResponse{
			ResCode: types.NotIdentityError.Code(),
			Error:   types.NotIdentityError.Error(),
		}, nil
	}
	var translationId *string
	var timestamp time.Time
	_ = timestamp.UnmarshalText([]byte(request.Timestamp))
	if translationId, err = engine.dbAccessor.SubmitTranslation(
		address,
		request.Word,
		request.Language,
		request.Name,
		request.Description,
		timestamp,
		engine.confirmedRate,
	); err != nil {
		if translationError, ok := err.(*types.TranslationError); ok {
			return types.SubmitTranslationResponse{
				ResCode: translationError.Code(),
				Error:   translationError.Error(),
			}, nil
		}
		return types.SubmitTranslationResponse{}, err
	}
	return types.SubmitTranslationResponse{
		ResCode:       types.SuccessResCode,
		TranslationId: *translationId,
	}, nil
}

func getTranslationSignedValue(request types.SubmitTranslationRequest) string {
	return strings.Join([]string{fmt.Sprint(request.Word), request.Language, request.Name, request.Description, request.Timestamp}, "")
}

func (engine *engineImpl) GetTranslations(wordId uint32, language string, continuationToken string) (types.GetTranslationsResponse, string, error) {
	translations, nextContinuationToken, err := engine.dbAccessor.GetTranslations(wordId, language, continuationToken, engine.itemsLimit, engine.confirmedRate)
	if err != nil {
		return types.GetTranslationsResponse{}, "", err
	}
	return types.GetTranslationsResponse{
		Translations: translations,
	}, nextContinuationToken, nil
}

func (engine *engineImpl) Vote(request types.VoteRequest) (types.VoteResponse, error) {
	if err := request.Validate(); err != nil {
		return types.VoteResponse{}, &types.BadRequestError{
			Message: err.Error(),
		}
	}
	address, err := engine.nodeClient.GetSignatureAddress(getVoteSignedValue(request), request.Signature)
	if err != nil {
		return types.VoteResponse{}, err
	}
	isIdentity, err := engine.nodeClient.IsIdentity(address)
	if err != nil {
		return types.VoteResponse{}, err
	}
	if !isIdentity {
		return types.VoteResponse{
			ResCode: types.NotIdentityError.Code(),
			Error:   types.NotIdentityError.Error(),
		}, nil
	}
	var timestamp time.Time
	_ = timestamp.UnmarshalText([]byte(request.Timestamp))
	var upVotes, downVotes int
	if upVotes, downVotes, err = engine.dbAccessor.Vote(
		address,
		request.TranslationId,
		request.Up,
		timestamp,
	); err != nil {
		if translationError, ok := err.(*types.TranslationError); ok {
			return types.VoteResponse{
				ResCode: translationError.Code(),
				Error:   translationError.Error(),
			}, nil
		}
		return types.VoteResponse{}, err
	}
	return types.VoteResponse{
		ResCode:   types.SuccessResCode,
		UpVotes:   upVotes,
		DownVotes: downVotes,
	}, nil
}

func getVoteSignedValue(request types.VoteRequest) string {
	return strings.Join([]string{request.TranslationId, fmt.Sprint(request.Up), request.Timestamp}, "")
}

func (engine *engineImpl) GetConfirmedTranslation(wordId uint32, language string) (types.GetConfirmedTranslationResponse, error) {
	translation, err := engine.dbAccessor.GetConfirmedTranslation(wordId, language, engine.confirmedRate)
	if err != nil {
		return types.GetConfirmedTranslationResponse{}, err
	}
	return types.GetConfirmedTranslationResponse{
		Translation: translation,
	}, nil
}
