package config

import (
	"errors"
	"log"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

var logger = log.New(os.Stdout, "[config] ", log.LstdFlags|log.Lshortfile)

var ErrInvalidConfigVariable = errors.New("Invalid config variable")

type ConfigLocation interface {
	GetString(string) string
	GetInt(string) (int, error)
}

type EnvConfig struct{}

func (e *EnvConfig) value(key string) string {
	key = strings.ToUpper(key)
	return os.Getenv(key)
}

func (e *EnvConfig) GetString(key string) string {
	return e.value(key)
}

func (e *EnvConfig) GetInt(key string) (int, error) {
	value := e.value(key)
	return strconv.Atoi(value)
}

func ParseFromLocation(config interface{}, location ConfigLocation) (err error) {
	// handle nasty panics and return them as an error instead
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	reflectValue := reflect.ValueOf(config)
	valueElem := reflectValue.Elem()

	reflectType := reflect.TypeOf(config)
	typeElem := reflectType.Elem()

	if valueElem.Kind() != reflect.Struct || typeElem.Kind() != reflect.Struct {
		return ErrInvalidConfigVariable
	}

	for i := 0; i < typeElem.NumField(); i++ {
		field := typeElem.Field(i)
		rawField := valueElem.Field(i)

		if !rawField.CanSet() {
			logger.Printf("skipping unassignable field=%q", field.Name)
			continue
		}

		defaultValue := field.Tag.Get("default")

		switch rawField.Kind() {
		case reflect.String:
			configValue := location.GetString(field.Name)
			if configValue == "" {
				configValue = defaultValue
			}
			logger.Printf("setting field %q to value %q", field.Name, configValue)

			rawField.SetString(configValue)
		case reflect.Int:
			configValue, err := location.GetInt(field.Name)
			if err != nil {
				// ignore error here, just take zero value
				n, _ := strconv.Atoi(defaultValue)
				configValue = n
			}
			logger.Printf("setting field %q to value %d", field.Name, configValue)

			rawField.SetInt(int64(configValue))

		default:
			logger.Printf("field %q type %v not supported", field.Name, rawField.Kind())
		}

	}
	return nil
}

func Parse(config interface{}) (err error) {
	return ParseFromLocation(config, &EnvConfig{})
}
