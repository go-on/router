package routergomniauth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-on/router"
	"github.com/go-on/router/route"
	"github.com/go-on/wrap"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/common"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/gomniauth/providers/soundcloud"
	"github.com/stretchr/objx"
)

type _login struct {
	state   *common.State
	options objx.Map
}

func login(state *common.State, options objx.Map) http.Handler {
	if state == nil {
		state = gomniauth.NewState("after", "success")
	}
	return &_login{state: state, options: options}
}

func (l *_login) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var provider common.Provider
	rw.(wrap.Contexter).Context(&provider)
	authUrl, err := provider.GetBeginAuthURL(l.state, l.options)

	if err != nil {
		rw.(wrap.Contexter).SetContext(&err)
		return
	}

	http.Redirect(rw, req, authUrl, 302)
}

type callback struct{}

func (cb callback) Wrap(next http.Handler) http.Handler {
	var f http.HandlerFunc
	f = func(rw http.ResponseWriter, req *http.Request) {

		var provider common.Provider

		ctx := rw.(wrap.Contexter)

		if ctx.Context(&provider) {

			m, errMap := objx.FromURLQuery(req.URL.RawQuery)
			if errMap != nil {
				ctx.SetContext(&errMap)
				return
			}

			creds, err := provider.CompleteAuth(m)

			if err != nil {
				ctx.SetContext(&err)
				return
			}

			user, userErr := provider.GetUser(creds)

			if userErr != nil {
				ctx.SetContext(&userErr)
				return
			}

			ctx.SetContext(&user)
		}
		next.ServeHTTP(rw, req)
	}
	return f
}

var providers = map[string]struct{}{}

func addProvider(s string) {
	providers[s] = struct{}{}
}

type setProvider struct{}

func (s setProvider) Wrap(next http.Handler) http.Handler {
	var f http.HandlerFunc
	f = func(rw http.ResponseWriter, req *http.Request) {

		providerStr := router.GetRouteParam(req, "gomniauth_provider")

		if providerStr == "" {
			next.ServeHTTP(rw, req)
		}

		ctx := rw.(wrap.Contexter)

		if _, has := providers[providerStr]; !has {
			err := errors.New("unsupported provider: " + providerStr)
			ctx.SetContext(&err)
			return
		}

		provider, err := gomniauth.Provider(providerStr)

		if err != nil {
			ctx.SetContext(&err)
			return
		}
		ctx.SetContext(&provider)

		next.ServeHTTP(rw, req)
	}
	return f
}

var Providers = []common.Provider{}

var loginRoute *route.Route
var callbackRoute *route.Route

func LoginURL(provider string) string {
	if _, has := providers[provider]; !has {
		panic("prodiver " + provider + " is not registered")
	}
	return loginRoute.MustURL("gomniauth_provider", provider)
}

func Router(app http.Handler) *router.Router {

	authRouter := router.New(setProvider{})

	loginRoute = authRouter.GET("/:gomniauth_provider/login", login(nil, nil))

	callbackRoute = authRouter.GET("/:gomniauth_provider/callback",
		wrap.New(
			callback{},
			wrap.Handler(app),
		),
	)

	return authRouter
}

type host struct {
	https    bool
	port     int
	hostname string
}

func (h *host) prefix() string {
	scheme := "http"
	if h.https {
		scheme += "s"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, h.hostname, h.port)
}

func (h *host) CallbackURL(provider string) string {
	return h.prefix() + callbackRoute.MustURL("gomniauth_provider", provider)
}

func Github(clientId, clientSecret, callbackURL string) {
	Providers = append(Providers, github.New(clientId, clientSecret, callbackURL))
	addProvider("github")
}

func Google(clientId, clientSecret, callbackURL string) {
	Providers = append(Providers, google.New(clientId, clientSecret, callbackURL))
	addProvider("google")
}

func FaceBook(clientId, clientSecret, callbackURL string) {
	Providers = append(Providers, facebook.New(clientId, clientSecret, callbackURL))
	addProvider("facebook")
}

func SoundCloud(clientId, clientSecret, callbackURL string) {
	Providers = append(Providers, soundcloud.New(clientId, clientSecret, callbackURL))
	addProvider("soundcloud")
}

func NewHTTPHost(hostname string, port int) *host {
	return &host{port: port, hostname: hostname}
}

func NewHTTPSHost(hostname string, port int) *host {
	return &host{port: port, hostname: hostname, https: true}
}
