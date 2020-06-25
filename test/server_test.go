package test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/idena-network/idena-translation/config"
	"github.com/idena-network/idena-translation/core"
	"github.com/idena-network/idena-translation/db"
	"github.com/idena-network/idena-translation/db/postgres"
	"github.com/idena-network/idena-translation/server"
	"github.com/idena-network/idena-translation/test/client"
	"github.com/idena-network/idena-translation/test/client/translation"
	"github.com/idena-network/idena-translation/test/models"
	"github.com/idena-network/idena-translation/types"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

const (
	port    = 10080
	connStr = "postgres://postgres@localhost?sslmode=disable"
	schema  = "translation_auto_test"
)

func Test_submitTranslation(t *testing.T) {
	s, dbAccessor, cl, nodeClient := startTestServer()
	defer s.Stop()

	address1 := "address1"
	nodeClient.IdentitiesByAddr[address1] = true

	// When
	invalidName := ""
	res, err := cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: invalidName, Description: "description", Timestamp: "2020-01-01T10:00:00+01:00",
		}, address1, nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.NotNil(t, err)
	require.IsType(t, &translation.SubmitTranslationBadRequest{}, err)

	// When
	invalidDescription := strings.Repeat("s", 151)
	res, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "name", Description: invalidDescription, Timestamp: "2020-01-01T10:00:00+01:00",
		}, address1, nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.NotNil(t, err)
	require.IsType(t, &translation.SubmitTranslationBadRequest{}, err)

	// When
	res, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "name", Description: "description", Timestamp: "2020-01-01T10:00:00+01:00",
		}, "anotherAddress", nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.NotIdentityError.Code()), res.GetPayload().ResCode)

	// When
	wrongSignature := "wrongSignature"
	res, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: &models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "name", Description: "description", Timestamp: "2020-01-01T10:00:00+01:00", Signature: wrongSignature,
		},
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.NotIdentityError.Code()), res.GetPayload().ResCode)

	// When
	res, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "name", Description: "description", Timestamp: "2020-01-01T10:00:00+01:00",
		}, address1, nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.SuccessResCode), res.GetPayload().ResCode)
	require.Empty(t, res.GetPayload().Error)

	// When
	res, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "new name", Description: "description", Timestamp: "2020-01-01T10:00:00+01:00",
		}, address1, nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.OutdatedSubmissionError.Code()), res.GetPayload().ResCode)

	// When
	upVotes, downVotes, err := dbAccessor.Vote("address2", "1", true, time.Now())
	require.Nil(t, err)
	require.Equal(t, 1, upVotes)
	require.Equal(t, 0, downVotes)
	res, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "new name", Description: "description", Timestamp: "2020-01-01T09:30:00Z",
		}, address1, nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.SuccessResCode), res.GetPayload().ResCode)
	require.Empty(t, res.GetPayload().Error)

	// When
	upVotes, downVotes, err = dbAccessor.Vote("address2", "2", true, time.Now())
	require.Nil(t, err)
	require.Equal(t, 1, upVotes)
	require.Equal(t, 0, downVotes)
	upVotes, downVotes, err = dbAccessor.Vote("address3", "2", true, time.Now())
	require.Nil(t, err)
	require.Equal(t, 2, upVotes)
	require.Equal(t, 0, downVotes)
	upVotes, downVotes, err = dbAccessor.Vote("address4", "2", true, time.Now())
	require.Nil(t, err)
	require.Equal(t, 3, upVotes)
	require.Equal(t, 0, downVotes)
	res, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "new name 2", Description: "description", Timestamp: "2020-01-01T10:00:00+01:00",
		}, address1, nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.ConfirmedTranslationExistsError.Code()), res.GetPayload().ResCode)
}

