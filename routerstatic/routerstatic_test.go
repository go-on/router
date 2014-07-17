package routerstatic

import "testing"

func TestTransformLink(t *testing.T) {
	corpus := map[string]string{
		transformLink("http://abc.de"): "http://abc.de",
		transformLink("/abc.de"):       "/abc.de",
		transformLink("/abc"):          "/abc.html",
	}

	for got, expected := range corpus {
		if got != expected {
			t.Errorf("expected: %#v, got %#v", expected, got)
		}
	}
}
