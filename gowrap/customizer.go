package fproto_gowrap_validator

import (
	"strconv"
	"strings"

	"github.com/RangelReale/fdep"
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap/gowrap"
)

// Adds a json tag to all struct fields, using snake case formatting
type Customizer_Validator struct {
	FileId         string
	Validators     []ValidatorPlugin
	TypeValidators []TypeValidatorPlugin
	GenAllElements bool // whether to generate validation for all elements, even if they don't have validation options.
}

func NewCustomizer_Validator() *Customizer_Validator {
	return &Customizer_Validator{
		FileId: fproto_gowrap.FILEID_MAIN,
	}
}

// Gets a validator for an option type
func (c *Customizer_Validator) GetValidator(optionType *fdep.OptionType) Validator {
	for _, tcp := range c.Validators {
		tc := tcp.GetValidator(optionType)
		if tc != nil {
			return tc
		}
	}
	return nil
}

// Gets a type validator
func (c *Customizer_Validator) GetTypeValidator(validatorType *fdep.OptionType, typeinfo fproto_gowrap.TypeInfo, tp *fdep.DepType) TypeValidator {
	for _, tcp := range c.TypeValidators {
		tc := tcp.GetTypeValidator(validatorType, typeinfo, tp)
		if tc != nil {
			return tc
		}
	}
	return nil
}

