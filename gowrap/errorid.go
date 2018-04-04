package fproto_gowrap_validator

type ValidationErrorId string

func (v ValidationErrorId) String() string {
	return string(v)
}

const (
	VEID_UNKNOWN        ValidationErrorId = "fpv.VEID_UNKNOWN"
	VEID_INTERNAL_ERROR ValidationErrorId = "fpv.VEID_INTERNAL_ERROR"
	VEID_REQUIRED       ValidationErrorId = "fpv.VEID_REQUIRED"
	VEID_LENGTH         ValidationErrorId = "fpv.VEID_LENGTH"
	VEID_PATTERN        ValidationErrorId = "fpv.VEID_PATTERN"
	VEID_MINMAX         ValidationErrorId = "fpv.VEID_MINMAX"
	VEID_INVALID_VALUE  ValidationErrorId = "fpv.VEID_INVALID_VALUE"
)
