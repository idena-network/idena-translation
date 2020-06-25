package postgres

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/idena-network/idena-translation/db"
	"github.com/idena-network/idena-translation/types"
	log "github.com/inconshreveable/log15"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	initQuery                    = "init.sql"
	submitTranslationQuery       = "submitTranslation.sql"
	getTranslationsQuery         = "getTranslations.sql"
	voteQuery                    = "vote.sql"
	getConfirmedTranslationQuery = "getConfirmedTranslation.sql"
)

type accessor struct {
	db      *sql.DB
	queries map[string]string
}

func NewAccessor(connStr string, scriptsDirPath string) db.Accessor {
	sqlDb, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	a := &accessor{
		db:      sqlDb,
		queries: readQueries(scriptsDirPath),
	}
	for {
		if err := a.init(); err != nil {
			log.Error(fmt.Sprintf("Unable to initialize postgres connection: %v", err))
			time.Sleep(time.Second * 10)
			continue
		}
		break
	}
	return a
}

func readQueries(scriptsDirPath string) map[string]string {
	files, err := ioutil.ReadDir(scriptsDirPath)
	if err != nil {
		panic(err)
	}
	queries := make(map[string]string)
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		bytes, err := ioutil.ReadFile(filepath.Join(scriptsDirPath, file.Name()))
		if err != nil {
			panic(err)
		}
		queryName := file.Name()
		query := string(bytes)
		queries[queryName] = query
		log.Debug(fmt.Sprintf("Read query %s from %s", queryName, scriptsDirPath))
	}
	return queries
}

func (a *accessor) init() error {
	if err := a.db.Ping(); err != nil {
		return err
	}
	if _, err := a.db.Exec(a.getQuery(initQuery)); err != nil {
		return err
	}
	return nil
}

func (a *accessor) getQuery(name string) string {
	if query, present := a.queries[name]; present {
		return query
	}
	panic(fmt.Sprintf("There is no query '%s'", name))
}

func (a *accessor) SubmitTranslation(address string, wordId uint32, language string, name string, description string, timestamp time.Time, confirmedRate uint8) (*string, error) {
	var resCode int
	var translationId string
	if err := a.db.QueryRow(a.getQuery(submitTranslationQuery),
		address, wordId, language, name, description, timestamp, confirmedRate).Scan(&resCode, &translationId); err != nil {
		return nil, err
	}
	switch resCode {
	case 0:
		return &translationId, nil
	case -1:
		return nil, &types.BadRequestError{
			Message: "invalid value 'language'",
		}
	case 2:
		return nil, types.ConfirmedTranslationExistsError
	case 3:
		return nil, types.OutdatedSubmissionError
	default:
		return nil, errors.New(fmt.Sprintf("unknown res code %d", resCode))
	}
}

func (a *accessor) GetTranslations(wordId uint32, language string, continuationToken string, limit uint8, confirmedRate uint8) ([]types.Translation, string, error) {
	id, rate, err := parseContinuationToken(continuationToken)
	if err != nil {
		return nil, "", err
	}
	rows, err := a.db.Query(a.getQuery(getTranslationsQuery), wordId, language, rate, id, limit+1, confirmedRate)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var res []types.Translation
	for rows.Next() {
		var item types.Translation
		err := rows.Scan(&item.Id, &item.Name, &item.Description, &item.UpVotes, &item.DownVotes, &item.Confirmed)
		if err != nil {
			return nil, "", err
		}
		res = append(res, item)
	}
	var nextContinuationToken string
	if len(res) > 0 && len(res) == int(limit+1) {
		nextItem := res[len(res)-1]
		id, _ := strconv.Atoi(nextItem.Id)
		nextContinuationToken = buildContinuationToken(id, nextItem.UpVotes-nextItem.DownVotes)
		res = res[:len(res)-1]
	}
	return res, nextContinuationToken, err
}

func buildContinuationToken(id int, rate int) string {
	val := strconv.Itoa(id) + "|" + strconv.Itoa(rate)
	return hex.EncodeToString([]byte(val))
}

var invalidContinuationToken = &types.BadRequestError{
	Message: "invalid value 'continuation-token'",
}

func parseContinuationToken(token string) (id, rate int, err error) {
	if len(token) == 0 {
		return 0, 0, nil
	}
	b, err := hex.DecodeString(token)
	if err != nil {
		return 0, 0, invalidContinuationToken
	}
	s := string(b)
	fields := strings.Split(s, "|")
	if len(fields) != 2 {
		return 0, 0, invalidContinuationToken
	}
	id, err = strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, invalidContinuationToken
	}
	rate, err = strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, invalidContinuationToken
	}
	return
}

func (a *accessor) Vote(address string, translationId string, up bool, timestamp time.Time) (int, int, error) {
	translationIdNum, err := strconv.Atoi(translationId)
	if err != nil {
		return 0, 0, &types.BadRequestError{
			Message: "invalid value 'translationId'",
		}
	}
	var resCode, upVotes, downVotes int
	if err := a.db.QueryRow(a.getQuery(voteQuery), address, translationIdNum, up, timestamp).Scan(&resCode, &upVotes, &downVotes); err != nil {
		return 0, 0, err
	}
	switch resCode {
	case 0:
		return upVotes, downVotes, nil
	case -1:
		return 0, 0, &types.BadRequestError{
			Message: "invalid value 'translationId'",
		}
	case 1:
		return 0, 0, types.SelfVotingError
	case 2:
		return 0, 0, types.OutdatedSubmissionError
	case 3:
		return 0, 0, types.DuplicatedVoteError
	default:
		return 0, 0, errors.New(fmt.Sprintf("unknown res code %d", resCode))
	}
}

func (a *accessor) GetConfirmedTranslation(wordId uint32, language string, confirmedRate uint8) (*types.Translation, error) {
	res := types.Translation{}
	err := a.db.QueryRow(a.getQuery(getConfirmedTranslationQuery), wordId, language, confirmedRate).
		Scan(&res.Id, &res.Name, &res.Description, &res.UpVotes, &res.DownVotes, &res.Confirmed)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}
