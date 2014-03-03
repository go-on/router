package routerfat

import (
	"reflect"
	. "github.com/go-on/fat"
	"github.com/go-on/meta"
	"github.com/go-on/router"
)

var strTy = reflect.TypeOf("")

func Url(rt *router.Route, øfatstruct interface{}, tag string) (string, error) {
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

func MustUrl(rt *router.Route, øfatstruct interface{}, tag string) string {
	u, err := Url(rt, øfatstruct, tag)
	if err != nil {
		panic(err.Error())
	}
	return u
}

func Set(vars *router.Vars, ptrToStruct interface{}, key string) (err error) {
	var stru *meta.Struct
	stru, err = meta.StructByValue(reflect.ValueOf(ptrToStruct))
	if err != nil {
		return
	}
	fn := func(f *meta.Field, tagVal string) {
		if err != nil {
			return
		}
		if vars.Has(tagVal) {
			vv := vars.Get(tagVal)
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

func MustSet(vars *router.Vars, ptrToStruct interface{}, key string) {
	err := Set(vars, ptrToStruct, key)
	if err != nil {
		panic(err.Error())
	}
}
