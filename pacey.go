package main

import (
	"net/http"
)

type App struct {
	routes map[string]map[string]http.HandlerFunc
}

func NewApp() *App {
	return &App{routes: make(map[string]map[string]http.HandlerFunc)}
}

func (a *App) addRoute(method, path string, handler http.HandlerFunc) {
	if a.routes[path] == nil {
		a.routes[path] = make(map[string]http.HandlerFunc)
	}
	a.routes[path][method] = handler
}

func (a *App) GET(path string, handler http.HandlerFunc) {
	a.addRoute("GET", path, handler)
}

func (a *App) POST(path string, handler http.HandlerFunc) {
	a.addRoute("POST", path, handler)
}
func (a *App) PUT(path string, handler http.HandlerFunc) {
	a.addRoute("PUT", path, handler)
}
func (a *App) PATCH(path string, handler http.HandlerFunc) {
	a.addRoute("PATCH", path, handler)
}
func (a *App) DELETE(path string, handler http.HandlerFunc) {
	a.addRoute("DELETE", path, handler)
}

func (a *App) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if rm, ok := a.routes[req.URL.Path]; ok {
		if handler, ok := rm[req.Method]; ok {
			handler(res, req)
			return
		}
	}
	http.Error(res, "Not found", http.StatusNotFound)
}
