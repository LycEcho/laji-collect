package bootstrap

import (
	"lajiCollect/app/route"
	"lajiCollect/config"
	"lajiCollect/core"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/view"
	"os/exec"
	"runtime"
	"time"
)

type Bootstrap struct {
	Application *iris.Application
	Port        int
	LoggerLevel string
}

func New(port int, loggerLevel string) *Bootstrap {
	var bootstrap Bootstrap
	bootstrap.Application = iris.New()
	bootstrap.Port = port
	bootstrap.LoggerLevel = loggerLevel

	//crond
	go core.Crond()

	return &bootstrap
}
//加载中间件
func (bootstrap *Bootstrap) loadGlobalMiddleware() {
	bootstrap.Application.Use(recover.New())
}

//加载路由
func (bootstrap *Bootstrap) loadRoutes() {
	route.Register(bootstrap.Application)
}

//增加模板引擎
func (bootstrap *Bootstrap) loadEngineFuc(pugEngine *view.DjangoEngine){
	pugEngine.AddFunc("stampToDate", TimestampToDate)
}

func (bootstrap *Bootstrap) Serve() {
	bootstrap.Application.Logger().SetLevel(bootstrap.LoggerLevel)
	bootstrap.loadGlobalMiddleware()
	bootstrap.loadRoutes()

	pugEngine := iris.Django(fmt.Sprintf("%stemplate", config.ExecPath), ".html")
	if config.ServerConfig.Env == "development" {
		//测试环境下动态加载
		pugEngine.Reload(true)
	}

	bootstrap.loadEngineFuc(pugEngine)
	bootstrap.Application.RegisterView(pugEngine)
	go Open(fmt.Sprintf("http://127.0.0.1:%d", config.ServerConfig.Port))
	bootstrap.Application.Run(
		iris.Addr(fmt.Sprintf(":%d", bootstrap.Port)),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithoutBodyConsumptionOnUnmarshal,
	)
}

//时间戳格式化
func TimestampToDate(in uint, layout string) string {
	t := time.Unix(int64(in), 0)
	return t.Format(layout)
}

//运行起来自动打开浏览器
func Open(uri string) {
	time.Sleep(1 * time.Second)
	var commands = map[string]string{
		"windows": "cmd /c start",
		"darwin":  "open",
		"linux":   "xdg-open",
	}

	run, ok := commands[runtime.GOOS]
	if !ok {
		fmt.Println(fmt.Sprintf("请手动在浏览器中打开网址： %s", uri))
		return
	}

	cmd := exec.Command(run, uri)
	err := cmd.Start()
	if err != nil {
		fmt.Println(fmt.Sprintf("请手动在浏览器中打开网址： %s", uri))
	}
}
