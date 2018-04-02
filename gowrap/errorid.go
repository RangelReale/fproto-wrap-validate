package fproto_gowrap_validator

type ValidationErrorId string

func (v ValidationErrorId) String() string {
	return string(v)
}

const (
	VEID_UNKNOWN  ValidationErrorId = "fpv.VEID_UNKNOWN"
	VEID_REQUIRED ValidationErrorId = "fpv.VEID_REQUIRED"
	VEID_LENGTH   ValidationErrorId = "fpv.VEID_LENGTH"
)
