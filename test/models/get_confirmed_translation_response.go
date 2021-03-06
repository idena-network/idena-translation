// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// GetConfirmedTranslationResponse get confirmed translation response
//
// swagger:model GetConfirmedTranslationResponse
type GetConfirmedTranslationResponse struct {

	// translation
	Translation *Translation `json:"translation,omitempty"`
}

// Validate validates this get confirmed translation response
func (m *GetConfirmedTranslationResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateTranslation(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *GetConfirmedTranslationResponse) validateTranslation(formats strfmt.Registry) error {

	if swag.IsZero(m.Translation) { // not required
		return nil
	}

	if m.Translation != nil {
		if err := m.Translation.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("translation")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *GetConfirmedTranslationResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *GetConfirmedTranslationResponse) UnmarshalBinary(b []byte) error {
	var res GetConfirmedTranslationResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
