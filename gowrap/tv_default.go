package fproto_gowrap_validate

import (
	"fmt"
	"strings"

	"github.com/RangelReale/fdep"
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap/gowrap"
	"github.com/RangelReale/fproto-wrap/gowrap/tc/uuid"
)

type TypeValidatorPlugin_Default struct {
}

func (tp *TypeValidatorPlugin_Default) GetTypeValidator(validatorType *fdep.OptionType) TypeValidator {
	// validate.field
	if validatorType.Option != nil &&
		validatorType.Option.DepFile.FilePath == "github.com/RangelReale/fproto-wrap-validate/validate.proto" &&
		validatorType.Option.DepFile.ProtoFile.PackageName == "validate" &&
		validatorType.Name == "field" {
		return &TypeValidator_Default{}
	}
	return nil
}

func (tp *TypeValidatorPlugin_Default) ValidatorPrefixes() []string {
	return []string{"validate"}
}

type TypeValidator_Default struct {
}

func (t *TypeValidator_Default) GenerateValidation(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) error {
	tinfo := g.G().GetTypeInfo(tp)

	if tinfo.Converter().TCID() == fproto_gowrap.TCID_SCALAR {
		return t.generateValidation_scalar(g, tp, tinfo, option, varSrc, varError)
	}

	if tinfo.Converter().TCID() == fproto_gowrap_uuid.TCID_UUID || tinfo.Converter().TCID() == fproto_gowrap_uuid.TCID_NULLUUID {
		return t.generateValidation_uuid(g, tp, tinfo, option, varSrc, varError)
	}

	return fmt.Errorf("Unknown type for validator: %s", tp.FullOriginalName())
}

func (t *TypeValidator_Default) generateValidation_scalar(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, tinfo fproto_gowrap.TypeInfo, option *fproto.OptionElement, varSrc string, varError string) error {
	errors_alias := g.DeclDep("errors", "errors")

	var opag []string
	for agn, agv := range option.AggregatedValues {
		opag = append(opag, fmt.Sprintf("%s=%s", agn, agv.Source))
	}

	g.P("// ", option.Name, " -- ", option.ParenthesizedName, " ** ", option.NPName, " @@ ", option.Value.Source, " %% ", strings.Join(opag, ", "))

	for agn, agv := range option.AggregatedValues {
		supported := false

		switch *tp.ScalarType {
		//
		// INTEGER
		//
		case fproto.Fixed32Scalar, fproto.Fixed64Scalar, fproto.Int32Scalar, fproto.Int64Scalar,
			fproto.Sfixed32Scalar, fproto.Sfixed64Scalar, fproto.Sint32Scalar, fproto.Sint64Scalar,
			fproto.Uint32Scalar, fproto.Uint64Scalar:
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
		//
		// FLOAT
		//
		case fproto.DoubleScalar, fproto.FloatScalar:
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
		//
		// STRING
		//
		case fproto.StringScalar:
			//
			// xrequired
			//
			if agn == "xrequired" {
				supported = true
				if agv.Source == "true" {
					g.P("if ", varSrc, " == \"\" {")
					g.In()
					g.P("err = ", errors_alias, ".New(\"Cannot be blank\")")
					g.Out()
					g.P("}")
					g.GenerateSimpleErrorCheck()
				}
			} else if agn == "length_eq" {
				supported = true
				g.P("if len(", varSrc, ") != ", agv.Source, " {")
				g.In()
				g.P("err = ", errors_alias, ".New(\"Length must be ", agv.Source, "\")")
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

func (t *TypeValidator_Default) generateValidation_uuid(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, tinfo fproto_gowrap.TypeInfo, option *fproto.OptionElement, varSrc string, varError string) error {
	uuid_alias := g.DeclDep("github.com/RangelReale/go.uuid", "uuid")
	errors_alias := g.DeclDep("errors", "errors")

	for agn, agv := range option.AggregatedValues {
		supported := false

		switch tinfo.Converter().TCID() {
		case fproto_gowrap_uuid.TCID_UUID:
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

		case fproto_gowrap_uuid.TCID_NULLUUID:
			//
			// xrequired
			//
			if agn == "xrequired" {
				supported = true
				if agv.Source == "true" {
					g.P("if !", varSrc, ".Valid && ", uuid_alias, ".Equals(", varSrc, ".UUID, uuid.Nil) {")
					g.In()
					g.P("err = ", errors_alias, ".New(\"Cannot be blank\")")
					g.Out()
					g.P("}")
					g.GenerateSimpleErrorCheck()
				}
			}
		}

		if !supported {
			return fmt.Errorf("Validation %s not supported for type %s", agn, tp.FullOriginalName())
		}
	}
	return nil
}
