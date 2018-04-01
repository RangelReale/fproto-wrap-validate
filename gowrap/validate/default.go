package fproto_gowrap_validate_default

import (
	"fmt"
	"strings"

	"github.com/RangelReale/fdep"
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap-validate/gowrap"
	"github.com/RangelReale/fproto-wrap/gowrap"
)

type TypeValidatorPlugin_Default struct {
	DefaultTypeValidators []DefaultTypeValidatorPlugin
}

func (tp *TypeValidatorPlugin_Default) GetTypeValidator(validatorType *fdep.OptionType) fproto_gowrap_validate.TypeValidator {
	// validate.field
	if validatorType.Option != nil &&
		validatorType.Option.DepFile.FilePath == "github.com/RangelReale/fproto-wrap-validate/validate.proto" &&
		validatorType.Option.DepFile.ProtoFile.PackageName == "validate" &&
		validatorType.Name == "field" {
		return &TypeValidator_Default{DefaultTypeValidators: tp.DefaultTypeValidators}
	}
	return nil
}

func (tp *TypeValidatorPlugin_Default) ValidatorPrefixes() []string {
	return []string{"validate"}
}

type TypeValidator_Default struct {
	DefaultTypeValidators []DefaultTypeValidatorPlugin
}

func (t *TypeValidator_Default) GenerateValidation(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) error {
	tinfo := g.G().GetTypeInfo(tp)

	if tinfo.Converter().TCID() == fproto_gowrap.TCID_SCALAR {
		return t.generateValidation_scalar(g, tp, tinfo, option, varSrc, varError)
	}

	var tv DefaultTypeValidator
	for _, tvp := range t.DefaultTypeValidators {
		if tv = tvp.GetDefaultTypeValidator(tp); tv != nil {
			break
		}
	}

	if tv != nil {
		return tv.GenerateValidation(g, tp, option, varSrc, varError)
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
