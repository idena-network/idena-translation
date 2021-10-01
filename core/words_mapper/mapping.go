package words_mapper

import (
	"encoding/json"
	"fmt"
	log "github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"time"
)

type WordsMapper interface {
	GetInitialWordId(wordId uint32) uint32
}

func NewWordsMapper(wordsUrl string) WordsMapper {
	initialWordIdsByWordId, err := initInitialWordIdsByWordId(wordsUrl)
	if err != nil {
		panic(err)
	}
	log.Info("Words mapper initialized", "size", len(initialWordIdsByWordId))
	return &wordsMapperImpl{
		initialWordIdsByWordId: initialWordIdsByWordId,
	}
}

type wordsMapperImpl struct {
	initialWordIdsByWordId map[uint32]uint32
}

func (wordsMapper *wordsMapperImpl) GetInitialWordId(wordId uint32) uint32 {
	if i, ok := wordsMapper.initialWordIdsByWordId[wordId]; ok {
		return i
	}
	return wordId
}

type words struct {
	Words []word `json:"words"`
}

type word struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func (w word) key() string {
	data, _ := json.Marshal(w)
	return string(data)
}

func initInitialWordIdsByWordId(wordsUrl string) (map[uint32]uint32, error) {
	if len(wordsUrl) == 0 {
		return nil, nil
	}
	wordsBytes, err := sendRequest(wordsUrl)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load words")
	}
	wordsData := words{}
	if err := json.Unmarshal(wordsBytes, &wordsData); err != nil {
		return nil, errors.Wrap(err, "unable to deserialize words")
	}
	res := make(map[uint32]uint32, len(wordsData.Words))
	firstIndexes := make(map[string]uint32)
	for i, word := range wordsData.Words {
		key := word.key()
		if firstIndex, ok := firstIndexes[key]; ok {
			res[uint32(i)] = firstIndex
		} else {
			firstIndexes[key] = uint32(i)
		}
	}
	return res, nil
}

func sendRequest(req string) ([]byte, error) {
	httpReq, err := http.NewRequest("GET", req, nil)
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	defer func() {
		if resp == nil || resp.Body == nil {
			return
		}
		resp.Body.Close()
	}()
	httpClient := &http.Client{
		Timeout: time.Second * 5,
	}
	resp, err = httpClient.Do(httpReq)
	if err == nil && resp.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("resp code %v", resp.StatusCode))
	}
	if err != nil {
		return nil, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read resp")
	}
	return respBody, nil
}
