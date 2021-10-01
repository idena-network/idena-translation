package types

const AppVersion = "0.1.0"

type ErrorResponse struct {
	Error string `json:"error"`
} // @Name ErrorResponse

type SubmitTranslationRequest struct {
	Word        uint32 `json:"word",minimum:"0" maximum:"3939"`
	Language    string `json:"language" example:"en"`
	Name        string `json:"name",minLength:"1" maxLength:"30"`
	Description string `json:"description",minLength:"0" maxLength:"150"`
	Timestamp   string `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	Signature   string `json:"signature"`
} // @Name SubmitTranslationRequest

type SubmitTranslationResponse struct {
	ResCode       byte   `json:"resCode" enums:"0,1,2,4"`
	TranslationId string `json:"translationId,omitempty"`
	Error         string `json:"error,omitempty"`
} // @Name SubmitTranslationResponse

type GetTranslationsResponse struct {
	Translations []Translation `json:"translations"`
} // @Name GetTranslationsResponse

type Translation struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UpVotes     int    `json:"upVotes"`
	DownVotes   int    `json:"downVotes"`
	Confirmed   bool   `json:"confirmed"`
} // @Name Translation

type VoteRequest struct {
	TranslationId string `json:"translationId"`
	Up            bool   `json:"up"`
	Timestamp     string `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	Signature     string `json:"signature"`
} // @Name VoteRequest

type VoteResponse struct {
	ResCode   byte   `json:"resCode" enums:"0,3,4,5"`
	UpVotes   int    `json:"upVotes"`
	DownVotes int    `json:"downVotes"`
	Error     string `json:"error,omitempty"`
} // @Name VoteResponse

type GetConfirmedTranslationResponse struct {
	Translation *Translation `json:"translation"`
} // @Name GetConfirmedTranslationResponse
