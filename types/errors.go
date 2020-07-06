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
		error: "Sender is not validated",
	}
	ConfirmedTranslationExistsError = &TranslationError{
		code:  confirmedTranslationExistsResCode,
		error: "Confirmed translation exists",
	}
	SelfVotingError = &TranslationError{
		code:  selfVotingResCode,
		error: "Voting for own translation is not allowed",
	}
	OutdatedSubmissionError = &TranslationError{
		code:  outdatedSubmissionResCode,
		error: "Outdated submission",
	}
	DuplicatedVoteError = &TranslationError{
		code:  duplicatedVoteResCode,
		error: "Duplicated vote",
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
