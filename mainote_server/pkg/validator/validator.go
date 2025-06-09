package validator

import (
	"reflect"
	"strings"
	"time"

	v "github.com/go-playground/validator/v10"
	"github.com/gobeam/stringy"
	"github.com/rs/zerolog/log"
)

const timeTolerance = time.Hour

func New() *v.Validate {
	validator := v.New(v.WithRequiredStructEnabled())

	err := validator.RegisterValidation("gtfield_if_set", func(fl v.FieldLevel) bool {
		other := fl.Parent().FieldByName(fl.Param())
		if other.IsNil() {
			return true
		}

		return validator.VarWithValue(fl.Field().Interface(), other.Interface(), "gtfield") == nil
	})
	if err != nil {
		log.Err(err).Msg("failed to register gtfield_if_set validator")
	}

	err = validator.RegisterValidation("ltfield_if_set", func(fl v.FieldLevel) bool {
		other := fl.Parent().FieldByName(fl.Param())
		if other.IsNil() {
			return true
		}

		return validator.VarWithValue(fl.Field().Interface(), other.Interface(), "ltfield") == nil
	})
	if err != nil {
		log.Warn().Err(err).Msg("failed to register ltfield_if_set validator")
	}

	err = validator.RegisterValidation("timestamp_not_in_future", func(fl v.FieldLevel) bool {
		var timestamp int64
		now := time.Now().Add(timeTolerance)

		//nolint:exhaustive
		switch fl.Field().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			timestamp = fl.Field().Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			timestamp = int64(fl.Field().Uint())
		case reflect.Float32, reflect.Float64:
			timestamp = int64(fl.Field().Float())
		default:
			log.Warn().Msg("timestamp_not_in_future validator is used on non-timestamp field")
		}

		return time.Unix(timestamp, 0).Before(now)
	})
	if err != nil {
		log.Err(err).Msg("failed to register timestamp_not_in_future validator")
	}

	return validator
}

func ConvertValidatorError(errors v.ValidationErrors) ValidationError {
	errorFields := make([]Field, 0, len(errors))

	for _, err := range errors {
		fieldName := stringy.New(err.StructNamespace()).SnakeCase(".", ".").ToLower()
		fieldName = fieldName[strings.Index(fieldName, ".")+1:]

		msg := ""
		switch err.Tag() {
		case "timestamp_not_in_future":
			msg = "timestamp should not be in the future"
		case "ltfield_if_set":
			msg = "should be less than " + stringy.New(err.Param()).SnakeCase().ToLower()
		case "gtfield_if_set":
			msg = "should be greater than " + stringy.New(err.Param()).SnakeCase().ToLower()
		case "oneof":
			msg = "should be one of: " + strings.ReplaceAll(err.Param(), " ", ", ")
		}

		errorFields = append(errorFields, Field{
			Field:   fieldName,
			Message: msg,
		})
	}

	return ValidationError{fields: errorFields}
}
