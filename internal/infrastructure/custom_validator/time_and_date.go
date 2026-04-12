package custom_validator

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var timeHHMMRegex = regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)

func TimeHHMM(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if !timeHHMMRegex.MatchString(val) {
		return false
	}
	_, err := parseTimeHHMM(val)
	return err == nil
}

func parseTimeHHMM(s string) (t time.Time, err error) {
	return time.Parse("15:04", s)
}

func IsUUID(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

var dateISORegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func DateISO(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if !dateISORegex.MatchString(val) {
		return false
	}
	_, err := time.Parse("2006-01-02", val)
	return err == nil
}

func Register(v *validator.Validate) {
	_ = v.RegisterValidation("time_hhmm", TimeHHMM)
	_ = v.RegisterValidation("uuid_custom", IsUUID)
	_ = v.RegisterValidation("date_iso", DateISO)
}
