package routerfat

import (
	"reflect"

	. "github.com/go-on/lib/internal/fat"
	"github.com/go-on/lib/internal/meta"
	"github.com/go-on/router/route"
)

var strTy = reflect.TypeOf("")

func Url(rt *route.Route, øfatstruct interface{}, tag string) (string, error) {
	val := reflect.ValueOf(øfatstruct)
	params := map[string]string{}
	stru, err := meta.StructByValue(val)
	if err != nil {
		return "", err
	}

	fn := func(field *meta.Field, tagVal string) {
		fatfld, isFat := field.Value.Interface().(*Field)
		if isFat {
			params[tagVal] = fatfld.String()
		} else {
			params[tagVal] = field.Value.Convert(strTy).String()
		}
	}
	stru.EachTag(tag, fn)
	return rt.URLMap(params)
}

func MustUrl(rt *route.Route, øfatstruct interface{}, tag string) string {
	u, err := Url(rt, øfatstruct, tag)
	if err != nil {
		panic(err.Error())
	}
	return u
}

func Set(vars map[string]string, ptrToStruct interface{}, key string) (err error) {
	var stru *meta.Struct
	stru, err = meta.StructByValue(reflect.ValueOf(ptrToStruct))
	if err != nil {
		return
	}
	fn := func(f *meta.Field, tagVal string) {
		if err != nil {
			return
		}
		if vv, has := vars[tagVal]; has {
			// vv := vars.Get(tagVal)
			fatfld, isFat := f.Value.Interface().(*Field)
			if isFat {
				err = fatfld.ScanString(vv)
			} else {
				f.Value.SetString(vv)
			}

		}
	}
	stru.EachTag(key, fn)
	return
}

func MustSet(vars map[string]string, ptrToStruct interface{}, key string) {
	err := Set(vars, ptrToStruct, key)
	if err != nil {
		panic(err.Error())
	}
}
