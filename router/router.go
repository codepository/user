package router

import (
	"net/http"

	"github.com/codepository/user/config"
	"github.com/codepository/user/controller"
)

// Mux 路由
var Mux = http.NewServeMux()
var conf = *config.Config

func interceptor(h http.HandlerFunc) http.HandlerFunc {
	return crossOrigin(h)
}
func crossOrigin(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", conf.AccessControlAllowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", conf.AccessControlAllowMethods)
		w.Header().Set("Access-Control-Allow-Headers", conf.AccessControlAllowHeaders)
		h(w, r)
	}
}
func init() {
	setMux()
}
func setMux() {
	Mux.HandleFunc("/api/v1/test/index", interceptor(controller.Index))
	Mux.HandleFunc("/api/v1/user/getData", interceptor(controller.GetData))
	Mux.HandleFunc("/api/v1/user/login", interceptor(controller.Login))
	Mux.HandleFunc("/api/v1/user/logout", interceptor(controller.Logout))
	Mux.HandleFunc("/api/v1/user/checknew", interceptor(controller.CheckNewUser))
	Mux.HandleFunc("/api/v1/user/alive", interceptor(controller.Alive))
	Mux.HandleFunc("/user/permission/hasPermission", interceptor(controller.HasPermission))
}
