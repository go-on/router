package router

import (
	"net/http"
	"reflect"

	"github.com/go-on/meta"
)

type Vars struct {
	http.ResponseWriter
	v map[string]string
}

func (v *Vars) Get(key string) string {
	return v.v[key]
}

func (v *Vars) SetStruct(ptrToStruct interface{}, key string) error {
	stru, err := meta.StructByValue(reflect.ValueOf(ptrToStruct))
	if err != nil {
		return err
	}
	fn := func(f *meta.Field, tagVal string) {
		vv, has := v.v[tagVal]
		if has {
			f.Value.SetString(vv)
		}
	}
	stru.EachTag(key, fn)
	return nil
}

func (v *Vars) Has(key string) bool {
	_, has := v.v[key]
	return has
}