// Generate code after message definitions
func (c *Customizer_Validator) GenerateCode(g *fproto_gowrap.Generator) error {
	var validate_elements []fproto.FProtoElement

	if g.GetDepFile().ProtoFile != nil {
		for _, msg := range g.GetDepFile().ProtoFile.CollectMessages() {
			fhas, err := c.TypeHasValidator(g, msg)
			if err != nil {
				return err
			}

			if c.GenAllElements || fhas {
				validate_elements = append(validate_elements, msg)
			}
		}

		for _, oofield := range g.GetDepFile().ProtoFile.CollectFields() {
			if oof, isoof := oofield.(*fproto.OneOfFieldElement); isoof {
				fhas, err := c.TypeHasValidator(g, oof)
				if err != nil {
					return err
				}

				if c.GenAllElements || fhas {
					validate_elements = append(validate_elements, oof)
				}
			}
		}
	}

	if len(validate_elements) > 0 {
		g.F(c.FileId).P("//")
		g.F(c.FileId).P("// VALIDATION")
		g.F(c.FileId).P("//")
		g.F(c.FileId).P()

		for _, ve := range validate_elements {
			err := c.generateValidationForElement(g, ve)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Generate validation for fproto element
func (c *Customizer_Validator) generateValidationForElement(g *fproto_gowrap.Generator, element fproto.FProtoElement) error {
	switch el := element.(type) {
	case *fproto.MessageElement:
		return c.generateValidationForMessageOrOneOf(g, el)
	case *fproto.OneOfFieldElement:
		return c.generateValidationForOneOf(g, el)
	}
	return nil
}

// Generate validation for message or oneof
func (c *Customizer_Validator) generateValidationForMessageOrOneOf(g *fproto_gowrap.Generator, element fproto.FProtoElement) error {
	vruntime_alias := g.F(c.FileId).DeclDep("github.com/RangelReale/fproto-wrap-validator/gowrap/runtime", "validator_runtime")

	var eleGoName string
	var fields []fproto.FieldElementTag

	switch el := element.(type) {
	case *fproto.MessageElement:
		eleGoName, _ = g.BuildMessageName(el)
		fields = el.Fields
	case fproto.FieldElementTag:
		// assume parent is oneof field
		eleGoName, _ = g.BuildOneOfFieldName(el)

		// set the field as itself
		fields = append(fields, el)
	}

	tpMsg := fdep.NewDepTypeFromElement(g.GetDepFile(), element)

	// func (m* MyElement) Validate() err
	g.F(c.FileId).P("func (m *", eleGoName, ") Validate() error {")

	g.F(c.FileId).In()
	g.F(c.FileId).P("validations := []*", vruntime_alias, ".ValidationProcessField {")

	for _, fld := range fields {
		fldGoName, _ := g.BuildFieldName(fld)

		// check if the field itself has validators
		fvals, err := c.FieldGetValidators(g, element, fld)
		if err != nil {
			return err
		}

		// check if the field type has validation
		fhas, err := c.FieldTypeHasValidator(g, element, fld)
		if err != nil {
			return err
		}

		if len(fvals) > 0 || fhas {

			g.F(c.FileId).P("{")
			g.F(c.FileId).In()
			g.F(c.FileId).P(`ProtoName: "`, tpMsg.FullOriginalName(), `", `)
			g.F(c.FileId).P(`FieldName: "`, fld.FieldName(), `", `)
			g.F(c.FileId).P(`ItemValidate: func(iverr `, vruntime_alias, `.ValidationErrorProcessor) {`)
			g.F(c.FileId).In()

			if len(fvals) > 0 {
				var fldType string
				is_array := false
				repeatedType := RT_ARRAY
				var ooFld *fproto.OneOfFieldElement

				switch xfld := fld.(type) {
				case *fproto.FieldElement:
					fldType = xfld.Type
					is_array = xfld.Repeated == true
				case *fproto.MapFieldElement:
					fldType = xfld.Type
					is_array = true
					repeatedType = RT_MAP
				case *fproto.OneOfFieldElement:
					ooFld = xfld
				}

				var ftypedt *fdep.DepType
				if ooFld == nil {
					ftypedt, err = tpMsg.GetType(fldType)
					if err != nil {
						return err
					}
				} else {
					ftypedt = fdep.NewDepTypeOneOf(tpMsg.DepFile, ooFld)
				}

				v_fldName := "m." + fldGoName
				v_index := "nil"
				if is_array {
					// generate validation for repeated field
					for _, fval := range fvals {
						if fvrepeat, fvrepeatok := fval.TypeValidator.(ValidatorRepeated); fvrepeatok {
							g.F(c.FileId).P("// Validation: ", fld.ElementName(), " (REPEATED) = ", fval.Option.Name)
							g.F(c.FileId).P(`iverr.SetContext(nil, "`, fval.Option.Name, `")`)

							err := fvrepeat.GenerateValidationRepeated(g.F(c.FileId), c, repeatedType, ftypedt, fval.Option, v_fldName)
							if err != nil {
								return err
							}
						}
					}

					g.F(c.FileId).P("for msi, ms := range m.", fldGoName, "{")
					g.F(c.FileId).In()

					v_fldName = "ms"
					v_index = "msi"
				}

				for _, fval := range fvals {
					if fvnormal, fvnormalok := fval.TypeValidator.(ValidatorNormal); fvnormalok {
						g.F(c.FileId).P("// Validation: ", fld.ElementName(), " = ", fval.Option.Name)
						g.F(c.FileId).P(`iverr.SetContext(`, v_index, `, "`, fval.Option.Name, `")`)

						err := fvnormal.GenerateValidation(g.F(c.FileId), c, ftypedt, fval.Option, v_fldName)
						if err != nil {
							return err
						}
					}
				}

				if is_array {
					g.F(c.FileId).Out()
					g.F(c.FileId).P("}")
				}
			}

			if fhas {
				// check if the field type has validation

				// err = MyFieldStruct.Validate()

				switch xfld := fld.(type) {
				case *fproto.FieldElement:
					// err = MyFieldStruct.Validate()
					fieldname := "m." + fldGoName
					idxField := "nil"
					if xfld.Repeated {
						g.F(c.FileId).P("for msi, ms := range m.", fldGoName, " {")
						g.F(c.FileId).In()
						fieldname = "ms"
						idxField = "msi"
					}

					g.F(c.FileId).P("if ", fieldname, " != nil {")
					g.F(c.FileId).In()

					g.F(c.FileId).P("err := ", fieldname, ".Validate()")
					//g.F(c.FileId).GenerateErrorCheck("")
					c.GenerateSubvalidationErrorCheck(g, xfld.Name, idxField)

					g.F(c.FileId).Out()
					g.F(c.FileId).P("}")

					if xfld.Repeated {
						g.F(c.FileId).Out()
						g.F(c.FileId).P("}")
					}
				case *fproto.MapFieldElement:
					g.F(c.FileId).P("for msidx, ms := range s.", fldGoName, " {")
					g.F(c.FileId).In()

					g.F(c.FileId).P("err := ms.Validate()")
					//c.GenerateSubvalidationMapErrorCheck(g, xfld.Name, "msidx")
					c.GenerateSubvalidationErrorCheck(g, xfld.Name, "msidx")

					g.F(c.FileId).Out()
					g.F(c.FileId).P("}")
				case *fproto.OneOfFieldElement:
					ooGoName, _ := g.BuildOneOfName(xfld)
					g.F(c.FileId).P("err := ", ooGoName, "_Validate(m.", fldGoName, ")")
					c.GenerateSubvalidationErrorCheck(g, xfld.Name, "nil")
				}
			}

			g.F(c.FileId).Out()
			g.F(c.FileId).P("},")
			g.F(c.FileId).Out()
			g.F(c.FileId).P("},")
		}
	}
	g.F(c.FileId).P("}")

	g.F(c.FileId).P("return ", vruntime_alias, ".ValidationProcess(validations)")
	g.F(c.FileId).Out()
	g.F(c.FileId).P("}")
	g.F(c.FileId).P()

	return nil
}

func (c *Customizer_Validator) generateValidationForOneOf(g *fproto_gowrap.Generator, element *fproto.OneOfFieldElement) error {
	eleGoName, _ := g.BuildOneOfName(element)

	var ooFields []fproto.FieldElementTag

	// func MyOneOf_Validate(m MyOneOf) err
	g.F(c.FileId).P("func ", eleGoName, "_Validate(m ", eleGoName, ") error {")

	g.F(c.FileId).In()
	g.F(c.FileId).P("var err error")
	g.F(c.FileId).P()
	g.F(c.FileId).P("switch me := m.(type) {")

	for _, fld := range element.Fields {
		fldGoName, _ := g.BuildOneOfFieldName(fld)

		// check if the field type has validation
		fhas, err := c.FieldHasValidator(g, element, fld)
		if err != nil {
			return err
		}

		if fhas {
			ooFields = append(ooFields, fld)

			// err = MyFieldStruct.Validate()
			g.F(c.FileId).P("case *", fldGoName, ":")

			g.F(c.FileId).P("err = me.Validate()")
			g.F(c.FileId).GenerateErrorCheck("")
		}
	}

	g.F(c.FileId).P("}")
	g.F(c.FileId).P()

	g.F(c.FileId).P("return err")
	g.F(c.FileId).Out()
	g.F(c.FileId).P("}")
	g.F(c.FileId).P()

	for _, o := range ooFields {
		err := c.generateValidationForMessageOrOneOf(g, o)
		if err != nil {
			return err
		}
	}

	return nil
}

// Check if any field of this type has a known validation type
func (c *Customizer_Validator) TypeHasValidator(g *fproto_gowrap.Generator, element fproto.FProtoElement) (bool, error) {
	var fields []fproto.FieldElementTag

	switch xele := element.(type) {
	case *fproto.MessageElement:
		fields = xele.Fields
	case *fproto.OneOfFieldElement:
		fields = xele.Fields
	}

	for _, fld := range fields {
		fhas, err := c.FieldHasValidator(g, element, fld)
		if err != nil {
			return false, err
		}
		if fhas {
			return true, nil
		}
	}

	return false, nil
}

type FieldValidator struct {
	TypeValidator Validator
	Option        *fproto.OptionElement
}

// Get all validators for the field
func (c *Customizer_Validator) FieldGetValidators(g *fproto_gowrap.Generator, parentElement fproto.FProtoElement, field fproto.FieldElementTag) ([]*FieldValidator, error) {
	var ret []*FieldValidator

	var options []*fproto.OptionElement

	switch xfld := field.(type) {
	case *fproto.FieldElement:
		options = xfld.Options
	case *fproto.MapFieldElement:
		options = xfld.Options
	case *fproto.OneOfFieldElement:
		options = xfld.Options
	}

	for _, o := range options {
		tv, err := c.OptionGetValidator(g, o)
		if err != nil {
			return nil, err
		}

		if tv != nil {
			ret = append(ret, &FieldValidator{
				TypeValidator: tv,
				Option:        o,
			})
		}
	}

	return ret, nil
}

// Check whether the field has any knwon validator
func (c *Customizer_Validator) FieldHasValidator(g *fproto_gowrap.Generator, parentElement fproto.FProtoElement, field fproto.FieldElementTag) (bool, error) {
	var options []*fproto.OptionElement
	var fldType string
	var checkType fproto.FProtoElement

	switch xfld := field.(type) {
	case *fproto.FieldElement:
		options = xfld.Options
		fldType = xfld.Type
	case *fproto.MapFieldElement:
		options = xfld.Options
		fldType = xfld.Type
	case *fproto.OneOfFieldElement:
		options = xfld.Options
		checkType = xfld
	}

	for _, o := range options {
		ohas, err := c.OptionHasValidator(g, o)
		if err != nil {
			return false, err
		}

		if ohas {
			return true, nil
		}
	}

	// check subtype if available (oneof)
	if checkType != nil {
		fhas, err := c.TypeHasValidator(g, checkType)
		if err != nil {
			return false, err
		}
		if fhas {
			return true, nil
		}
	}

	// check if the field type has validators
	if fldType != "" {
		parent_dt := g.GetDepFile().Dep.DepTypeFromElement(parentElement)
		if parent_dt == nil {
			return false, nil
		}

		fdt, err := parent_dt.FindType(fldType)
		if err != nil {
			return false, err
		}

		if fdt == nil {
			return false, err
		}

		// Prevent recursion
		if !parent_dt.IsSame(fdt) {
			if fdt.Item != nil {
				fhas, err := c.TypeHasValidator(g, fdt.Item)
				if err != nil {
					return false, err
				}
				if fhas {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// Check if the field type has any validator
func (c *Customizer_Validator) FieldTypeHasValidator(g *fproto_gowrap.Generator, parentElement fproto.FProtoElement, field fproto.FieldElementTag) (bool, error) {
	var fldType string

	switch xfld := field.(type) {
	case *fproto.FieldElement:
		fldType = xfld.Type
	case *fproto.MapFieldElement:
		fldType = xfld.Type
	case *fproto.OneOfFieldElement:
		return c.TypeHasValidator(g, field)
	}

	// check if the field type has validators
	if fldType != "" {
		parent_dt := g.GetDepFile().Dep.DepTypeFromElement(parentElement)
		if parent_dt == nil {
			return false, nil
		}

		fdt, err := parent_dt.FindType(fldType)
		if err != nil {
			return false, err
		}

		if fdt == nil {
			return false, err
		}

		// Prevent recursion
		if !parent_dt.IsSame(fdt) && !fdt.IsScalar() {
			return c.TypeHasValidator(g, fdt.Item)
		}
	}

	return false, nil
}

// Gets the FIELD option validation
func (c *Customizer_Validator) OptionGetValidator(g *fproto_gowrap.Generator, opt *fproto.OptionElement) (Validator, error) {
	opttype, err := g.GetDep().GetOption(fdep.FIELD_OPTION, opt.ParenthesizedName)
	if err != nil {
		return nil, err
	}

	if opttype == nil {
		return nil, nil
	}

	if tv := c.FindValidatorForOption(opttype); tv != nil {
		return tv, nil
	}

	return nil, nil
}

// Checks if the option has a validator
func (c *Customizer_Validator) OptionHasValidator(g *fproto_gowrap.Generator, opt *fproto.OptionElement) (bool, error) {
	opttype, err := g.GetDep().GetOption(fdep.FIELD_OPTION, opt.ParenthesizedName)
	if err != nil {
		return false, err
	}

	if opttype == nil {
		return false, nil
	}

	if tv := c.FindValidatorForOption(opttype); tv != nil {
		return true, nil
	}

	return false, nil
}

// Finds a validator for an option
func (c *Customizer_Validator) FindValidatorForOption(optType *fdep.OptionType) Validator {
	for _, v := range c.Validators {
		tv := v.GetValidator(optType)
		if tv != nil {
			return tv
		}
	}
	return nil
}

func (c *Customizer_Validator) GenerateServiceCode(g *fproto_gowrap.Generator) error {
	return nil
}

func (c *Customizer_Validator) GenerateValidationErrorCheck(g *fproto_gowrap.Generator, varError string, validationItem string, errorId ValidationErrorId, errorParams ...string) {
	g.F(c.FileId).P("if ", varError, " != nil {")
	g.F(c.FileId).In()

	c.GenerateValidationErrorAdd(g, varError, validationItem, errorId, errorParams...)

	g.F(c.FileId).Out()
	g.F(c.FileId).P("}")
}

func (c *Customizer_Validator) GenerateValidationErrorAdd(g *fproto_gowrap.Generator, varError string, validationItem string, errorId ValidationErrorId, errorParams ...string) {
	var ep []string
	for _, errp := range errorParams {
		ep = append(ep, strconv.Quote(errp))
	}
	epstr := ""
	if len(ep) > 0 {
		epstr = ", " + strings.Join(ep, ", ")
	}

	g.F(c.FileId).P(`iverr.AddError("`, validationItem, `", `, varError, `, "`, errorId, `"`, epstr, `)`)
}

func (c *Customizer_Validator) GenerateSubvalidationErrorCheck(g *fproto_gowrap.Generator, fieldName string, varIndex string) {
	g.F(c.FileId).P("if err != nil {")
	g.F(c.FileId).In()
	g.F(c.FileId).P(`iverr.AddValidateError(`, varIndex, `, err)`)
	g.F(c.FileId).Out()
	g.F(c.FileId).P("}")
}
