package validator_runtime

import (
	"reflect"

	"github.com/RangelReale/fproto-wrap-validator/gowrap"
)

//
// Error
//
type Error struct {
	Fields map[string]*ErrorField
}

func (e *Error) Error() string {
	return "Validation error"
}

func (e *Error) IsEmpty() bool {
	return e.Fields == nil || len(e.Fields) == 0
}

type ErrorField struct {
	ProtoName        string
	FieldName        string
	ValidationErrors []*ErrorValidatorError
}

type ErrorValidatorError struct {
	Index            int
	MapIndex         string
	ValidationOption string
	ValidationItem   string
	Err              error
	ErrorId          fproto_gowrap_validator.ValidationErrorId
	ErrorParams      map[string]string
}

//
// Process
//

type ValidationProcessField struct {
	ProtoName    string
	FieldName    string
	ItemValidate func(ValidationErrorProcessor)
}

type ValidationErrorProcessor interface {
	SetContext(index interface{}, validationOption string)
	AddError(validationItem string, err error, errorId string, errorParams ...string)
	AddValidateError(index interface{}, err error, errorParams ...string)
}

type vep struct {
	curfield            *ValidationProcessField
	curindex            interface{}
	curvalidationoption string

	err *Error
}

func (v *vep) SetContext(index interface{}, validationOption string) {
	v.curindex = index
	v.curvalidationoption = validationOption
}

func (v *vep) AddError(validationItem string, err error, errorId string, errorParams ...string) {
	ef, efok := v.err.Fields[v.curfield.FieldName]
	if !efok {
		ef = &ErrorField{
			FieldName: v.curfield.FieldName,
			ProtoName: v.curfield.ProtoName,
		}
		v.err.Fields[v.curfield.FieldName] = ef
	}

	verror := &ErrorValidatorError{
		ValidationOption: v.curvalidationoption,
		ValidationItem:   validationItem,
		Err:              err,
		ErrorId:          fproto_gowrap_validator.ValidationErrorId(errorId),
		ErrorParams:      v.parseErrorParams(errorParams...),
	}
	switch iv := v.curindex.(type) {
	case nil:
		verror.Index = -1
	case string:
		verror.MapIndex = iv
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		verror.Index = int(reflect.ValueOf(v.curindex).Int())
	default:
		panic("Unknown index type")
	}

	ef.ValidationErrors = append(ef.ValidationErrors, verror)
}

func (v *vep) AddValidateError(index interface{}, err error, errorParams ...string) {
	ef, efok := v.err.Fields[v.curfield.FieldName]
	if !efok {
		ef = &ErrorField{
			FieldName: v.curfield.FieldName,
			ProtoName: v.curfield.ProtoName,
		}
		v.err.Fields[v.curfield.FieldName] = ef
	}

	verror := &ErrorValidatorError{
		Err:         err,
		ErrorParams: v.parseErrorParams(errorParams...),
	}
	switch iv := index.(type) {
	case nil:
		verror.Index = -1
	case string:
		verror.MapIndex = iv
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		verror.Index = int(reflect.ValueOf(index).Int())
	default:
		panic("Unknown index type")
	}

	ef.ValidationErrors = append(ef.ValidationErrors, verror)
}

func (v *vep) setCurField(f *ValidationProcessField) {
	v.curfield = f
	v.curindex = nil
	v.curvalidationoption = ""
}

func (e *vep) parseErrorParams(errorParams ...string) map[string]string {
	if len(errorParams) > 0 {
		ret := make(map[string]string)
		var lastval *string
		for _, p := range errorParams {
			if lastval == nil {
				lastval = &p
			} else {
				ret[*lastval] = p
				lastval = nil
			}
		}
		return ret
	}
	return nil
}

func ValidationProcess(validations []*ValidationProcessField) error {
	v := &vep{
		err: &Error{
			Fields: make(map[string]*ErrorField),
		},
	}
	for _, val := range validations {
		v.setCurField(val)
		val.ItemValidate(v)
	}

	if v.err.IsEmpty() {
		return nil
	}

	return v.err
}
