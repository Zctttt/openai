package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pandodao/tokenizer-go"
	gogpt "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"go-chatgpt-api/cache"
	"go-chatgpt-api/config"
	"go-chatgpt-api/initialize/mysql"
	"go-chatgpt-api/models"
	"go-chatgpt-api/openai"
	services "go-chatgpt-api/service"
	"net/http"
	"strings"
	"time"
)

var askLogService services.AskLogService

func Ask(c *gin.Context) {
	secret := c.Query("secret")

	if secret == "" || secret != *config.GetRequestSecret() {
		c.JSON(500, gin.H{
			"error": "秘钥错误，拒绝访问",
			"code":  "500",
		})
		return
	}
	content := c.Query("content")
	result, err := requestOpenAIChat(c, content)
	if err != nil {
		c.JSON(500, gin.H{
			"msg":  "OpenAI官方系统繁忙，请稍后再试",
			"code": "500",
		})
		return
	}
	c.JSON(200, gin.H{
		"data": result,
		"msg":  "success",
		"code": "200",
	})
}

func requestOpenAIChat(c *gin.Context, content string) (string, error) {
	result, err := openai.Ask(content, c.ClientIP())
	if err != nil {
		return "", err
	}
	return result, nil
}

func requestOpenAICreateImg(c *gin.Context, content string) (string, error) {
	result, err := openai.GenerateImg(content, c.ClientIP())
	if err != nil {
		return "", err
	}
	return result, nil
}

func AskSearch(c *gin.Context) {
	var req struct {
		Question string `json:"question"`
	}
	content := c.BindJSON(&req)
	fmt.Println(req)
	var result string

	queryLog := askLogService.QueryRecentAsk(&models.AskLog{
		Request:   req.Question,
		Method:    "Ask",
		RequestIp: c.ClientIP(),
	})
	if queryLog.Content != "" && !strings.Contains(queryLog.Content, "error") {
		log.Printf("%s get data from db. data :%s\n", c.ClientIP(), queryLog.Content)
		//c.HTML(http.StatusOK, "search.html", gin.H{
		//	"data":    queryLog.Content,
		//	"content": content,
		//})
		fmt.Println("err", "question")
		c.JSON(200, map[string]interface{}{
			"data":    queryLog.Content,
			"content": req.Question,
		})
		return
	}
	cacheKey := fmt.Sprintf("ASK_%s-%s", c.ClientIP(), content)

	if _, found := cache.AskRequestLockCache.Get(cacheKey); found {
		c.HTML(http.StatusOK, "search.html", gin.H{
			"data":    "你的问题正在请求中,一会再来看看...",
			"content": content,
		})
		return
	}
	if req.Question != "" {
		var err error
		fmt.Println("find and set cache")
		cache.AskRequestLockCache.Set(cacheKey, 1, 10*60*time.Second)
		result, err = requestOpenAIChat(c, req.Question)
		fmt.Println(result)
		if err != nil {
			log.Println(err)
			c.HTML(http.StatusOK, "search.html", gin.H{
				"data":    "OpenAI官方系统繁忙，请稍后再试",
				"content": content,
			})
			return
		}
		cache.AskRequestLockCache.Delete(cacheKey)
	}
	c.JSON(200, map[string]interface{}{
		"data":    result,
		"content": req.Question,
	})
	//c.HTML(http.StatusOK, "search.html", gin.H{
	//	"data":    result,
	//	"content": content,
	//})
}

func CreateImg(c *gin.Context) {
	content := c.PostForm("createImgMsg")
	var url string
	if content != "" {
		cacheKey := fmt.Sprintf("CREATE_IMG-%s-%s", c.ClientIP(), content)
		if left, found := cache.AskRequestLockCache.Get(cacheKey); found {
			left := left.(*int)
			if *left > 0 {
				cache.AskRequestLockCache.Set(cacheKey, *left-1, 10*60*time.Second)
			} else if *left == 0 {
				c.HTML(http.StatusOK, "search.html", gin.H{
					"data":    "你的次数用完啦",
					"content": content,
				})
			}

		} else {
			cache.AskRequestLockCache.Set(cacheKey, 3, 10*60*time.Second)
		}
		var err error
		url, err = requestOpenAICreateImg(c, content)
		if err != nil {
			log.Println(err)
			c.HTML(http.StatusOK, "search.html", gin.H{
				"data":         "OpenAI官方系统繁忙，请稍后再试",
				"createImgMsg": content,
			})
			return
		}
	}
	c.HTML(http.StatusOK, "search.html", gin.H{
		"url":          url,
		"createImgMsg": content,
	})
}

func AskStream(c *gin.Context) {
	var req struct {
		Msg []gogpt.ChatCompletionMessage `json:"messages"`
	}
	c.BindJSON(&req)
	idstr, _ := c.Get("id")
	var points models.AskPoint
	if err := mysql.GetChatDb().Model(models.AskPoint{}).Where("user_id = ?", idstr.(int64)).Find(&points).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"data": points.Points,
			"msg":  "success",
			"code": 201,
		})
		return
	}
	if points.Points < 0 {
		c.JSON(http.StatusOK, gin.H{
			"data": "积分h",
			"msg":  "failed",
			"code": 201,
		})
		return
	}
	openai.AskStream(req.Msg, c.ClientIP(), c)
	//if err :!= nil {
	//	log.Println(err)
	//	c.JSON(500, gin.H{
	//		"error": "系统繁忙",
	//	})
	//	return
	//}
	//c.JSON(200, gin.H{
	//	"data": 1,
	//})
}

func GenerateImg(c *gin.Context) {
	content := c.Query("content")

	url, err := requestOpenAICreateImg(c, content)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"error": "OpenAI官方系统繁忙，请稍后再试",
		})
		return
	}
	c.JSON(200, gin.H{
		"data": url,
	})
}

func CheckToken(c *gin.Context) {
	var req struct {
		Msg []gogpt.ChatCompletionMessage `json:"messages"`
	}
	c.BindJSON(&req)
	var num int
	for _, v := range req.Msg {
		num += tokenizer.MustCalToken(v.Content)
	}

	c.JSON(200, gin.H{
		"data": num,
		"msg":  "success",
		"code": http.StatusOK,
	})
}

func GetToken(c *gin.Context) {
	idstr, _ := c.Get("id")
	var points models.AskPoint
	if err := mysql.GetChatDb().Model(models.AskPoint{}).Where("user_id = ?", idstr.(int64)).Find(&points).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"data": points.Points,
			"msg":  "success",
			"code": 201,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data": points.Points,
		"msg":  "success",
		"code": http.StatusOK,
	})
}
