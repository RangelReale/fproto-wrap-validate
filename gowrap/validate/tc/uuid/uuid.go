package fproto_gowrap_validate_default_uuid

import (
	"fmt"

	"github.com/RangelReale/fdep"
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap-validate/gowrap/validate"
	"github.com/RangelReale/fproto-wrap/gowrap"
	"github.com/RangelReale/fproto-wrap/gowrap/tc/uuid"
)

//
// UUID
// Validates fproto_wrap.UUID
//

type DefaultTypeValidatorPlugin_UUID struct {
}

func (t *DefaultTypeValidatorPlugin_UUID) GetDefaultTypeValidator(typeinfo fproto_gowrap.TypeInfo, tp *fdep.DepType) fproto_gowrap_validate_default.DefaultTypeValidator {
	if typeinfo.Converter().TCID() == fproto_gowrap_uuid.TCID_UUID {
		return &DefaultTypeValidator_UUID{}
	}
	if typeinfo.Converter().TCID() == fproto_gowrap_uuid.TCID_NULLUUID {
		return &DefaultTypeValidator_NullUUID{}
	}

	return nil
}

//
// UUID
//
type DefaultTypeValidator_UUID struct {
}

func (v *DefaultTypeValidator_UUID) GenerateValidation(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) error {
	uuid_alias := g.DeclDep("github.com/RangelReale/go.uuid", "uuid")
	errors_alias := g.DeclDep("errors", "errors")

	for agn, agv := range option.AggregatedValues {
		supported := false

		//
		// xrequired
		//
		if agn == "xrequired" {
			supported = true
			if agv.Source == "true" {
				g.P("if ", uuid_alias, ".Equal(", varSrc, ", uuid.Nil) {")
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
// NullUUID
//
type DefaultTypeValidator_NullUUID struct {
}

func (v *DefaultTypeValidator_NullUUID) GenerateValidation(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) error {
	uuid_alias := g.DeclDep("github.com/RangelReale/go.uuid", "uuid")
	errors_alias := g.DeclDep("errors", "errors")

	for agn, agv := range option.AggregatedValues {
		supported := false

		//
		// xrequired
		//
		if agn == "xrequired" {
			supported = true
			if agv.Source == "true" {
				g.P("if !", varSrc, ".Valid || ", uuid_alias, ".Equal(", varSrc, ".UUID, uuid.Nil) {")
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
