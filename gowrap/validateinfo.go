package fproto_gowrap_validate

import (
	"fmt"
	"strings"

	"github.com/RangelReale/fproto"
	"github.com/RangelReale/fproto-wrap/gowrap"
	"github.com/RangelReale/fproto/fdep"
)

type ValidateInfo interface {
	TypeHasValidation(tp *fdep.DepType) bool
}

type validateInfo_Default struct {
	g *fproto_gowrap.Generator

	typeValidators []TypeValidatorPlugin

	// used validator by parenthesized name
	used_validators map[string]TypeValidator

	// messages that have validation
	msg_validation []*fdep.DepType

	cur_msg_validation []*fdep.DepType
}

func newvalidateInfo_Default(g *fproto_gowrap.Generator, typeValidators []TypeValidatorPlugin) (*validateInfo_Default, error) {
	ret := &validateInfo_Default{
		g:               g,
		typeValidators:  typeValidators,
		used_validators: make(map[string]TypeValidator),
	}
	err := ret.load()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (vi *validateInfo_Default) load() error {
	type tvp struct {
		Prefix              string
		TypeValidatorPlugin TypeValidatorPlugin
	}
	var tvplist []*tvp
	for _, tv := range vi.typeValidators {
		for _, tp := range tv.ValidatorPrefixes() {
			tvplist = append(tvplist, &tvp{tp, tv})
		}
	}

	for _, pf := range vi.g.GetDep().Files {
		for _, m := range vi.g.GetFileDep().ProtoFile.CollectMessages() {
			has_validator := false
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
							_, used := vi.used_validators[o.ParenthesizedName]

							if !used {
								opttype, err := vi.g.GetDep().GetOption(fdep.FIELD_OPTION, o.ParenthesizedName)
								if err != nil {
									return fmt.Errorf("Error retrieving file: %v", err)
								}

								if opttype == nil {
									return fmt.Errorf("Couldn't find type for option %s", o.ParenthesizedName)
								}

								typeconv := vp.TypeValidatorPlugin.GetTypeConverter(opttype)
								if typeconv != nil {
									vi.used_validators[o.ParenthesizedName] = typeconv
									has_validator = true
								}
							} else {
								has_validator = true
							}
						}
					}
				}
			}
			if has_validator {
				vi.msg_validation = append(vi.msg_validation, fdep.NewDepTypeFromElement(vi.g.GetFileDep(), m))

				if pf.IsSame(vi.g.GetFileDep()) {
					vi.cur_msg_validation = append(vi.cur_msg_validation, fdep.NewDepTypeFromElement(vi.g.GetFileDep(), m))
				}
			}
		}
	}

	return nil
}

func (vi *validateInfo_Default) IsFieldValidate(field fproto.FieldElementTag) (TypeValidator, *fproto.OptionElement, bool) {
	var opt []*fproto.OptionElement
	switch xfld := field.(type) {
	case *fproto.FieldElement:
		opt = xfld.Options
	case *fproto.MapFieldElement:
		opt = xfld.Options
	case *fproto.OneofFieldElement:
		opt = xfld.Options
	}

	for _, o := range opt {
		tc, ok := vi.used_validators[o.ParenthesizedName]
		if ok {
			return tc, o, true
		}
	}

	return nil, nil, false
}

func (vi *validateInfo_Default) TypeHasValidation(tp *fdep.DepType) bool {
	for _, m := range vi.msg_validation {
		if m.IsSame(tp) {
			return true
		}
	}
	return false
}
