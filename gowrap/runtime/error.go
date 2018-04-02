package validator_runtime

import (
	"fmt"
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

func (e *ValidationProcess) SetContext(protoName string, fieldName string, index int, validationOption string) {
	e.context = &ValidationErrorItem{
		ProtoName:        protoName,
		FieldName:        fieldName,
		Index:            index,
		ValidationOption: validationOption,
	}
}

func (e *ValidationProcess) AddError(validationItem string, err error, errorId string) {
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
	})
}

func (e *ValidationProcess) AddValidateError(fieldName string, index int, err error) {
	e.errors = append(e.errors, &ValidationErrorItem{
		FieldName: fieldName,
		Index:     index,
		Err:       err,
	})
}

func (e *ValidationProcess) AddValidateMapError(fieldName string, mapIndex string, err error) {
	e.errors = append(e.errors, &ValidationErrorItem{
		FieldName: fieldName,
		MapIndex:  mapIndex,
		Err:       err,
	})
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
}

func (vi *ValidationErrorItem) ErrorDescription(parent string) string {
	if ivi, ok := vi.Err.(*Error); ok {
		return ivi.ErrorDescription(parent + "." + vi.FieldName)
	}
	return fmt.Sprintf("%s.%s: %s", parent, vi.FieldName, vi.Err.Error())
}
