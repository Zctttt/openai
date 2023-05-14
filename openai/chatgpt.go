package openai

import (
	"context"
	"errors"
	"fmt"
	"github.com/pandodao/tokenizer-go"
	"go-chatgpt-api/cache"
	"go-chatgpt-api/config"
	"go-chatgpt-api/initialize/mysql"
	"go-chatgpt-api/models"
	services "go-chatgpt-api/service"
	"go-chatgpt-api/utils/ip"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	gogpt "github.com/sashabaranov/go-openai"
)

var askLogService services.AskLogService

// ChatGPTResponseBody 请求体
type ChatGPTResponseBody struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int                      `json:"created"`
	Model   string                   `json:"model"`
	Choices []map[string]interface{} `json:"choices"`
	Usage   map[string]interface{}   `json:"usage"`
}

// ChatGPTRequestBody 响应体
type ChatGPTRequestBody struct {
	Model            string  `json:"model"`
	Prompt           string  `json:"prompt"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float32 `json:"temperature"`
	TopP             int     `json:"top_p"`
	FrequencyPenalty int     `json:"frequency_penalty"`
	PresencePenalty  int     `json:"presence_penalty"`
}

func Ask(msg, ipParam string) (string, error) {
	log.Printf("ask request content:%s", msg)

	openAiClient := GetOpenAIClientFromCache(ipParam)

	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:     gogpt.GPT3TextDavinci003,
		MaxTokens: 4000,
		Prompt:    msg,
	}
	resp, err := openAiClient.CreateCompletion(ctx, req)
	if err != nil {
		go askLogService.Record(&models.AskLog{
			UserId:     0,
			Request:    msg,
			Content:    err.Error(),
			Method:     "Ask",
			CreateTime: time.Now(),
			RequestIp:  ipParam,
			Address:    *ip.GetIpAddress(ipParam),
		})
		return "", err
	}
	result := resp.Choices[0].Text
	go askLogService.Record(&models.AskLog{
		UserId:     0,
		Request:    msg,
		Content:    result,
		Method:     "Ask",
		CreateTime: time.Now(),
		RequestIp:  ipParam,
		Address:    *ip.GetIpAddress(ipParam),
	})
	fmt.Println(result)
	return result, nil
}

func GetOpenAIClientFromCache(ip string) *gogpt.Client {
	var openAiClient *gogpt.Client
	if x, found := cache.OpenAiClientCache.Get(ip); found {
		openAiClient = x.(*gogpt.Client)
		log.Printf("ip:%s openAIclient from cache", ip)
	} else {
		apiKey := config.GetOpenAiApiKey()
		if apiKey == nil {
			panic("未配置apiKey")
		}
		openAiClient = gogpt.NewClient(*apiKey)
		cache.OpenAiClientCache.Set(ip, openAiClient, 600*time.Second)
		log.Printf("ip:%s create new openAIclient", ip)
	}
	return openAiClient
}

func AskStream(msg []gogpt.ChatCompletionMessage, ip string, c *gin.Context) {
	// ctx := context.Background()
	var token = 0
	for _, v := range msg {
		token += tokenizer.MustCalToken(v.Content)
	}
	flusher, _ := c.Writer.(http.Flusher)
	openAiClient := GetOpenAIClientFromCache(ip)
	req := gogpt.ChatCompletionRequest{
		Model: gogpt.GPT3Dot5Turbo0301,
		//MaxTokens: 4097,
		Messages: msg,
		Stream:   true,
	}
	stream, err := openAiClient.CreateChatCompletionStream(c, req)
	if err != nil {
		fmt.Println(err)

		return
	}
	defer func() {
		c.SSEvent("message", map[string]interface{}{
			"status": `[done]`,
		})
		flusher.Flush()
		stream.Close()
	}()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream finished")
			idstr, _ := c.Get("id")
			var points = new(models.AskPoint)
			if err = mysql.GetChatDb().Model(models.AskPoint{}).Where("user_id = ?", idstr.(int64)).Find(points).Error; err != nil {
				fmt.Println(err)
				return
			}
			points.Points -= int64(token)
			if err = mysql.GetChatDb().Model(models.AskPoint{}).Where("id = ?", points.Id).Update("points", points.Points).Error; err != nil {
				fmt.Println(err)
				return
			}
			return
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return
		}

		fmt.Printf("Stream response: %v\n", response)

		for _, v := range response.Choices {
			c.SSEvent("message", map[string]interface{}{
				"id":      response.ID,
				"object":  response.Object,
				"created": response.Created,
				"model":   response.Model,
				"choice":  []openai.ChatCompletionStreamChoice{v},
			})
			flusher.Flush()
		}

	}

}

func GenerateImg(msg, ipParam string) (string, error) {
	openAiClient := GetOpenAIClientFromCache(ipParam)
	ctx := context.Background()
	req := gogpt.ImageRequest{
		Prompt:         msg,
		ResponseFormat: "url",
		Size:           "512x512",
	}
	resp, err := openAiClient.CreateImage(ctx, req)
	if err != nil {
		go askLogService.Record(&models.AskLog{
			UserId:     0,
			Request:    msg,
			Method:     "GenerateImg",
			Content:    err.Error(),
			CreateTime: time.Now(),
			RequestIp:  ipParam,
			Address:    *ip.GetIpAddress(ipParam),
		})
		return "", err
	}

	go askLogService.Record(&models.AskLog{
		UserId:     0,
		Request:    msg,
		Method:     "GenerateImg",
		Content:    resp.Data[0].URL,
		CreateTime: time.Now(),
		RequestIp:  ipParam,
		Address:    *ip.GetIpAddress(ipParam),
	})

	return resp.Data[0].URL, nil
}
