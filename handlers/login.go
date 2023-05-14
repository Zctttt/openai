package handlers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go-chatgpt-api/config"
	"go-chatgpt-api/models"
	services "go-chatgpt-api/service"
	"time"
)

type UserReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func setToken(u models.UserClaims) string {
	u.Exp = time.Now().Add(time.Hour * time.Duration(1)).Unix()
	u.Iat = time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)

	tokenString, err := token.SignedString([]byte(config.GetJwtSecret()))
	if err != nil {
		return ""
	}
	return tokenString
}

func Register(c *gin.Context) {
	var (
		req UserReq
		err error
	)
	c.BindJSON(&req)
	if err = services.CreateUser(req.Username, req.Password); err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"msg":  err.Error(),
			"code": "500",
		})
		return
	}
	c.JSON(200, gin.H{
		"data": "ok",
		"msg":  "success",
		"code": "200",
	})
}

func Login(c *gin.Context) {
	var (
		req  UserReq
		user *models.MtUser
		err  error
	)
	c.BindJSON(&req)
	if user, err = services.FindUser(req.Username, req.Password); err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"msg":  err.Error(),
			"code": "500",
		})
		return
	}
	c.JSON(200, gin.H{
		"data": setToken(models.UserClaims{
			Id:             user.Id,
			Username:       user.NAME,
			Password:       user.PASSWORD,
			CreateTime:     user.CREATETIME,
			StandardClaims: jwt.StandardClaims{},
		}),
		"msg":  "success",
		"code": "200",
	})
}

func Info(c *gin.Context) {
	idstr, _ := c.Get("id")
	c.JSON(200, gin.H{
		"data": idstr.(int64),
		"msg":  "success",
		"code": "200",
	})
	return
}
