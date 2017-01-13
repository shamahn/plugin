package gorillamux

import (
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/kataras/iris"
)

// order matters in gorilla mux
type bySubdomain []iris.Route

// Sorting happens when the mux's request handler initialized
func (s bySubdomain) Len() int {
	return len(s)
}
func (s bySubdomain) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s bySubdomain) Less(i, j int) bool {
	return len(s[i].Subdomain()) > len(s[j].Subdomain())
}

// New returns a new gorilla mux router which can be plugged inside iris.
// This is magic.
func New() iris.Plugin {
	return gorillaMux{}
}

// gorillaMux is the plugin which converts all Iris handlers and routes to have gorilla's regexp features
// and sets the end result of the *mux.Router to the iris.Router
type gorillaMux struct{}

// PreBuild state because on PreLookup the iris.UseGlobal/UseGlobalFunc may not be catched.
func (g gorillaMux) PreBuild(s *iris.Framework) {
	router := mux.NewRouter()
	routes := s.Lookups()
	// gorilla mux order matters, so order them by subdomain before looping
	sort.Sort(bySubdomain(routes))

	for i := range routes {
		route := routes[i]
		registerRoute(route, router, s)
	}
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := s.AcquireCtx(w, r)
		// to catch custom 404 not found http errors may registered by user
		ctx.EmitError(iris.StatusNotFound)
		s.ReleaseCtx(ctx)
	})
	s.Router = router
}

// so easy:
func registerRoute(route iris.Route, gorillaRouter *mux.Router, s *iris.Framework) {

	if route.IsOnline() {
		handler := func(w http.ResponseWriter, r *http.Request) {

			ctx := s.AcquireCtx(w, r)

			if params := mux.Vars(r); params != nil && len(params) > 0 {
				// set them with ctx.Set in order to be accesible by ctx.Param in the user's handler
				for k, v := range params {
					ctx.Set(k, v)
				}
			}
			// including the iris.Use/UseFunc and the route's middleware,
			// main handler and any done handlers.
			ctx.Middleware = route.Middleware()
			ctx.Do()

			s.ReleaseCtx(ctx)
		}
		// remember, we get a new iris.Route foreach of the HTTP Methods, so this should be work
		gorillaRoute := gorillaRouter.HandleFunc(route.Path(), handler).Methods(route.Method()).Name(route.Name())
		subdomain := route.Subdomain()
		if subdomain != "" {
			if subdomain == "*." {
				// it's an iris wildcard subdomain
				// so register it as wildcard on gorilla mux too (hopefuly, it supports these things)
				subdomain = "{subdomain}."
			} else {
				// it's a static subdomain (which contains the dot)
			}
			// host = subdomain  + listening host
			gorillaRoute.Host(subdomain + s.Config.VHost)
		}
	}

	// AUTHOR NOTE:
	// the only feature I can think right now that is missing is the
	// iris offline routing (which after a little research can be done with gorilla's mux BuildOnly
	// but I don't know how can I activate that again, I probably need to get its handler and execute dynamically
	// this will slow down the things, so I don't do it, until I think a better way,
	// offline routing is a 2-day feature so I suppose no many people know about that yet,
	// and if they want offline routing they should not change to a custom router (yet). so we are ok.
}
