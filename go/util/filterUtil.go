package util

import (
	"errors"
	"fmt"
	"github.com/Baidu-AIP/golang-sdk/aip/censor"
	"github.com/bytedance/sonic"
	"github.com/importcjj/sensitive"
	"strings"
)

var filter = sensitive.New()
var clinet = censor.NewClient("6MxNPX8BFZZtxLGXSbhB05Rl", "zKi5bX4rhkHlrHwDj6qCZvBUOSXzDej7")

type ValidationResponse struct {
	Pass  bool
	Words []any
	Tips  string
}

func init() {
	filter.LoadWordDict("dict.txt") // 抛出panic
}

// LocalValidation 本地验证
func LocalValidation(text string) (ValidationResponse, error) {
	words := filter.FindAll(text)
	var res ValidationResponse
	if len(words) == 0 {
		res.Pass = true
	} else {
		res.Pass = false
		// 将字符串切片转换为接口切片 []interface{}
		interfaceSlice := make([]any, len(words))
		for i, v := range words {
			interfaceSlice[i] = v
		}
		res.Words = interfaceSlice
		res.Tips = "存在敏感词汇不合规"
	}
	return res, nil
}

// OnlineValidation 网络模型验证
func OnlineValidation(text string) (ValidationResponse, error) {
	var res ValidationResponse
	var onlineRes map[string]any
	err := sonic.UnmarshalString(clinet.TextCensor(text), &onlineRes)
	if err != nil {
		return ValidationResponse{}, err
	}

	fmt.Println(onlineRes)

	if int(onlineRes["conclusionType"].(float64)) != 1 {
		res.Pass = false
		datas := onlineRes["data"].([]any)
		data := datas[0].(map[string]any)
		res.Tips = data["msg"].(string)
		hits := data["hits"].([]any)
		hit := hits[0].(map[string]any)
		res.Words = hit["words"].([]any)
	} else {
		res.Pass = true
	}
	return res, nil
}

// DoubleValidation 双重验证
func DoubleValidation(text string) (ValidationResponse, error) {
	if strings.TrimSpace(text) == "" {
		return ValidationResponse{}, errors.New("文本内容为空")
	}

	res, err := LocalValidation(text)
	if err != nil {
		return ValidationResponse{}, err
	}
	if !res.Pass {
		return res, nil
	}
	res, err = OnlineValidation(text)
	if err != nil {
		return ValidationResponse{}, err
	}
	return res, nil
}
