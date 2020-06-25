package db

import (
	"github.com/idena-network/idena-translation/types"
	"time"
)

type Accessor interface {
	SubmitTranslation(address string, wordId uint32, language string, name string, description string, timestamp time.Time, confirmedRate uint8) (*string, error)
	GetTranslations(wordId uint32, language string, continuationToken string, limit uint8, confirmedRate uint8) ([]types.Translation, string, error)
	Vote(address string, translationId string, up bool, timestamp time.Time) (int, int, error)
	GetConfirmedTranslation(wordId uint32, language string, confirmedRate uint8) (*types.Translation, error)
}
