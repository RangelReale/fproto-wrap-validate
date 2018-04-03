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

type RepeatedType int

const (
	RT_ARRAY RepeatedType = iota
	RT_MAP
)

type Validator interface {
	FPValidator() // tag interface
}

type ValidatorNormal interface {
	GenerateValidation(g *fproto_gowrap.GeneratorFile, vh ValidatorHelper, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) error
}

type ValidatorRepeated interface {
	GenerateValidationRepeated(g *fproto_gowrap.GeneratorFile, vh ValidatorHelper, repeatedType RepeatedType, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) error
}
