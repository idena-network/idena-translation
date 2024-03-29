// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// SubmitTranslationRequest submit translation request
//
// swagger:model SubmitTranslationRequest
type SubmitTranslationRequest struct {

	// description
	// Max Length: 100
	Description string `json:"description,omitempty"`

	// language
	Language string `json:"language,omitempty"`

	// name
	// Max Length: 20
	Name string `json:"name,omitempty"`

	// signature
	Signature string `json:"signature,omitempty"`

	// timestamp
	Timestamp string `json:"timestamp,omitempty"`

	// word
	// Maximum: 4615
	Word int64 `json:"word,omitempty"`
}

// Validate validates this submit translation request
func (m *SubmitTranslationRequest) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDescription(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateWord(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SubmitTranslationRequest) validateDescription(formats strfmt.Registry) error {

	if swag.IsZero(m.Description) { // not required
		return nil
	}

	if err := validate.MaxLength("description", "body", string(m.Description), 100); err != nil {
		return err
	}

	return nil
}

func (m *SubmitTranslationRequest) validateName(formats strfmt.Registry) error {

	if swag.IsZero(m.Name) { // not required
		return nil
	}

	if err := validate.MaxLength("name", "body", string(m.Name), 20); err != nil {
		return err
	}

	return nil
}

func (m *SubmitTranslationRequest) validateWord(formats strfmt.Registry) error {

	if swag.IsZero(m.Word) { // not required
		return nil
	}

	if err := validate.MaximumInt("word", "body", int64(m.Word), 4615, false); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *SubmitTranslationRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *SubmitTranslationRequest) UnmarshalBinary(b []byte) error {
	var res SubmitTranslationRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
