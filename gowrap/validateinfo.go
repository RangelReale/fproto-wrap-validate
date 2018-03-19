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

	// check all fields for validators
	for _, pf := range vi.g.GetDep().Files {
		for _, m := range pf.ProtoFile.CollectMessages() {
			has_validator := false
			for _, mf := range m.(*fproto.MessageElement).Fields {
				var opt []*fproto.OptionElement
				switch xfld := mf.(type) {
				case *fproto.FieldElement:
					opt = xfld.Options
				case *fproto.MapFieldElement:
					opt = xfld.Options
				case *fproto.OneOfFieldElement:
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

								typeconv := vp.TypeValidatorPlugin.GetTypeValidator(opttype)
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
				vi.addMessageValidation(pf, m.(*fproto.MessageElement))
			}
		}
	}

	// check message field types for validator. If have any, also validate the message.
	// Do multiple passes
	is_add := true
	for is_add {
		is_add = false
		for _, pf := range vi.g.GetDep().Files {
			for _, m := range pf.ProtoFile.CollectMessages() {
				dtMsg := vi.g.GetDep().DepTypeFromElement(m)

				has_validator := false
				for _, mf := range m.(*fproto.MessageElement).Fields {
					switch xfld := mf.(type) {
					case *fproto.FieldElement:
						fdt, err := dtMsg.GetType(xfld.Type)
						if err != nil {
							return err
						}

						if vi.TypeHasValidation(fdt) {
							has_validator = true
						}
					case *fproto.MapFieldElement:
						fdt, err := dtMsg.GetType(xfld.Type)
						if err != nil {
							return err
						}
						kfdt, err := dtMsg.GetType(xfld.KeyType)
						if err != nil {
							return err
						}

						if vi.TypeHasValidation(fdt) {
							has_validator = true
						}
						if vi.TypeHasValidation(kfdt) {
							has_validator = true
						}
					case *fproto.OneOfFieldElement:
						// TODO
					}

				}
				if has_validator {
					if vi.addMessageValidation(pf, m.(*fproto.MessageElement)) {
						is_add = true
					}
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
	case *fproto.OneOfFieldElement:
		opt = xfld.Options
	}

	for _, o := range opt {
		tc, ok := vi.used_validators[o.ParenthesizedName]
		if ok {
			return tc, o, true
		}
	}

	// check if the field type have field validation

	return nil, nil, false
}

func (vi *validateInfo_Default) IsFieldTypeValidate(msg *fproto.MessageElement, field fproto.FieldElementTag) (bool, error) {
	dtMsg := vi.g.GetDep().DepTypeFromElement(msg)

	switch xfld := field.(type) {
	case *fproto.FieldElement:
		fdt, err := dtMsg.GetType(xfld.Type)
		if err != nil {
			return false, err
		}

		if vi.TypeHasValidation(fdt) {
			return true, nil
		}
	case *fproto.MapFieldElement:
		fdt, err := dtMsg.GetType(xfld.Type)
		if err != nil {
			return false, err
		}
		kfdt, err := dtMsg.GetType(xfld.KeyType)
		if err != nil {
			return false, err
		}

		if vi.TypeHasValidation(fdt) {
			return true, nil
		}
		if vi.TypeHasValidation(kfdt) {
			return true, nil
		}
	case *fproto.OneOfFieldElement:
		// TODO
	}
	return false, nil
}

func (vi *validateInfo_Default) addMessageValidation(filedep *fdep.FileDep, message *fproto.MessageElement) bool {
	mtp := fdep.NewDepTypeFromElement(vi.g.GetFileDep(), message)

	if !vi.TypeHasValidation(mtp) {
		vi.msg_validation = append(vi.msg_validation, mtp)
		if filedep.IsSame(vi.g.GetFileDep()) {
			vi.cur_msg_validation = append(vi.cur_msg_validation, mtp)
		}
		return true
	}
	return false
}

func (vi *validateInfo_Default) TypeHasValidation(tp *fdep.DepType) bool {
	for _, m := range vi.msg_validation {
		if m.IsSame(tp) {
			return true
		}
	}
	return false
}
