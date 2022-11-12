package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	imageServer := http.FileServer(http.Dir("./images/"))
	router.Handler(http.MethodGet, "/images/*filepath", http.StripPrefix("/images", imageServer))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf)
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	router.Handler(http.MethodGet, "/admin/dashboard", dynamic.ThenFunc(app.dashboardTable))
	router.Handler(http.MethodGet, "/admin/createadmin", dynamic.ThenFunc(app.adminCreate))
	router.Handler(http.MethodPost, "/admin/createadmin", dynamic.ThenFunc(app.adminCreatePost))

	router.Handler(http.MethodGet, "/admin/login", dynamic.ThenFunc(app.adminLogin))
	router.Handler(http.MethodPost, "/admin/login", dynamic.ThenFunc(app.adminLoginPost))

	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodPost, "/admin/deleteuser/:id", dynamic.ThenFunc(app.DeleteUser))
	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(router)
}
