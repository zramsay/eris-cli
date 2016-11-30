package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
)

//-------------------------------------------------------
// reflection and json

func WriteJson(config interface{}, config_file string) error {
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(config_file, out.Bytes(), 0600)
	return err
}

func ReadJson(config interface{}, config_file string) error {
	b, err := ioutil.ReadFile(config_file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, config)
	if err != nil {
		fmt.Println("error unmarshalling config from file:", err)
		return err
	}
	return nil
}

func NewInvalidKindErr(kind, k reflect.Kind) error {
	return fmt.Errorf("Invalid kind. Expected %s, received %s", kind, k)
}

func FieldFromTag(v reflect.Value, field string) (string, error) {
	iv := v.Interface()
	st := reflect.TypeOf(iv)
	for i := 0; i < v.NumField(); i++ {
		tag := st.Field(i).Tag.Get("json")
		if tag == field {
			return st.Field(i).Name, nil
		}
	}
	return "", fmt.Errorf("Invalid field name")
}

// Set a field in a struct value
// Field can be field name or json tag name
// Values can be strings that can be cast to int or bool
//  only handles strings, ints, bool
func SetProperty(cv reflect.Value, field string, value interface{}) error {
	f := cv.FieldByName(field)
	if !f.IsValid() {
		name, err := FieldFromTag(cv, field)
		if err != nil {
			return err
		}
		f = cv.FieldByName(name)
	}
	kind := f.Kind()

	k := reflect.ValueOf(value).Kind()
	if k != kind && k != reflect.String {
		return NewInvalidKindErr(kind, k)
	}

	if kind == reflect.String {
		f.SetString(value.(string))
	} else if kind == reflect.Int {
		if k != kind {
			v, err := strconv.Atoi(value.(string))
			if err != nil {
				return err
			}
			f.SetInt(int64(v))
		} else {
			f.SetInt(int64(value.(int)))
		}
	} else if kind == reflect.Bool {
		if k != kind {
			v, err := strconv.ParseBool(value.(string))
			if err != nil {
				return err
			}
			f.SetBool(v)
		} else {
			f.SetBool(value.(bool))
		}
	}
	return nil
}
