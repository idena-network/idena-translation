package types

type ResCode = byte

const (
	SuccessResCode                    ResCode = 0
	notIdentityResCode                ResCode = 1
	confirmedTranslationExistsResCode ResCode = 2
	selfVotingResCode                 ResCode = 3
	outdatedSubmissionResCode         ResCode = 4
	duplicatedVoteResCode             ResCode = 5
)

var (
	NotIdentityError = &TranslationError{
		code:  notIdentityResCode,
		error: "sender is not validated",
	}
	ConfirmedTranslationExistsError = &TranslationError{
		code:  confirmedTranslationExistsResCode,
		error: "confirmed translation exists",
	}
	SelfVotingError = &TranslationError{
		code:  selfVotingResCode,
		error: "voting for own translation is not allowed",
	}
	OutdatedSubmissionError = &TranslationError{
		code:  outdatedSubmissionResCode,
		error: "outdated submission",
	}
	DuplicatedVoteError = &TranslationError{
		code:  duplicatedVoteResCode,
		error: "duplicated vote",
	}
)

type TranslationError struct {
	code  uint8
	error string
}

func (e *TranslationError) Error() string {
	return e.error
}

func (e *TranslationError) Code() uint8 {
	return e.code
}

type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return e.Message
}
