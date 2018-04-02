package fproto_gowrap_validator

import (
	"github.com/RangelReale/fdep"
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap/gowrap"
)

type ValidatorHelper interface {
	GenerateValidationErrorCheck(g *fproto_gowrap.Generator, validationItem string, errorId ValidationErrorId)

	// Gets a type validator
	GetTypeValidator(validatorType *fdep.OptionType, typeinfo fproto_gowrap.TypeInfo, tp *fdep.DepType) TypeValidator

	// Check if any field of this type has a known validation type
	TypeHasValidator(g *fproto_gowrap.Generator, element fproto.FProtoElement) (bool, error)

	// Get all validators for the field
	FieldGetValidators(g *fproto_gowrap.Generator, parentElement fproto.FProtoElement, field fproto.FieldElementTag) ([]*FieldValidator, error)

	// Check whether the field has any knwon validator
	FieldHasValidator(g *fproto_gowrap.Generator, parentElement fproto.FProtoElement, field fproto.FieldElementTag) (bool, error)

	// Check if the field type has any validator
	FieldTypeHasValidator(g *fproto_gowrap.Generator, parentElement fproto.FProtoElement, field fproto.FieldElementTag) (bool, error)

	// Gets the FIELD option validation
	OptionGetValidator(g *fproto_gowrap.Generator, opt *fproto.OptionElement) (Validator, error)

	// Checks if the option has a validator
	OptionHasValidator(g *fproto_gowrap.Generator, opt *fproto.OptionElement) (bool, error)

	// Finds a validator for an option
	FindValidatorForOption(optType *fdep.OptionType) Validator
}
