package routehtml

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-on/lib/html"

	"github.com/go-on/method"
	"github.com/go-on/router/route"
)

func TestRoutehtml(t *testing.T) {

	rt := route.New("/hello/:name", method.GET)

	// println(html.FORM("hi").String())

	route.Mount("/", rt)

	params := Params("name", "world")

	tests := map[http.Handler]string{
		AHref(&rt, params, "link"):                      "<a href=\"/hello/world\">link</a>",
		FORM(&rt, params, "form"):                       "<form action=\"/hello/world\">form</form>",
		FormGet(&rt, params, "formget"):                 "<form method=\"get\" action=\"/hello/world\">formget</form>",
		FormPost(&rt, params, "formpost"):               "<form method=\"post\" action=\"/hello/world\">formpost</form>",
		FormDelete(&rt, params, "formdelete"):           "<form method=\"post\" action=\"/hello/world\"><input name=\"_method\" value=\"DELETE\" type=\"hidden\" />formdelete</form>",
		FormPatch(&rt, params, "formpatch"):             "<form method=\"post\" action=\"/hello/world\"><input name=\"_method\" value=\"PATCH\" type=\"hidden\" />formpatch</form>",
		FormPut(&rt, params, "formput"):                 "<form method=\"post\" action=\"/hello/world\"><input name=\"_method\" value=\"PUT\" type=\"hidden\" />formput</form>",
		FormPostMultipart(&rt, params, "formpostmulti"): "<form method=\"post\" action=\"/hello/world\" enctype=\"multipart/form-data\">formpostmulti</form>",
		JsSrc(&rt, params):                              "<script type=\"text/javascript\" src=\"/hello/world\"></script>",
		CssHref(&rt, params, html.Media_("screen")):     "<link rel=\"stylesheet\" type=\"text/css\" href=\"/hello/world\" media=\"screen\" />",
		ImgSrc(&rt, params, html.Width_("200px")):       "<img src=\"/hello/world\" width=\"200px\" />",
	}

	for h, expected := range tests {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, nil)
		got := rec.Body.String()
		if got != expected {
			t.Errorf("got: %#v expected: %#v", got, expected)
		}
	}

}

func errorMustBe(err interface{}, class interface{}) string {
	classTy := reflect.TypeOf(class)
	if err == nil {
		return fmt.Sprintf("error must be of type %s but is nil", classTy)
	}

	errTy := reflect.TypeOf(err)
	if errTy.String() != classTy.String() {
		return fmt.Sprintf("error must be of type %s but is of type %s", classTy, errTy)
	}
	return ""
}

func TestParamsError(t *testing.T) {
	rt := route.New("/hello/:name", method.GET)

	// println(html.FORM("hi").String())

	route.Mount("/", rt)

	params := Params()
	tag := AHref(&rt, params)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, &ErrParams{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(*ErrParams)
		_ = err.Error()

		if err.Tag != tag {
			t.Errorf("wrong tag: %#v, expecting %#v", *err.Tag, *tag)
		}
	}()

	rec := httptest.NewRecorder()

	tag.ServeHTTP(rec, nil)
}

func TestRouteIsNilError(t *testing.T) {
	var rt *route.Route

	params := Params()
	tag := AHref(&rt, params)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, &ErrRouteIsNil{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(*ErrRouteIsNil)
		_ = err.Error()

		if err.Tag != tag {
			t.Errorf("wrong tag: %#v, expecting %#v", *err.Tag, *tag)
		}
	}()

	rec := httptest.NewRecorder()

	tag.ServeHTTP(rec, nil)
}

func TestTagAdd(t *testing.T) {
	var (
		rt       = route.New("/hello/:name", method.GET)
		tag      = AHref(&rt, Params("name", "world"), "link")
		expected = "<a href=\"/hello/world\">link-added</a>"
		rec      = httptest.NewRecorder()
	)

	route.Mount("/", rt)
	tag.Add("-added")
	tag.ServeHTTP(rec, nil)
	got := rec.Body.String()
	if got != expected {
		t.Errorf("got: %#v expected: %#v", got, expected)
	}
}

func TestErrorPairParams(t *testing.T) {

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, route.ErrPairParams{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(route.ErrPairParams)
		_ = err.Error()

	}()

	Params("one")
}
