package main

import (
	"fmt"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/core"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/initialize"
)

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy
//go:generate go mod download

// @title                       Gin-Vue-Admin Swagger API接口文档
// @version                     v2.6.2
// @description                 使用gin+vue进行极速开发的全栈开发基础平台
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        x-token
// @BasePath                    /

// DirExists 检查指定的路径是否存在且为目录
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func RunFileServer(distDir string) {
	// 设置文件服务器的根目录
	time.Sleep(1 * time.Second)
	if !DirExists(distDir) {
		fmt.Println("web 前端目录不存在", distDir)
		return
	}

	// 后端服务的地址
	target := "http://localhost:8888"

	// 解析后端服务的URL
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("无法解析目标URL: %v", err)
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 代理服务器的处理器函数
	proxyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			//去掉/api前缀
			newPath := strings.TrimPrefix(r.URL.Path, "/api/")
			// 修改请求的URL
			r.URL.Path = newPath
			// 将请求转发到后端服务
			proxy.ServeHTTP(w, r)
		} else {
			// 处理其他请求，提供dist目录下的静态文件
			http.StripPrefix("", http.FileServer(http.Dir(distDir))).ServeHTTP(w, r)
		}
	})

	// 定义服务器监听的端口
	port := "8090"

	// 启动代理服务器
	log.Printf("Serving %s at http://localhost:%s\n", distDir, port)
	if err := http.ListenAndServe(":"+port, proxyHandler); err != nil {
		log.Fatal("启动代理服务器失败: ", err)
	}
}

func main() {
	global.GVA_VP = core.Viper() // 初始化Viper
	initialize.OtherInit()
	global.GVA_LOG = core.Zap() // 初始化zap日志库
	zap.ReplaceGlobals(global.GVA_LOG)
	global.GVA_DB = initialize.Gorm() // gorm连接数据库
	initialize.Timer()
	initialize.DBList()
	if global.GVA_DB != nil {
		initialize.RegisterTables() // 初始化表
		// 程序结束前关闭数据库链接
		db, _ := global.GVA_DB.DB()
		defer db.Close()
	}
	go RunFileServer("../web/dist")
	core.RunWindowsServer()
}