func Test_vote(t *testing.T) {
	s, dbAccessor, cl, nodeClient := startTestServer()
	defer s.Stop()

	translationId, err := dbAccessor.SubmitTranslation("translationAuthorAddress", 1, "id", "name", "description", time.Now(), 3)
	require.Nil(t, err)
	require.Equal(t, "1", *translationId)
	nodeClient.IdentitiesByAddr["translationAuthorAddress"] = true
	nodeClient.IdentitiesByAddr["address2"] = true

	// When
	res, err := cl.Translation.Vote(&translation.VoteParams{
		Vote: signedVoteRequest(&models.VoteRequest{
			TranslationID: "1", Up: true, Timestamp: "2020-01-01T10:00:00Z",
		}, "translationAuthorAddress", nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.SelfVotingError.Code()), res.GetPayload().ResCode)

	// When
	res, err = cl.Translation.Vote(&translation.VoteParams{
		Vote: signedVoteRequest(&models.VoteRequest{
			TranslationID: "1", Up: true, Timestamp: "2020-01-01T10:00:00Z",
		}, "notIdentityAddress", nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.NotIdentityError.Code()), res.GetPayload().ResCode)

	// When
	res, err = cl.Translation.Vote(&translation.VoteParams{
		Vote: signedVoteRequest(&models.VoteRequest{
			TranslationID: "2", Up: true, Timestamp: "2020-01-01T10:00:00Z",
		}, "address2", nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.NotNil(t, err)
	require.IsType(t, &translation.VoteBadRequest{}, err)

	// Vote up
	// When
	res, err = cl.Translation.Vote(&translation.VoteParams{
		Vote: signedVoteRequest(&models.VoteRequest{
			TranslationID: "1", Up: true, Timestamp: "2020-01-01T10:00:00Z",
		}, "address2", nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.SuccessResCode), res.GetPayload().ResCode)
	require.Equal(t, int64(1), res.GetPayload().UpVotes)
	require.Equal(t, int64(0), res.GetPayload().DownVotes)
	require.Empty(t, res.GetPayload().Error)

	// Vote up again
	// When
	res, err = cl.Translation.Vote(&translation.VoteParams{
		Vote: signedVoteRequest(&models.VoteRequest{
			TranslationID: "1", Up: true, Timestamp: "2020-01-01T11:00:00Z",
		}, "address2", nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.DuplicatedVoteError.Code()), res.GetPayload().ResCode)

	// Vote down with same timestamp
	// When
	res, err = cl.Translation.Vote(&translation.VoteParams{
		Vote: signedVoteRequest(&models.VoteRequest{
			TranslationID: "1", Up: false, Timestamp: "2020-01-01T10:00:00Z",
		}, "address2", nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.OutdatedSubmissionError.Code()), res.GetPayload().ResCode)

	// Vote down
	// When
	res, err = cl.Translation.Vote(&translation.VoteParams{
		Vote: signedVoteRequest(&models.VoteRequest{
			TranslationID: "1", Up: false, Timestamp: "2020-01-01T11:00:00Z",
		}, "address2", nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	// Then
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(types.SuccessResCode), res.GetPayload().ResCode)
	require.Equal(t, int64(0), res.GetPayload().UpVotes)
	require.Equal(t, int64(1), res.GetPayload().DownVotes)
	require.Empty(t, res.GetPayload().Error)
	translations, _, err := dbAccessor.GetTranslations(1, "id", "", 1, 1)
	require.Nil(t, err)
	require.Zero(t, translations[0].UpVotes)
	require.Equal(t, 1, translations[0].DownVotes)
}

func Test_fullFlow(t *testing.T) {
	s, _, cl, nodeClient := startTestServer()
	defer s.Stop()
	addresses := []string{
		"address0",
		"address1",
		"address2",
		"address3",
		"address4",
		"address5",
	}
	for _, address := range addresses {
		nodeClient.IdentitiesByAddr[address] = true
	}

	submitRes, err := cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "name0", Description: "description0", Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[0], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, submitRes)
	require.Equal(t, int64(types.SuccessResCode), submitRes.GetPayload().ResCode)

	_, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "name1", Description: "description1", Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[1], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, submitRes)
	require.Equal(t, int64(types.SuccessResCode), submitRes.GetPayload().ResCode)

	_, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "name2", Description: "description2", Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[2], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, submitRes)
	require.Equal(t, int64(types.SuccessResCode), submitRes.GetPayload().ResCode)

	_, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "name3", Description: "description3", Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[3], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, submitRes)
	require.Equal(t, int64(types.SuccessResCode), submitRes.GetPayload().ResCode)

	submitRes, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "name4", Description: "description4", Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[4], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, submitRes)
	require.Equal(t, int64(types.SuccessResCode), submitRes.GetPayload().ResCode)

	submitRes, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "id", Name: "name5", Description: "description5", Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[5], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, submitRes)
	require.Equal(t, int64(types.SuccessResCode), submitRes.GetPayload().ResCode)

	submitRes, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 2, Language: "id", Name: "name0", Description: "description0", Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[0], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, submitRes)
	require.Equal(t, int64(types.SuccessResCode), submitRes.GetPayload().ResCode)

	submitRes, err = cl.Translation.SubmitTranslation(&translation.SubmitTranslationParams{
		Translation: signedSubmitTransactionRequest(&models.SubmitTranslationRequest{
			Word: 1, Language: "fr", Name: "name0", Description: "description0", Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[0], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, submitRes)
	require.Equal(t, int64(types.SuccessResCode), submitRes.GetPayload().ResCode)

	listRes, err := cl.Translation.GetTranslations(&translation.GetTranslationsParams{
		Word: 1, Language: "id", Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, listRes)
	require.NotEmpty(t, listRes.ContinuationToken)
	require.Equal(t, 5, len(listRes.GetPayload().Translations))
	require.Equal(t, "1", listRes.GetPayload().Translations[0].ID)
	require.Equal(t, "5", listRes.GetPayload().Translations[4].ID)

	listRes, err = cl.Translation.GetTranslations(&translation.GetTranslationsParams{
		Word: 1, Language: "id", ContinuationToken: &listRes.ContinuationToken, Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, listRes)
	require.Empty(t, listRes.ContinuationToken)
	require.Equal(t, 1, len(listRes.GetPayload().Translations))
	require.Equal(t, "6", listRes.GetPayload().Translations[0].ID)
	require.Equal(t, "name5", listRes.GetPayload().Translations[0].Name)
	require.Equal(t, "description5", listRes.GetPayload().Translations[0].Description)

	voteRes, err := cl.Translation.Vote(&translation.VoteParams{
		Vote: signedVoteRequest(&models.VoteRequest{
			TranslationID: "6", Up: true, Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[0], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, voteRes)
	require.Equal(t, int64(types.SuccessResCode), voteRes.GetPayload().ResCode)

	voteRes, err = cl.Translation.Vote(&translation.VoteParams{
		Vote: signedVoteRequest(&models.VoteRequest{
			TranslationID: "1", Up: false, Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[5], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, voteRes)
	require.Equal(t, int64(types.SuccessResCode), voteRes.GetPayload().ResCode)

	listRes, err = cl.Translation.GetTranslations(&translation.GetTranslationsParams{
		Word: 1, Language: "id", Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, listRes)
	require.NotEmpty(t, listRes.ContinuationToken)
	require.Equal(t, 5, len(listRes.GetPayload().Translations))
	require.Equal(t, "6", listRes.GetPayload().Translations[0].ID)
	require.False(t, listRes.GetPayload().Translations[0].Confirmed)
	require.Equal(t, "5", listRes.GetPayload().Translations[4].ID)

	listRes, err = cl.Translation.GetTranslations(&translation.GetTranslationsParams{
		Word: 1, Language: "id", ContinuationToken: &listRes.ContinuationToken, Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, listRes)
	require.Empty(t, listRes.ContinuationToken)
	require.Equal(t, 1, len(listRes.GetPayload().Translations))
	require.Equal(t, "1", listRes.GetPayload().Translations[0].ID)

	confirmedRes, err := cl.Translation.GetConfirmedTranslation(&translation.GetConfirmedTranslationParams{
		Word: 1, Language: "id", Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, confirmedRes.GetPayload())
	require.Nil(t, confirmedRes.GetPayload().Translation)

	voteRes, err = cl.Translation.Vote(&translation.VoteParams{
		Vote: signedVoteRequest(&models.VoteRequest{
			TranslationID: "6", Up: true, Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[1], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, voteRes)
	require.Equal(t, int64(types.SuccessResCode), voteRes.GetPayload().ResCode)

	voteRes, err = cl.Translation.Vote(&translation.VoteParams{
		Vote: signedVoteRequest(&models.VoteRequest{
			TranslationID: "6", Up: true, Timestamp: "2020-01-01T01:00:00Z",
		}, addresses[2], nodeClient.AddressesByValueAndSignature),
		Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, voteRes)
	require.Equal(t, int64(types.SuccessResCode), voteRes.GetPayload().ResCode)

	confirmedRes, err = cl.Translation.GetConfirmedTranslation(&translation.GetConfirmedTranslationParams{
		Word: 1, Language: "id", Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, confirmedRes.GetPayload())
	require.NotNil(t, confirmedRes.GetPayload().Translation)
	require.NotNil(t, "name5", confirmedRes.GetPayload().Translation.Name)
	require.NotNil(t, "description5", confirmedRes.GetPayload().Translation.Description)
	require.True(t, confirmedRes.GetPayload().Translation.Confirmed)

	listRes, err = cl.Translation.GetTranslations(&translation.GetTranslationsParams{
		Word: 1, Language: "id", Context: context.Background(),
	})
	require.Nil(t, err)
	require.NotNil(t, listRes)
	require.True(t, listRes.GetPayload().Translations[0].Confirmed)
}

func signedSubmitTransactionRequest(
	r *models.SubmitTranslationRequest,
	address string,
	addressesByValueAndSignature map[string]string,
) *models.SubmitTranslationRequest {
	val := strings.Join([]string{fmt.Sprint(r.Word), r.Language, r.Name, r.Description, r.Timestamp}, "")
	signature := val
	addressesByValueAndSignature[val+signature] = address
	r.Signature = signature
	return r
}

func signedVoteRequest(
	r *models.VoteRequest,
	address string,
	addressesByValueAndSignature map[string]string,
) *models.VoteRequest {
	val := strings.Join([]string{r.TranslationID, fmt.Sprint(r.Up), r.Timestamp}, "")
	signature := val
	addressesByValueAndSignature[val+signature] = address
	r.Signature = signature
	return r
}

func startTestServer() (*server.Server, db.Accessor, *client.IdenaFlipWordsTranslation, *TestNodeClient) {
	dbConnector, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	_, err = dbConnector.Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE")
	if err != nil {
		panic(err)
	}
	_, err = dbConnector.Exec("CREATE SCHEMA " + schema)
	if err != nil {
		panic(err)
	}
	dbAccessor := postgres.NewAccessor(connStr+"&search_path="+schema, "../resources")
	nodeClient := &TestNodeClient{
		IdentitiesByAddr:             make(map[string]bool),
		AddressesByValueAndSignature: make(map[string]string),
	}
	auth := core.NewEngine(dbAccessor, nodeClient, 5, 3)
	s := server.NewServer(port, auth)
	go s.Start(config.SwaggerConfig{})
	clConfig := client.DefaultTransportConfig().WithHost(fmt.Sprintf("localhost:%v", port))
	cl := client.NewHTTPClientWithConfig(nil, clConfig)
	return s, dbAccessor, cl, nodeClient
}

type TestNodeClient struct {
	IdentitiesByAddr             map[string]bool
	AddressesByValueAndSignature map[string]string
}

func (t *TestNodeClient) GetSignatureAddress(value, signature string) (string, error) {
	return t.AddressesByValueAndSignature[value+signature], nil
}

func (t *TestNodeClient) IsIdentity(address string) (bool, error) {
	return t.IdentitiesByAddr[address], nil
}
