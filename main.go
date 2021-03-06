package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/codepository/user/conmgr"
	"github.com/codepository/user/controller"
	"github.com/codepository/user/router"

	"github.com/codepository/user/config"
	"github.com/codepository/user/model"
)

var conf = *config.Config

func goMain() error {
	// 启动数据库连接
	model.StartDB()
	defer func() {
		model.StopDB()
	}()
	// 启动连接管理器
	conmgr.New()
	defer func() {
		conmgr.Conmgr.Stop()
	}()
	// 启动函数路由
	controller.SetRouters()
	// http 路由
	mux := router.Mux
	readTimeout, err := strconv.Atoi(conf.ReadTimeout)
	if err != nil {
		return err
	}
	writeTimeout, err := strconv.Atoi(conf.WriteTimeout)
	if err != nil {
		return err
	}

	// 监测内存

	isMemPprof, _ := strconv.ParseBool(conf.SaveHeapProfile)
	if isMemPprof {
		// go func() {
		// 	s, _ := strconv.Atoi(conf.SaveHeapProfileTimePeriod)
		// 	log.Printf("每%d秒生成一个内存使用情况图\n", s)
		// 	time.Sleep(time.Duration(s) * time.Second)
		// 	saveHeapProfile()
		// }()
		log.Printf("pprof监听6060端口,打开网址：http://localhost:6060/debug/pprof/  查看内存CPU使用情况\n")
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}

	// 创建server服务
	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", conf.Port),
		Handler:        mux,
		ReadTimeout:    time.Duration(readTimeout * int(time.Second)),
		WriteTimeout:   time.Duration(writeTimeout * int(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}
	// 监听关闭请求和关闭信号（Ctrl+C）
	interrupt := interruptListener(server)
	log.Printf("the application start up at port%s\n", server.Addr)
	if conf.TLSOpen == "true" {
		err = server.ListenAndServeTLS(conf.TLSCrt, conf.TLSKey)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil {
		log.Printf("Server err: %v", err)
		return err
	}
	<-interrupt
	return nil
}
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(10)
	if err := goMain(); err != nil {
		os.Exit(1)
	}
}
