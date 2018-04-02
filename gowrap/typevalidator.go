package fproto_gowrap_validator

import (
	"github.com/RangelReale/fdep"
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap/gowrap"
)

type TypeValidatorPlugin interface {
	// Returns a type validator for the type
	GetTypeValidator(validatorType *fdep.OptionType) TypeValidator

	ValidatorPrefixes() []string
}

type TypeValidator interface {
	GenerateValidation(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) error
}
