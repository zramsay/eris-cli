package util

import (
	"errors"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"

	dirs "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

var (
	ErrMergeParameters = errors.New("parameters are not pointers to struct")
)

// Merge merges maps and slices of base and over and overwrites other base fields.
// Base and over are pointers to structs. The result is stored in base.
// Merge returns ErrMergeParameters if either base or over are not
// pointers to structs.
func Merge(base, over interface{}) error {
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
	baseFields := reflect.TypeOf(base).Elem().NumField()
	overFields := reflect.TypeOf(over).Elem().NumField()
	if baseFields != overFields {
		return ErrMergeParameters
	}

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

// AutoMagic will return the highest container number which would represent the most recent
// container to work on unless newCont == true in which case it would return the highest
// container number plus one.
func AutoMagic(cNum int, typ string, newCont bool) int {
	logger.Debugf("Automagic (base) =>\t\t%s:%d\n", typ, cNum)
	contns := ErisContainersByType(typ, true)

	contnums := make([]int, len(contns))
	for i, c := range contns {
		contnums[i] = c.Number
	}

	// get highest container number
	g := 0
	for _, n := range contnums {
		if n >= g {
			g = n
		}
	}

	// ensure outcomes appropriate
	result := g
	if newCont {
		result = g + 1
	}
	if result == 0 {
		result = 1
	}

	logger.Debugf("Automagic (result) =>\t\t%s:%d\n", typ, result)
	return result
}

// Parse dependencies for internalName (ie. in /etc/hosts), whether to link, and whether to mount volumes-from
// eg. `tinydns:tiny:l` to link to the tinydns container as tiny
func ParseDependency(nameAndOpts string) (name, internalName string, link, mount bool) {
	spl := strings.Split(nameAndOpts, ":")
	name = spl[0]

	// defaults
	link, mount = true, true
	internalName = name

	if len(spl) > 1 {
		if spl[1] != "" {
			internalName = spl[1]
		}
	}
	if len(spl) > 2 {
		switch spl[2] {
		case "l": // link
			link, mount = true, false
		case "m", "v": // mount/volumes-from
			link, mount = false, true
		case "_": // nothing!
			link, mount = false, false
		}
	}
	return
}

// $(pwd) doesn't execute properly in golangs subshells; replace it
// use $eris as a shortcut
func FixDirs(arg []string) ([]string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return []string{}, err
	}

	for n, a := range arg {
		if strings.Contains(a, "$eris") {
			tmp := strings.Split(a, ":")[0]
			keep := strings.Replace(a, tmp+":", "", 1)
			if runtime.GOOS == "windows" {
				winTmp := strings.Split(tmp, "/")
				tmp = path.Join(winTmp...)
			}
			tmp = strings.Replace(tmp, "$eris", dirs.ErisRoot, 1)
			arg[n] = strings.Join([]string{tmp, keep}, ":")
			continue
		}

		if strings.Contains(a, "$pwd") {
			arg[n] = strings.Replace(a, "$pwd", dir, 1)
		}
	}

	return arg, nil
}
