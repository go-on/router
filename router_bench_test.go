package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type noop struct{}

func (noop) ServeHTTP(http.ResponseWriter, *http.Request) {}

// Benchmark stdhandler without routing
func BenchmarkGetNoParams(b *testing.B) {
	r := New()
	r.GET("/hi/ho/hu", noop{})
	mount(r, "/he")

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/he/hi/ho/hu", nil)
	if err != nil {
		b.Fatal(err)
	}

	// b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkGet2Params(b *testing.B) {
	_ = fmt.Print
	r := New()
	r.GET("/ho/:hi/hu/:he", noop{})
	// r.GET("/ho/:hi", noop{})

	mount(r, "/")

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ho/hi/hu/he", nil)
	// req, err := http.NewRequest("GET", "/ho/hi", nil)
	if err != nil {
		b.Fatal(err)
	}

	// b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkGet4Params(b *testing.B) {
	_ = fmt.Print
	r := New()
	r.GET("/ho/:hi/hu/:he/:a/:b", noop{})
	// r.GET("/ho/:hi", noop{})

	mount(r, "/")

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ho/hi/hu/he/c/d", nil)
	// req, err := http.NewRequest("GET", "/ho/hi", nil)
	if err != nil {
		b.Fatal(err)
	}

	// b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkGet6Params(b *testing.B) {
	_ = fmt.Print
	r := New()
	r.GET("/ho/:hi/hu/:he/:a/:b/:x/:y", noop{})
	// r.GET("/ho/:hi", noop{})

	mount(r, "/")

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ho/hi/hu/he/c/d/e/f", nil)
	// req, err := http.NewRequest("GET", "/ho/hi", nil)
	if err != nil {
		b.Fatal(err)
	}

	// b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkGet1Param(b *testing.B) {
	_ = fmt.Print
	r := New()
	//r.GET("/ho/:hi/hu/:he", noop{})
	r.GET("/ho/:hi", noop{})

	mount(r, "/")

	rec := httptest.NewRecorder()
	//req, err := http.NewRequest("GET", "/ho/hi/hu/he", nil)
	req, err := http.NewRequest("GET", "/ho/hi", nil)
	if err != nil {
		b.Fatal(err)
	}

	// b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.ServeHTTP(rec, req)
	}
}

type fetchX struct{}

func (f fetchX) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(GetRouteParam(r, "x")))
}

func BenchmarkFetchParam(b *testing.B) {
	_ = fmt.Print
	r := New()
	r.GET("/ho/:x", fetchX{})
	mount(r, "/")

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ho/hi", nil)
	if err != nil {
		b.Fatal(err)
	}

	// b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.ServeHTTP(rec, req)
	}
}
