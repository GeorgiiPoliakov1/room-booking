package validator

import (
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/validator"
)

var (
	once     sync.Once
	instance *validator.Validate
)

func GetValidator() *validator.Validate {
	once.Do(func() {
		instance = validator.New()

		// Настройка тегов JSON для ошибок
		instance.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		registerCustomValidators(instance)
	})
	return instance
}

func registerCustomValidators(v *validator.Validate) {

	v.RegisterValidation("room_name", validateRoomName)

	v.RegisterValidation("email", validateEmail)

	v.RegisterValidation("role", validateRole)

	v.RegisterValidation("time_slot", validateTimeSlot)
}

func validateRoomName(fl validator.FieldLevel) bool {
	name := fl.Field().String()

	forbidden := []string{"admin", "root", "system", "test"}
	for _, f := range forbidden {
		if strings.EqualFold(name, f) {
			return false
		}
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9\\s\\-_\u0400-\u04FF]+$", name)
	return matched
}

func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return emailRegex.MatchString(strings.ToLower(email))
}

func validateRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	validRoles := map[string]bool{
		"admin": true,
		"user":  true,
	}
	return validRoles[strings.ToLower(role)]
}

func validateTimeSlot(fl validator.FieldLevel) bool {
	timeStr := fl.Field().String()
	matched, _ := regexp.MatchString(`^([01][0-9]|2[0-3]):[0-5][0-9]$`, timeStr)
	return matched
}
