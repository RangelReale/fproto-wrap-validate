package fproto_gowrap_validate_default_duration

import (
	"fmt"

	"github.com/RangelReale/fdep"
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap-validate/gowrap/validate"
	"github.com/RangelReale/fproto-wrap/gowrap"
	"github.com/RangelReale/fproto-wrap/gowrap/tc/duration"
)

//
// Duration
// Validates google.protobuf.Duration as time.Duration
//

type DefaultTypeValidatorPlugin_Duration struct {
}

func (t *DefaultTypeValidatorPlugin_Duration) GetDefaultTypeValidator(typeinfo fproto_gowrap.TypeInfo, tp *fdep.DepType) fproto_gowrap_validate_default.DefaultTypeValidator {
	if typeinfo.Converter().TCID() == fproto_gowrap_duration.TCID_DURATION {
		return &DefaultTypeValidator_Duration{}
	}

	return nil
}

//
// Time
//
type DefaultTypeValidator_Duration struct {
}

func (v *DefaultTypeValidator_Duration) GenerateValidation(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) error {
	errors_alias := g.DeclDep("errors", "errors")

	for agn, agv := range option.AggregatedValues {
		supported := false

		//
		// xrequired
		//
		if agn == "xrequired" {
			supported = true
			if agv.Source == "true" {
				g.P("if ", varSrc, " == 0 {")
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
