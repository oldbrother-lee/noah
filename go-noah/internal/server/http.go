package server

import (
	"go-noah/docs"
	"go-noah/internal/middleware"
	"go-noah/internal/router"
	"go-noah/pkg/global"
	"go-noah/pkg/jwt"
	"go-noah/pkg/log"
	"go-noah/pkg/server/http"
	nethttp "net/http"
	"os"

	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewHTTPServer(
	logger *log.Logger,
	conf *viper.Viper,
	jwt *jwt.JWT,
	e *casbin.SyncedEnforcer,
) *http.Server {
	gin.SetMode(gin.DebugMode)
	s := http.NewServer(
		gin.Default(),
		logger,
		http.WithServerHost(conf.GetString("http.host")),
		http.WithServerPort(conf.GetInt("http.port")),
	)
	// 设置前端静态资源（使用外部 web 目录）
	s.Use(static.Serve("/", static.LocalFile("../web/dist", true)))
	s.NoRoute(func(c *gin.Context) {
		indexPageData, err := os.ReadFile("../web/dist/index.html")
		if err != nil {
			c.String(nethttp.StatusNotFound, "404 page not found")
			return
		}
		c.Data(nethttp.StatusOK, "text/html; charset=utf-8", indexPageData)
	})
	// swagger doc
	docs.SwaggerInfo.BasePath = "/"
	s.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		//ginSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", conf.GetInt("app.http.port"))),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.PersistAuthorization(true),
	))

	s.Use(
		middleware.CORSMiddleware(),
		middleware.ResponseLogMiddleware(logger),
		middleware.RequestLogMiddleware(logger),
		//middleware.SignMiddleware(log),
	)

	// 使用 router 包注册路由
	router.InitRouter(s.Engine, jwt, e, logger)

	// 供 API 同步使用：对比代码路由与数据库（与 gin-vue-admin 的 GVA_ROUTERS 类似）
	global.Engine = s.Engine

	return s
}
