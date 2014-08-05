package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-on/router"
	"github.com/go-on/router/third-party/routergomniauth"
	"github.com/go-on/wrap"
	"github.com/go-on/wrap-contrib/wraps"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/common"
	"github.com/stretchr/signature"
)

// context is an example how a wrap.Contexter can be build in order to store the common.Provider, common.User and errors
type context struct {
	http.ResponseWriter
	provider common.Provider
	user     common.User
	err      error
}

var _ wrap.Contexter = &context{}

func (c *context) Context(ctxPtr interface{}) (found bool) {
	found = true
	switch ty := ctxPtr.(type) {
	case *error:
		if c.err == nil {
			return false
		}
		*ty = c.err
	case *common.Provider:
		if c.provider == nil {
			return false
		}
		*ty = c.provider
	case *common.User:
		if c.user == nil {
			return false
		}
		*ty = c.user
	default:
		panic(fmt.Sprintf("unsupported context: %T", ctxPtr))
	}
	return
}

func (c *context) SetContext(ctxPtr interface{}) {
	switch ty := ctxPtr.(type) {
	case *error:
		c.err = *ty
	case *common.Provider:
		c.provider = *ty
	case *common.User:
		c.user = *ty
	default:
		panic(fmt.Sprintf("unsupported context: %T", ctxPtr))
	}
}

// Wrap implements the wrap.Wrapper interface.
func (c context) Wrap(next http.Handler) http.Handler {
	var f http.HandlerFunc
	f = func(rw http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(&context{ResponseWriter: rw}, req)
	}
	return f
}

func login(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(`
<html>
	<body>
		<h2>Log in with...</h2>
		<ul>
			<li><a href="` + routergomniauth.LoginURL("github") + `">GitHub</a></li>
			<li><a href="` + routergomniauth.LoginURL("google") + `">Google</a></li>
			<li><a href="` + routergomniauth.LoginURL("facebook") + `">Facebook</a></li>
		</ul>
	</body>
</html>
`))
}

func handleError(rw http.ResponseWriter, req *http.Request) {
	var err error
	rw.(wrap.Contexter).Context(&err)
	rw.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(rw, "an error happened: %s", err.Error())
}

func catchPanics(p interface{}, rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(rw, "a panic happened: %v", p)
}

func authApp(rw http.ResponseWriter, req *http.Request) {
	var user common.User
	rw.(wrap.Contexter).Context(&user)
	fmt.Fprintf(rw, "email: %s name: %s", user.Email(), user.Name())
}

func main() {
	// has to be done once
	gomniauth.SetSecurityKey(signature.RandomKey(64))

	// first setup the context in an entouring wrapper/router
	mainRouter := router.New(
		context{},
		wraps.ErrorHandler(http.HandlerFunc(handleError)),
		wraps.CatchFunc(catchPanics),
	)

	mainRouter.GETFunc("/", login)

	// then setup the auth router
	authRouter := routergomniauth.Router(http.HandlerFunc(authApp))
	authRouter.Mount("/auth", mainRouter)

	// then mount your main router
	mainRouter.Mount("/", nil)

	// then setup the providers
	host := routergomniauth.NewHTTPHost("localhost", 8080)

	// you will have to setup the corresponding callback url at each provider
	routergomniauth.Github("3d1e6ba69036e0624b61", "7e8938928d802e7582908a5eadaaaf22d64babf1", host.CallbackURL("github"))
	routergomniauth.Google("1051709296778.apps.googleusercontent.com", "7oZxBGwpCI3UgFMgCq80Kx94", host.CallbackURL("google"))
	routergomniauth.FaceBook("537611606322077", "f9f4d77b3d3f4f5775369f5c9f88f65e", host.CallbackURL("facebook"))
	gomniauth.WithProviders(routergomniauth.Providers...)

	// and go
	log.Println("Starting...")
	fmt.Print("Gomniauth - Example web app\n")
	fmt.Print(" \n")
	fmt.Print("Starting go-on powered server...\n")

	err := http.ListenAndServe(":8080", mainRouter.ServingHandler())

	if err != nil {
		fmt.Println("can't listen to localhost:8080")
	}
}
