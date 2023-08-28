package pacey

import (
	"fmt"
	"net/http"
	"strconv"
)

type App interface {
	GET(path string, handler http.HandlerFunc)
	POST(path string, handler http.HandlerFunc)
	PUT(path string, handler http.HandlerFunc)
	PATCH(path string, handler http.HandlerFunc)
	DELETE(path string, handler http.HandlerFunc)
	GoLive(port int)
}

type appImpl struct {
	routes map[string]map[string]http.HandlerFunc
}

func NewApp() App {
	return &appImpl{routes: make(map[string]map[string]http.HandlerFunc)}
}

func (a *appImpl) addRoute(method, path string, handler http.HandlerFunc) {
	if a.routes[path] == nil {
		a.routes[path] = make(map[string]http.HandlerFunc)
	}
	a.routes[path][method] = handler
}

func (a *appImpl) GET(path string, handler http.HandlerFunc) {
	a.addRoute("GET", path, handler)
}
func (a *appImpl) POST(path string, handler http.HandlerFunc) {
	a.addRoute("POST", path, handler)
}
func (a *appImpl) PATCH(path string, handler http.HandlerFunc) {
	a.addRoute("PATCH", path, handler)
}
func (a *appImpl) PUT(path string, handler http.HandlerFunc) {
	a.addRoute("PATCH", path, handler)
}
func (a *appImpl) DELETE(path string, handler http.HandlerFunc) {
	a.addRoute("PATCH", path, handler)
}

func (a *appImpl) GoLive(port int) {
	fmt.Println("Server is live on PORT: " + strconv.Itoa(port))
	http.ListenAndServe(":"+strconv.Itoa(port), a)
}

func (a *appImpl) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if rm, ok := a.routes[req.URL.Path]; ok {
		if handler, ok := rm[req.Method]; ok {
			handler(res, req)
			return
		}
	}
	http.Error(res, "Not found", http.StatusNotFound)
}
