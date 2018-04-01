package fproto_gowrap_validate

import (
	"github.com/RangelReale/fdep"
	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap/gowrap"
)

// Adds a json tag to all struct fields, using snake case formatting
type Customizer_Validate struct {
	FileId         string
	TypeValidators []TypeValidatorPlugin
}

func NewCustomizer_Validate() *Customizer_Validate {
	return &Customizer_Validate{
		FileId:         fproto_gowrap.FILEID_MAIN,
		TypeValidators: []TypeValidatorPlugin{&TypeValidatorPlugin_Default{}},
	}
}

func NewCustomizer_Validate_Custom() *Customizer_Validate {
	return &Customizer_Validate{}
}

func (c *Customizer_Validate) GetValidator(validatorType *fdep.OptionType) TypeValidator {
	for _, tcp := range c.TypeValidators {
		tc := tcp.GetTypeValidator(validatorType)
		if tc != nil {
			return tc
		}
	}
	return nil
}

func (c *Customizer_Validate) GenerateCode(g *fproto_gowrap.Generator) error {
	var validate_elements []fproto.FProtoElement

	if g.GetDepFile().ProtoFile != nil {
		for _, msg := range g.GetDepFile().ProtoFile.CollectMessages() {
			fhas, err := c.TypeHasValidator(g, msg)
			if err != nil {
				return err
			}

			if fhas {
				validate_elements = append(validate_elements, msg)
			}
		}

		for _, oofield := range g.GetDepFile().ProtoFile.CollectFields() {
			if oof, isoof := oofield.(*fproto.OneOfFieldElement); isoof {
				fhas, err := c.TypeHasValidator(g, oof)
				if err != nil {
					return err
				}

				if fhas {
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

func (c *Customizer_Validate) generateValidationForElement(g *fproto_gowrap.Generator, element fproto.FProtoElement) error {
	switch el := element.(type) {
	case *fproto.MessageElement:
		return c.generateValidationForMessageOrOneOf(g, el)
	case *fproto.OneOfFieldElement:
		return c.generateValidationForOneOf(g, el)
	}
	return nil
}

func (c *Customizer_Validate) generateValidationForMessageOrOneOf(g *fproto_gowrap.Generator, element fproto.FProtoElement) error {
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
	g.F(c.FileId).P("var err error")
	for _, fld := range fields {
		fldGoName, _ := g.BuildFieldName(fld)

		// check if the field itself has validators
		fvals, err := c.FieldGetValidators(g, element, fld)
		if err != nil {
			return err
		}

		if len(fvals) > 0 {
			var fldType string

			switch xfld := fld.(type) {
			case *fproto.FieldElement:
				fldType = xfld.Type
			case *fproto.MapFieldElement:
				fldType = xfld.Type
			}

			if fldType != "" {
				ftypedt, err := tpMsg.GetType(fldType)
				if err != nil {
					return err
				}

				for _, fval := range fvals {
					err := fval.TypeValidator.GenerateValidation(g.F(c.FileId), ftypedt, fval.Option, "m."+fldGoName, "err")
					if err != nil {
						return err
					}
				}
			}
		}

		// check if the field type has validation
		fhas, err := c.FieldTypeHasValidator(g, element, fld)

		if fhas {
			// err = MyFieldStruct.Validate()

			switch xfld := fld.(type) {
			case *fproto.FieldElement:
				// err = MyFieldStruct.Validate()
				fieldname := "m." + fldGoName
				if xfld.Repeated {
					g.F(c.FileId).P("for _, ms := range m.", fldGoName, " {")
					g.F(c.FileId).In()
					fieldname = "ms"
				}

				g.F(c.FileId).P("if ", fieldname, " != nil {")
				g.F(c.FileId).In()

				g.F(c.FileId).P("err = ", fieldname, ".Validate()")
				g.F(c.FileId).GenerateErrorCheck("")

				g.F(c.FileId).Out()
				g.F(c.FileId).P("}")

				if xfld.Repeated {
					g.F(c.FileId).Out()
					g.F(c.FileId).P("}")
				}
			case *fproto.MapFieldElement:
				g.F(c.FileId).P("for msidx, ms := range s.", fldGoName, " {")
				g.F(c.FileId).In()

				g.F(c.FileId).P("err = ms.Validate()")
				g.F(c.FileId).GenerateErrorCheck("")

				g.F(c.FileId).Out()
				g.F(c.FileId).P("}")
			case *fproto.OneOfFieldElement:
				// Will be validated separatelly
			}
		}
	}

	g.F(c.FileId).P("return err")
	g.F(c.FileId).Out()
	g.F(c.FileId).P("}")
	g.F(c.FileId).P()

	return nil
}

func (c *Customizer_Validate) generateValidationForOneOf(g *fproto_gowrap.Generator, element *fproto.OneOfFieldElement) error {
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

func (c *Customizer_Validate) TypeHasValidator(g *fproto_gowrap.Generator, element fproto.FProtoElement) (bool, error) {
	// Check if any field of this type has a known validation type
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

type fieldValidator struct {
	TypeValidator TypeValidator
	Option        *fproto.OptionElement
}

func (c *Customizer_Validate) FieldGetValidators(g *fproto_gowrap.Generator, parentElement fproto.FProtoElement, field fproto.FieldElementTag) ([]*fieldValidator, error) {
	var ret []*fieldValidator

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
			ret = append(ret, &fieldValidator{
				TypeValidator: tv,
				Option:        o,
			})
		}
	}

	return ret, nil
}

func (c *Customizer_Validate) FieldHasValidator(g *fproto_gowrap.Generator, parentElement fproto.FProtoElement, field fproto.FieldElementTag) (bool, error) {
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

func (c *Customizer_Validate) FieldTypeHasValidator(g *fproto_gowrap.Generator, parentElement fproto.FProtoElement, field fproto.FieldElementTag) (bool, error) {
	var fldType string

	switch xfld := field.(type) {
	case *fproto.FieldElement:
		fldType = xfld.Type
	case *fproto.MapFieldElement:
		fldType = xfld.Type
	case *fproto.OneOfFieldElement:
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

func (c *Customizer_Validate) OptionGetValidator(g *fproto_gowrap.Generator, opt *fproto.OptionElement) (TypeValidator, error) {
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

func (c *Customizer_Validate) OptionHasValidator(g *fproto_gowrap.Generator, opt *fproto.OptionElement) (bool, error) {
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

func (c *Customizer_Validate) FindValidatorForOption(optType *fdep.OptionType) TypeValidator {
	for _, v := range c.TypeValidators {
		tv := v.GetTypeValidator(optType)
		if tv != nil {
			return tv
		}
	}
	return nil
}

func (c *Customizer_Validate) GenerateServiceCode(g *fproto_gowrap.Generator) error {
	return nil
}
