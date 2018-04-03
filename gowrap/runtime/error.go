package validator_runtime

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/RangelReale/fproto-wrap-validator/gowrap"
)

type Error struct {
	Errors []*ValidationErrorItem
}

func (e *Error) ErrorDescription(parent string) string {
	if len(e.Errors) > 0 {
		var ret strings.Builder
		for _, ei := range e.Errors {
			ret.WriteString(ei.ErrorDescription(parent))
			ret.WriteByte('\n')
		}
		return ret.String()
	}
	return "Validation error"
}

func (e *Error) Error() string {
	return e.ErrorDescription("")
}

//
// ValidationProcess
//
type ValidationProcess struct {
	errors []*ValidationErrorItem

	context *ValidationErrorItem
}

func (e *ValidationProcess) Err() *Error {
	if len(e.errors) > 0 {
		return &Error{Errors: e.errors}
	}
	return nil
}

func (e *ValidationProcess) SetContext(protoName string, fieldName string, index interface{}, validationOption string) {
	newctx := &ValidationErrorItem{
		ProtoName:        protoName,
		FieldName:        fieldName,
		ValidationOption: validationOption,
	}
	switch iv := index.(type) {
	case nil:
		newctx.Index = -1
	case string:
		newctx.MapIndex = iv
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		newctx.Index = int(reflect.ValueOf(index).Int())
	default:
		panic("Unknown index type")
	}
	e.context = newctx
}

func (e *ValidationProcess) AddError(validationItem string, err error, errorId string, errorParams ...string) {
	if e.context == nil {
		panic("Validation context is nil - must call SetContext first")
	}

	e.errors = append(e.errors, &ValidationErrorItem{
		ProtoName:        e.context.ProtoName,
		FieldName:        e.context.FieldName,
		Index:            e.context.Index,
		ValidationOption: e.context.ValidationOption,
		ValidationItem:   validationItem,
		Err:              err,
		ErrorId:          fproto_gowrap_validator.ValidationErrorId(errorId),
		ErrorParams:      e.parseErrorParams(errorParams...),
	})
}

func (e *ValidationProcess) AddValidateError(fieldName string, index int, err error, errorParams ...string) {
	e.errors = append(e.errors, &ValidationErrorItem{
		FieldName: fieldName,
		Index:     index,
		Err:       err,
	})
}

func (e *ValidationProcess) AddValidateMapError(fieldName string, mapIndex string, err error, errorParams ...string) {
	e.errors = append(e.errors, &ValidationErrorItem{
		FieldName: fieldName,
		MapIndex:  mapIndex,
		Err:       err,
	})
}

func (e *ValidationProcess) parseErrorParams(errorParams ...string) map[string]string {
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

type ValidationErrorItem struct {
	ProtoName        string
	FieldName        string
	Index            int
	MapIndex         string
	ValidationOption string
	ValidationItem   string
	Err              error
	ErrorId          fproto_gowrap_validator.ValidationErrorId
	ErrorParams      map[string]string
}

func (vi *ValidationErrorItem) ErrorDescription(parent string) string {
	if ivi, ok := vi.Err.(*Error); ok {
		return ivi.ErrorDescription(parent + "." + vi.FieldName)
	}
	return fmt.Sprintf("%s.%s: %s", parent, vi.FieldName, vi.Err.Error())
}
