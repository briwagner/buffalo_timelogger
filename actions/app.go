package actions

import (
	"log"
	"reflect"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/events"
	forcessl "github.com/gobuffalo/mw-forcessl"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/unrolled/secure"

	"buftester/models"

	"github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
	csrf "github.com/gobuffalo/mw-csrf"
	i18n "github.com/gobuffalo/mw-i18n"
	"github.com/gobuffalo/packr/v2"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_buftester_session",
		})

		// Automatically redirect to SSL
		app.Use(forceSSL())

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))

		// Setup and use translations:
		app.Use(translations())
		app.Use(SetCurrentUser)

		// Prefer using groups to enable middleware that way.
		// app.Use(Authorize)
		// app.Middleware.Skip(Authorize, HomeHandler, UsersNew, UsersCreate, AuthNew, AuthCreate)

		app.GET("/", HomeHandler)

		app.GET("/signin", AnonOnly(AuthNew))
		app.POST("/signin", AuthCreate)
		app.DELETE("/signout", AuthDestroy)

		app.GET("/users", AnonOnly(UsersNew))
		app.POST("/users", UsersCreate)

		app.GET("/users/{user_id}", Authorize(IsOwner(UsersShow)))

		c := app.Group("/users")
		c.POST("/{user_id}", UsersUpdate)
		c.GET("/{user_id}/contracts", UsersContractsIndex)
		c.POST("/{user_id}/contracts", UsersContractCreate)
		c.GET("/{user_id}/contracts/new", UsersContractsNew)
		c.GET("/{user_id}/contracts/{contract_id}", UsersContractShow)
		c.Use(Authorize)

		b := app.Group("/bosses")
		b.GET("/index", BossesIndex)
		b.GET("/new", BossesNew)
		b.POST("/create", BossesCreate)
		b.GET("/{boss_id}", BossesShow)
		b.Use(Authorize)

		app.POST("/users/{user_id}/contracts{contract_id}/task/create", Authorize(UserTaskCreate))

		t := app.Group("/tasks")
		t.GET("/{task_id}", TasksShow)
		t.GET("/{task_id}/edit", TasksEdit)
		t.POST("/{task_id}/edit", TasksUpdate)
		t.Use(Authorize)

		admin := app.Group("/admin")
		admin.GET("/users", AdminUsersIndex)
		admin.GET("/users/{user_id}", AdminUserShow)
		admin.POST("/users/{user_id}", AdminUserUpdate)
		admin.Use(Authorize)
		admin.Use(IsAdmin)
		admin.Use(SetCurrentUser)

		app.ServeFiles("/", assetsBox) // serve files from the public directory
	}

	return app
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(packr.New("app:locales", "../locales"), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}

// Register event listeners.
func init() {
	events.Listen(func(e events.Event) {
		if e.Kind == buffalo.EvtRouteStarted {
			route, err := e.Payload.Pluck("route")
			if err != nil {
				return
			}
			// Assert to proper type, else it's viewed as interface{}.
			r, ok := route.(buffalo.RouteInfo)
			if !ok {
				return
			}
			if r.PathName == "newBossesPath" {
				log.Printf("User hitting the Bosses Create path %s", r.PathName)
			}
		}

		// Context will be present here.
		if e.Kind == buffalo.EvtRouteFinished {
			ctx, err := e.Payload.Pluck("context")
			if err != nil {
				return
			}
			context := ctx.(buffalo.Context)
			u := context.Value("current_user")
			if u != nil {
				user, ok := u.(*models.User)
				if !ok {
					log.Printf("Cannot convert %s\n", reflect.TypeOf(u).String())
					return
				}
				route, err := e.Payload.Pluck("route")
				if err != nil {
					return
				}
				r := route.(buffalo.RouteInfo)
				log.Printf("%s route finished %+v\n", r.PathName, user.Email)
			}
		}

		// Custom event.
		if e.Kind == "buftester:user:create" {
			log.Printf("New User: %s", e.Message)
		}
	})
}
