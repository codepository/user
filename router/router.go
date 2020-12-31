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
	// ------------------ 导出数据 ------------------------
	Mux.HandleFunc("/api/v1/user/export", interceptor(controller.Export))
	// 导入
	Mux.HandleFunc("/api/v1/user/import", interceptor(controller.Import))
	// 存储文件到数据库
	Mux.HandleFunc("/api/v1/user/savefiletodb", interceptor(controller.SaveFileToDB))
	// 从数据库下载文件
	Mux.HandleFunc("/api/v1/user/downloadfilefromdb", interceptor(controller.DownloadFileFromDB))
	Mux.HandleFunc("/api/v1/user/getData", interceptor(controller.GetData))
	Mux.HandleFunc("/api/v1/user/login", interceptor(controller.Login))
	Mux.HandleFunc("/api/v1/user/logout", interceptor(controller.Logout))
	Mux.HandleFunc("/api/v1/user/checknew", interceptor(controller.CheckNewUser))
	Mux.HandleFunc("/api/v1/user/alive", interceptor(controller.Alive))
	Mux.HandleFunc("/user/permission/hasPermission", interceptor(controller.HasPermission))
}
