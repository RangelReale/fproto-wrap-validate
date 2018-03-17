package fproto_gowrap_validate

import (
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap/gowrap"
	"github.com/RangelReale/fproto/fdep"
)

type TypeValidatorPlugin interface {
	// Returns a type validator for the type
	GetTypeConverter(validatorType *fdep.OptionType) TypeValidator

	ValidatorPrefixes() []string
}

type TypeValidator interface {
	GenerateValidation(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) (checkError bool, err error)
}
