package validation

import (
	"regexp"

	validator "github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Validator interface
type Validator interface {
	Validate() error
}

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("regexp", regexTag)
}

func regexTag(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	if field == "" {
		return true
	}
	regexString := fl.Param()
	regex := regexp.MustCompile(regexString)
	match := regex.MatchString(field)
	return match
}

// Validate for struct: Validate(&dto)
func Validate(dto interface{}) error {
	//kind := reflect.TypeOf(dto)
	//if kind.Kind() != reflect.Ptr {
	//	return fmt.Errorf("invalid dto type, must be pointer")
	//}

	// If struct implements the Validator interface, call it
	if v, ok := dto.(Validator); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	// Perform validate.Struct verification by default
	if validate == nil {
		return nil
	}
	return validate.Struct(dto)
}
