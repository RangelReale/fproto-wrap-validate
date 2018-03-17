package fproto_gowrap_validate

import (
	"fmt"
	"strings"

	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap/gowrap"
	"github.com/RangelReale/fproto/fdep"
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
		tc := tcp.GetTypeConverter(validatorType)
		if tc != nil {
			return tc
		}
	}
	return nil
}

func (c *Customizer_Validate) GetTag(g *fproto_gowrap.Generator, currentTag *fproto_gowrap.StructTag, parentItem fproto.FProtoElement, item fproto.FProtoElement) error {
	return nil
}

func (c *Customizer_Validate) GenerateCode(g *fproto_gowrap.Generator) error {
	vinfo, err := newvalidateInfo_Default(g, c.TypeValidators)
	if err != nil {
		return err
	}

	if len(vinfo.cur_msg_validation) > 0 {
		g.F(c.FileId).P("//")
		g.F(c.FileId).P("// VALIDATION")
		g.F(c.FileId).P("//")
		g.F(c.FileId).P()

		for _, m := range vinfo.cur_msg_validation {
			msg := m.Item.(*fproto.MessageElement)

			// func (m* MyMessage) Validate() err
			msgGoName, _, _ := g.BuildMessageName(msg)

			g.F(c.FileId).P("func (m *", msgGoName, ") Validate() error {")
			g.F(c.FileId).In()
			g.F(c.FileId).P("var err error")
			for _, fld := range msg.Fields {
				ftc, opt, fval := vinfo.IsFieldValidate(fld)
				if fval {
					fldGoName, _ := g.BuildFieldName(fld)

					check_err, err := ftc.GenerateValidation(g.F(c.FileId), fdep.NewDepTypeFromElement(g.GetFileDep(), fld), opt, "m."+fldGoName, "err")
					if err != nil {
						return err
					}
					if check_err {
						g.F(c.FileId).GenerateErrorCheck("")
					}
				}
			}
			g.F(c.FileId).P("return err")
			g.F(c.FileId).Out()
			g.F(c.FileId).P("}")
			g.F(c.FileId).P()
		}
	}

	return nil
}

func (c *Customizer_Validate) GenerateCodeOld(g *fproto_gowrap.Generator) error {
	// find all messages that have validation
	found_val := make(map[string]TypeValidator)

	type tvp struct {
		Prefix              string
		TypeValidatorPlugin TypeValidatorPlugin
	}
	var tvplist []*tvp
	for _, tv := range c.TypeValidators {
		for _, tp := range tv.ValidatorPrefixes() {
			tvplist = append(tvplist, &tvp{tp, tv})
		}
	}

	for _, m := range g.GetFileDep().ProtoFile.CollectMessages() {
		for _, mf := range m.(*fproto.MessageElement).Fields {
			var opt []*fproto.OptionElement
			switch xfld := mf.(type) {
			case *fproto.FieldElement:
				opt = xfld.Options
			case *fproto.MapFieldElement:
				opt = xfld.Options
			case *fproto.OneofFieldElement:
				opt = xfld.Options
			}

			for _, o := range opt {
				for _, vp := range tvplist {
					if strings.HasPrefix(o.ParenthesizedName, vp.Prefix) {
						opttype, err := g.GetDep().GetOption(fdep.FIELD_OPTION, o.ParenthesizedName)
						if err != nil {
							return fmt.Errorf("Error retrieving file: %v", err)
						}

						if opttype == nil {
							return fmt.Errorf("Couldn't find type for option %s", o.ParenthesizedName)
						}

						typeconv := vp.TypeValidatorPlugin.GetTypeConverter(opttype)
						if typeconv != nil {
							found_val[o.ParenthesizedName] = typeconv
						}
					}
				}
			}
		}
	}

	for fvn, _ := range found_val {
		fmt.Printf("Found val: %s\n", fvn)
	}

	return nil
}

func (c *Customizer_Validate) Testing(g *fproto_gowrap.Generator) error {
	// find types that have validation

	for _, m := range g.GetFileDep().ProtoFile.Messages {
		for _, mf := range m.Fields {
			o := mf.FindOption("validate.field")
			if o != nil {
				//op, err := g.GetDep().GetOption(fdep.FIELD_OPTION, "validate.field")
				op, err := g.GetDep().GetOption(fdep.FIELD_OPTION, "packed")
				if err != nil {
					return fmt.Errorf("Error retrieving file: %v", err)
				}

				if op != nil {
					if op.Option != nil {
						fmt.Printf("%s: %s [%s]\n", op.SourceOption.FileDep.FilePath, op.Option.FileDep.FilePath, op.Name)
					} else {
						fmt.Printf("%s: [%s]\n", op.SourceOption.FileDep.FilePath, op.Name)
					}
				} else {
					fmt.Printf("Option not found\n")
				}

				//fl, err := g.GetDep().GetFileOfName("validate.field")
				fl, err := g.GetFileDep().GetFileOfName("validate.field")
				if err != nil {
					return fmt.Errorf("Error retrieving file: %v", err)
				}

				if fl != nil {
					fmt.Println(fl.FileDep.FilePath)
				}

				ffodt, err := g.GetDepType("", "validate.google.protobuf.FieldOptions")
				if err != nil {
					return fmt.Errorf("Error retrieving type: %v", err)
				}

				if ffodt != nil {
					fmt.Println(ffodt.FullOriginalName())
				}

				fodt, err := g.GetDepType("", "google.protobuf.FieldOptions")
				if err != nil {
					return fmt.Errorf("Error retrieving type: %v", err)
				}

				if fodt != nil {
					fmt.Println(fodt.FullOriginalName())
				}

				fxdt, err := fodt.GetTypeExtension("validate")
				if err != nil {
					return fmt.Errorf("Error retrieving extension: %v", err)
				}

				/*
					fxdt, err := g.GetDep().GetTypeExtension("google.protobuf.FieldOptions", "validate")
					if err != nil {
						return fmt.Errorf("Error retrieving extension: %v", err)
					}
				*/

				if fxdt != nil {
					fmt.Println(fxdt.FullOriginalName())
				}

				/*
					_, err := g.GetDepType("", "validate.field")
					if err != nil {
						return fmt.Errorf("Could not find validate.field type: %v", err)
					}
				*/

				msgGoType, _, _ := g.BuildMessageName(m)

				fldGoType, _ := g.BuildFieldName(mf)

				g.FMain().P("// VALIDATE: ", msgGoType, " -- ", fldGoType, " @@@ ", o.Name)
				for ooname, oo := range o.AggregatedValues {
					g.FMain().P("// VALIDATE.VALIDATION: ", ooname, " -- ", oo.Source)
				}
				g.FMain().P()
			}
		}

	}

	return nil
}

func (c *Customizer_Validate) GenerateServiceCode(g *fproto_gowrap.Generator) error {
	return nil
}
