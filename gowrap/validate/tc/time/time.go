package fproto_gowrap_validate_default_time

import (
	"fmt"

	"github.com/RangelReale/fdep"
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap-validate/gowrap/validate"
	"github.com/RangelReale/fproto-wrap/gowrap"
	"github.com/RangelReale/fproto-wrap/gowrap/tc/time"
)

//
// UUID
// Validates google.protobuf.Timestamp as time.Time
//

type DefaultTypeValidatorPlugin_Time struct {
}

func (t *DefaultTypeValidatorPlugin_Time) GetDefaultTypeValidator(typeinfo fproto_gowrap.TypeInfo, tp *fdep.DepType) fproto_gowrap_validate_default.DefaultTypeValidator {
	if typeinfo.Converter().TCID() == fproto_gowrap_time.TCID_TIME {
		return &DefaultTypeValidator_Time{}
	}
	if typeinfo.Converter().TCID() == fproto_gowrap_time.TCID_NULLTIME {
		return &DefaultTypeValidator_NullTime{}
	}

	return nil
}

//
// Time
//
type DefaultTypeValidator_Time struct {
}

func (v *DefaultTypeValidator_Time) GenerateValidation(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) error {
	errors_alias := g.DeclDep("errors", "errors")

	for agn, agv := range option.AggregatedValues {
		supported := false

		//
		// xrequired
		//
		if agn == "xrequired" {
			supported = true
			if agv.Source == "true" {
				g.P("if ", varSrc, ".IsZero() {")
				g.In()
				g.P("err = ", errors_alias, ".New(\"Cannot be blank\")")
				g.Out()
				g.P("}")
				g.GenerateSimpleErrorCheck()
			}
		}

		if !supported {
			return fmt.Errorf("Validation %s not supported for type %s", agn, tp.FullOriginalName())
		}
	}

	return nil
}

//
// NullTime
//
type DefaultTypeValidator_NullTime struct {
}

func (v *DefaultTypeValidator_NullTime) GenerateValidation(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) error {
	errors_alias := g.DeclDep("errors", "errors")

	for agn, agv := range option.AggregatedValues {
		supported := false

		//
		// xrequired
		//
		if agn == "xrequired" {
			supported = true
			if agv.Source == "true" {
				g.P("if !", varSrc, ".Valid || ", varSrc, ".IsZero() {")
				g.In()
				g.P("err = ", errors_alias, ".New(\"Cannot be blank\")")
				g.Out()
				g.P("}")
				g.GenerateSimpleErrorCheck()
			}
		}

		if !supported {
			return fmt.Errorf("Validation %s not supported for type %s", agn, tp.FullOriginalName())
		}
	}

	return nil
}
