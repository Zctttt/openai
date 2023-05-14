package web

import (
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"go-chatgpt-api/config"
	"go-chatgpt-api/handlers"
	"net/http"
	"time"
)

func Init() {

	router := gin.Default()
	gin.ForceConsoleColor()
	router.Use(RateLimitMiddleware(time.Second, 5, 5))
	//// # Headers
	// Allow CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	})
	//// # Add routes

	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)
	//==============业务============
	//router.GET("/api/ask", handlers.Ask)
	//router.GET("/api/createImg", handlers.GenerateImg)
	//==============业务end============

	//==============工具start============
	//router.GET("/ip", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"data": ip.GetIpAddress(c.ClientIP()),
	//		"msg":  "success",
	//		"code": "200",
	//	})
	//})
	// Add a health endpoint
	//router.GET("/health", func(c *gin.Context) {
	//	c.String(http.StatusOK, "OK")
	//})
	//==============工具end============

	//router.LoadHTMLFiles("html/search.html")
	//获取form参数
	//router.GET("/index", func(c *gin.Context) {
	//	c.HTML(http.StatusOK, "search.html", nil)
	//})
	router.Use(JwtMiddleware())
	router.GET("/:auth/chat/info", handlers.Info)
	router.POST("/:auth/chat/check_token", handlers.CheckToken)
	router.POST("/:auth/chat/get_token", handlers.GetToken)
	//router.POST("/search", handlers.AskSearch)
	router.POST("/:auth/chat/completions", handlers.AskStream)
	//router.POST("/createImg", handlers.CreateImg)

	router.Run(fmt.Sprintf("%s:%s", *config.GetIp(), *config.GetPort()))
}

func RateLimitMiddleware(fillInterval time.Duration, cap, quantum int64) gin.HandlerFunc {
	bucket := ratelimit.NewBucketWithQuantum(fillInterval, cap, quantum)
	return func(c *gin.Context) {
		if bucket.TakeAvailable(1) < 1 {
			c.String(http.StatusForbidden, "请求人数过多请稍后再试..")
			c.Abort()
			return
		}
		c.Next()
	}
}

func JwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//token, err :=
		var Jwt struct {
			Auth string `uri:"auth"`
		}
		if err := c.BindUri(&Jwt); err != nil {
			return
		}
		c.Request.Header.Set("Authorization", Jwt.Auth)
		token, err := request.ParseFromRequest(c.Request, request.AuthorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(config.GetJwtSecret()), nil
			})
		if err == nil {
			if token.Valid {
				b, _ := json.Marshal(token.Claims)
				id, _ := jsonparser.GetInt(b, "id")
				c.Set("id", id)
				c.Next()
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"data": "Token is not valid",
					"msg":  "failed",
				})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"data": "Unauthorized access to this resource",
				"msg":  "failed",
			})
			c.Abort()
			return
		}
	}
}
