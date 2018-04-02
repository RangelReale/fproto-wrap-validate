package fproto_gowrap_validator

import (
	"github.com/RangelReale/fdep"
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap/gowrap"
)

type ValidatorPlugin interface {
	// Returns a type validator for the type
	GetValidator(validatorType *fdep.OptionType) Validator

	ValidatorPrefixes() []string
}

type Validator interface {
	GenerateValidation(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) error
}
