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
	r.GET("/ho", noop{})
	mount(r, "/")

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ho", nil)
	if err != nil {
		b.Fatal(err)
	}

	// b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkGetParams(b *testing.B) {
	_ = fmt.Print
	r := New()
	//r.GET("/ho/:hi/hu/:he", noop{})
	r.GET("/ho/:hi", noop{})
	/*
		r.GETFunc("/ho/:hi", func(rw http.ResponseWriter, req *http.Request) {
			fmt.Println(GetRouteParam(req, "hi"))
		})
	*/

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
