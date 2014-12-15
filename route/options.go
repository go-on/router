package route

import "gopkg.in/go-on/method.v1"

// Options returns an array that may be used by handlers for the OPTIONS http method
func Options(r *Route) []string {
	allow := []string{method.OPTIONS.String()}

	if r.HasMethod(method.GET) {
		allow = append(allow, method.GET.String())
		allow = append(allow, method.HEAD.String())
	}

	if r.HasMethod(method.POST) {
		allow = append(allow, method.POST.String())
	}

	if r.HasMethod(method.DELETE) {
		allow = append(allow, method.DELETE.String())
	}

	if r.HasMethod(method.PATCH) {
		allow = append(allow, method.PATCH.String())
	}

	if r.HasMethod(method.PUT) {
		allow = append(allow, method.PUT.String())
	}

	return allow
}
