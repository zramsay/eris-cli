package util

import (
	"errors"
	"reflect"
)

var (
	ErrMergeParameters = errors.New("parameters are not pointers to struct")
)

// Merge merges maps and slices of base and over and overwrites other base fields.
// Base and over are pointers to structs. The result is stored in base.
// Merge returns ErrMergeParameters if either base or over are not
// pointers to structs.
func Merge(base, over interface{}) error {
	if err := checkStructsAreMergeable(base, over); err != nil {
		return err
	}

	baseFields := reflect.TypeOf(base).Elem().NumField()
	for i := 0; i < baseFields; i++ {
		a := reflect.ValueOf(base).Elem().Field(i)
		b := reflect.ValueOf(over).Elem().Field(i)

		switch a.Kind() {
		case reflect.Slice:
			if b.IsNil() {
				continue
			}

			if a.IsNil() {
				a.Set(b)
				continue
			}

			a.Set(reflect.AppendSlice(a, b))
		case reflect.Map:
			if b.IsNil() {
				continue
			}

			if a.IsNil() {
				a.Set(b)
				continue
			}

			for _, key := range b.MapKeys() {
				a.SetMapIndex(key, b.MapIndex(key))
			}
		default:
			// Don't overwrite with zero values (0, "", false).
			if b.Interface() == reflect.Zero(b.Type()).Interface() {
				continue
			}
			a.Set(b)
		}
	}
	return nil
}

func checkStructsAreMergeable(base, over interface{}) error {
	if base == nil || over == nil {
		return ErrMergeParameters
	}

	// If not pointers, it won't be possible to store the result in base.
	if reflect.ValueOf(base).Kind() != reflect.Ptr ||
		reflect.ValueOf(over).Kind() != reflect.Ptr {
		return ErrMergeParameters
	}

	// Not structs.
	if reflect.ValueOf(base).Elem().Kind() != reflect.Struct ||
		reflect.ValueOf(over).Elem().Kind() != reflect.Struct {
		return ErrMergeParameters
	}

	// Structs, but varying number of fields.
	if reflect.TypeOf(base).Elem().NumField() != reflect.TypeOf(over).Elem().NumField() {
		return ErrMergeParameters
	}

	return nil
}
