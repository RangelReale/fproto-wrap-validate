package fproto_gowrap_validate

import (
	"fmt"
	"strings"

	"github.com/RangelReale/fdep"
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap/gowrap"
)

type TypeValidatorPlugin_Default struct {
}

func (tp *TypeValidatorPlugin_Default) GetTypeValidator(validatorType *fdep.OptionType) TypeValidator {
	// validate.field
	if validatorType.Option != nil &&
		validatorType.Option.FileDep.FilePath == "github.com/RangelReale/fproto-wrap-validate/validate.proto" &&
		validatorType.Option.FileDep.ProtoFile.PackageName == "validate" &&
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

func (t *TypeValidator_Default) GenerateValidation(g *fproto_gowrap.GeneratorFile, tp *fdep.DepType, option *fproto.OptionElement, varSrc string, varError string) (checkError bool, err error) {
	errors_alias := g.Dep("errors", "errors")

	/*
		if !tp.IsScalar() {
			return false, fmt.Errorf("Validator expected scalar field, got %s", tp.FullOriginalName())
		}
	*/

	tinfo := g.G().GetTypeInfo(tp)

	if tinfo.Converter().TCID() != fproto_gowrap.TCID_SCALAR {
		return false, fmt.Errorf("Validator expected scalar field, got %s", tp.FullOriginalName())
	}

	var opag []string
	for agn, agv := range option.AggregatedValues {
		opag = append(opag, fmt.Sprintf("%s=%s", agn, agv.Source))
	}

	g.P("// ", option.Name, " -- ", option.ParenthesizedName, " @@ ", option.Value.Source, " %% ", strings.Join(opag, ", "))

	g.P("if ", varSrc, " == \"\" {")
	g.In()
	g.P("err = ", errors_alias, ".New(\"Cannot be blank\")")
	g.Out()
	g.P("}")

	return true, nil
}
