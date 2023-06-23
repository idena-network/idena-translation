package types

import (
	"github.com/pkg/errors"
	"time"
)

func (r SubmitTranslationRequest) Validate() error {
	if r.Word > 4615 {
		return errors.New("Invalid value 'word'")
	}
	if name := []rune(r.Name); len(name) == 0 || len(name) > 30 {
		return errors.New("Translation exceeds the maximum length")
	}
	if description := []rune(r.Description); len(description) > 150 {
		return errors.New("Translation description exceeds the maximum length")
	}
	var timestamp time.Time
	if err := timestamp.UnmarshalText([]byte(r.Timestamp)); err != nil {
		return errors.New("Invalid value 'timestamp'")
	}
	return nil
}

func (r VoteRequest) Validate() error {
	var timestamp time.Time
	if err := timestamp.UnmarshalText([]byte(r.Timestamp)); err != nil {
		return errors.New("Invalid value 'timestamp'")
	}
	return nil
}
