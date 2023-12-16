package repository

import (
	"github.com/sashabaranov/go-openai"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/registry"
)

type (
	chatGPT struct {
		helper.SkeletonWithParameter
	}

	chatGPTParam struct {
		Token    string `json:"token"`
		ApiUrl   string `json:"api_url"`
		Question string `json:"question"`
	}
)

func registerChatGPT() {
	registry.Register("ChatGPT", "ChatGPT",
		func(id string) starriver.Executable {
			return &chatGPT{helper.NewSkeletonWithParameter(id, chatGPTParam{})}

		},
		registry.Input([]starriver.InputParam{
			{
				Key:      "Token",
				Required: true,
				Desc:     "ApiToken",
			},
			{
				Key:      "ApiUrl",
				Required: false,
				Desc:     "ApiUrl, instead of 'https://api.openai.com/v1'",
			},
			{
				Key:      "Question",
				Required: true,
				Desc:     "提问",
			},
		}),
		registry.Output(map[string]starriver.OutputValue{
			"Answer": {
				Desc: "执行结果",
			},
		}),
	)
}
func (c *chatGPT) Execute(dataContext starriver.DataContext, param interface{}) starriver.Response {
	cp := param.(*chatGPTParam)
	config := openai.DefaultConfig(cp.Token)
	if cp.ApiUrl != "" {
		config.BaseURL = cp.ApiUrl
	}
	client := openai.NewClientWithConfig(config)
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Content: cp.Question,
				Role:    "user",
			},
		},
	}
	resp, err := client.CreateChatCompletion(dataContext, req)
	if err != nil {
		return helper.NewErrorResponse(err)
	}
	return helper.NewSuccessDataResponse(map[string]interface{}{
		"Answer": resp.Choices[0].Message.Content,
	})

}
