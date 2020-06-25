package types

import (
	"github.com/pkg/errors"
	"time"
)

func (r SubmitTranslationRequest) Validate() error {
	if r.Word > 3299 {
		return errors.New("invalid value 'word'")
	}
	if len(r.Name) == 0 || len(r.Name) > 30 {
		return errors.New("invalid value 'name'")
	}
	if len(r.Description) > 150 {
		return errors.New("invalid value 'description'")
	}
	var timestamp time.Time
	if err := timestamp.UnmarshalText([]byte(r.Timestamp)); err != nil {
		return errors.New("invalid value 'timestamp'")
	}
	return nil
}

func (r VoteRequest) Validate() error {
	var timestamp time.Time
	if err := timestamp.UnmarshalText([]byte(r.Timestamp)); err != nil {
		return errors.New("invalid value 'timestamp'")
	}
	return nil
}
