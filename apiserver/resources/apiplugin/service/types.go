package service

import (
	"encoding/json"
	"fmt"
	"k8s.io/klog"
	"strings"
	"time"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/auth/cas"
)

type ApiPlugin struct {
	// basic information
	Id          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Namespace   string      `json:"namespace"`
	User        string      `json:"user"`
	UserName    string      `json:"userName"`
	Description string      `json:"description"`
	ConsumerId  string      `json:"consumerId"`
	Config      interface{} `json:"config"`

	ApiRelation []ApiPluginRelation `json:"apirelation"`

	CreatedAt           time.Time `json:"createdAt"`
	CreatedAtTimestamp  int64     `json:"createdAtTimestamp"`
	ReleasedAt          time.Time `json:"releasedAt"`
	ReleasedAtTimestamp int64     `json:"releasedAtTimestamp"`

	Status string `json:"status"`
}

type ApiPluginRelation struct {
	Id           int    `json:"id"`
	ApiPluginId  string `json:"apiGroupId"`
	ApiId        string `json:"apiId"`
	KongPluginId string
}

type ApiBind struct {
	ID string `json:"id"`
}

type FieldValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type FieldDetailInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}
type InputResTransformerConfig struct {
	Remove struct {
		Json    []string `json:"json,omitempty"`
		Headers []string `json:"headers,omitempty"`
	} `json:"remove,omitempty"`
	Rename struct {
		Headers []FieldValue `json:"headers,omitempty"`
	} `json:"rename,omitempty"`
	Replace struct {
		Json    []FieldDetailInfo `json:"json,omitempty"`
		Headers []FieldValue      `json:"headers,omitempty"`
	} `json:"replace,omitempty"`
	Add struct {
		Json    []FieldDetailInfo `json:"json,omitempty"`
		Headers []FieldValue      `json:"headers,omitempty"`
	} `json:"add,omitempty"`
	Append struct {
	} `json:"append,omitempty"`
}

type OutResponseTransformer struct {
	Name   string                    `json:"name"`
	Config InputResTransformerConfig `json:"config,omitempty"`
}

type ResponseTransformer struct {
	Name   string               `json:"name"`
	Config ResTransformerConfig `json:"config,omitempty"`
}
type ResTransformerConfig struct {
	Remove struct {
		Json    []string `json:"json,omitempty"`
		Headers []string `json:"headers,omitempty"`
	} `json:"remove,omitempty"`
	Rename struct {
		Headers []string `json:"headers,omitempty"`
	} `json:"rename,omitempty"`
	Replace struct {
		Json       []string `json:"json,omitempty"`
		Json_types []string `json:"json_types,omitempty"`
		Headers    []string `json:"headers,omitempty"`
	} `json:"replace,omitempty"`
	Add struct {
		Json       []string `json:"json,omitempty"`
		Json_types []string `json:"json_types,omitempty"`
		Headers    []string `json:"headers,omitempty"`
	} `json:"add,omitempty"`
	Append struct {
		Json       []string `json:"json,omitempty"`
		Json_types []string `json:"json_types,omitempty"`
		Headers    []string `json:"headers,omitempty"`
	} `json:"append,omitempty"`
}

func ToInputTransformerInfo(info ResTransformerConfig) InputResTransformerConfig {
	var output InputResTransformerConfig
	output.Remove = info.Remove

	for i := 0; i < len(info.Rename.Headers); i++ {
		arr := strings.Split(info.Rename.Headers[i], ":")
		output.Rename.Headers = append(output.Rename.Headers, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Replace.Headers); i++ {
		arr := strings.Split(info.Replace.Headers[i], ":")
		output.Replace.Headers = append(output.Replace.Headers, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Replace.Json); i++ {
		arr := strings.Split(info.Replace.Json[i], ":")
		output.Replace.Json = append(output.Replace.Json,
			FieldDetailInfo{Key: arr[0], Value: arr[1], Type: info.Replace.Json_types[i]})
	}
	for i := 0; i < len(info.Add.Headers); i++ {
		arr := strings.Split(info.Add.Headers[i], ":")
		output.Add.Headers = append(output.Add.Headers, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Add.Json); i++ {
		arr := strings.Split(info.Add.Json[i], ":")
		output.Add.Json = append(output.Add.Json,
			FieldDetailInfo{Key: arr[0], Value: arr[1], Type: info.Add.Json_types[i]})
	}
	return output
}

func FromModel(m model.ApiPlugin, ss []model.ApiPluginRelation) (ApiPlugin, error) {
	result := ApiPlugin{
		Id:                  m.Id,
		Name:                m.Name,
		Namespace:           m.Namespace,
		Type:                m.Type,
		ConsumerId:          m.ConsumerId,
		User:                m.User,
		CreatedAt:           m.CreatedAt,
		CreatedAtTimestamp:  m.CreatedAt.Unix(),
		ReleasedAt:          m.ReleasedAt,
		ReleasedAtTimestamp: m.ReleasedAt.Unix(),

		Status: m.Status,
	}

	switch m.Type {
	case "response-transformer":
		config := &ResponseTransformer{}
		err := json.Unmarshal([]byte(m.Raw), config)
		if err != nil {
			return result, fmt.Errorf("unmarshal crd v1.api error: %+v", err)
		}
		inputConfig := ToInputTransformerInfo(config.Config)
		outConfig := &OutResponseTransformer{}
		outConfig.Config = inputConfig
		result.Config = outConfig
		klog.Infof("get input result config %+v", config)
	default:
		klog.Infof("plugin type is invaild config %s", m.Type)
	}

	if ss != nil {
		scenarios := make([]ApiPluginRelation, len(ss))
		for i := range ss {
			scenarios[i] = FromModelScenario(ss[i])
		}
		result.ApiRelation = scenarios
	}

	username, err := cas.GetUserNameByID(m.User)
	if err == nil {
		result.UserName = username
	} else {
		result.UserName = "用户数据错误"
	}
	return result, nil
}

func ToModel(a ApiPlugin) (model.ApiPlugin, []model.ApiPluginRelation, error) {
	apis := make([]model.ApiPluginRelation, len(a.ApiRelation))
	for i := range a.ApiRelation {
		apis[i] = ToModelScenario(a.ApiRelation[i], a.Id, "")
	}

	result := model.ApiPlugin{
		Id:          a.Id,
		Name:        a.Name,
		Type:        a.Type,
		Namespace:   a.Namespace,
		User:        a.User,
		Description: a.Description,
		CreatedAt:   a.CreatedAt,
		ReleasedAt:  a.ReleasedAt,
		ConsumerId:  a.ConsumerId,

		Status: a.Status,
	}
	if a.Config != nil {
		err, config := assignmentConfig(a.Type, a.Config)
		if err != nil {
			return result, apis, fmt.Errorf("assignnmentConfig is error,req data: %v", err)
		}

		configJson, err := json.Marshal(config)
		if err != nil {
			return result, apis, fmt.Errorf("json.Marshal config error,: %v", err)
		}

		result.Raw = string(configJson)
		klog.Infof("get result raw config %+v, raw %s", config, result.Raw)
	}

	return result, apis, nil
}

func FromModelScenario(m model.ApiPluginRelation) ApiPluginRelation {
	return ApiPluginRelation{
		Id:           m.Id,
		ApiPluginId:  m.ApiPluginId,
		ApiId:        m.ApiId,
		KongPluginId: m.KongPluginId,
	}
}

func ToModelScenario(a ApiPluginRelation, productId string, apiId string) model.ApiPluginRelation {
	return model.ApiPluginRelation{
		Id:          a.Id,
		ApiPluginId: productId,
		ApiId:       apiId,
		//KongPluginId:
	}
}
func assignmentConfig(name string, reqData interface{}) (error, *ResponseTransformer) {
	klog.Infof("enter assignmentConfig name %s, config %+v", name, reqData)
	data, ok := reqData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("reqData type is error,req data: %v", reqData), nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err), nil
	}

	switch name {
	case "response-transformer":
		var input InputResTransformerConfig
		if err = json.Unmarshal(b, &input); err != nil {
			return fmt.Errorf("json.Unmarshal error,: %v", err), nil
		}
		responseTransformer := &ResponseTransformer{}
		responseTransformer.Name = name

		responseTransformer.Config.Remove = input.Remove
		for i := 0; i < len(input.Rename.Headers); i++ {
			responseTransformer.Config.Rename.Headers = append(responseTransformer.Config.Rename.Headers,
				input.Rename.Headers[i].Key+":"+input.Rename.Headers[i].Value)
		}
		for i := 0; i < len(input.Replace.Headers); i++ {
			responseTransformer.Config.Replace.Headers = append(responseTransformer.Config.Replace.Headers,
				input.Replace.Headers[i].Key+":"+input.Replace.Headers[i].Value)
		}
		for i := 0; i < len(input.Replace.Json); i++ {
			responseTransformer.Config.Replace.Json = append(responseTransformer.Config.Replace.Json,
				input.Replace.Json[i].Key+":"+input.Replace.Json[i].Value)
			responseTransformer.Config.Replace.Json_types = append(
				responseTransformer.Config.Replace.Json_types, input.Replace.Json[i].Type)
		}
		for i := 0; i < len(input.Add.Headers); i++ {
			responseTransformer.Config.Add.Headers = append(responseTransformer.Config.Add.Headers,
				input.Add.Headers[i].Key+":"+input.Add.Headers[i].Value)
		}
		for i := 0; i < len(input.Add.Json); i++ {
			responseTransformer.Config.Add.Json = append(responseTransformer.Config.Add.Json,
				input.Add.Json[i].Key+":"+input.Add.Json[i].Value)
			responseTransformer.Config.Add.Json_types = append(responseTransformer.Config.Add.Json_types,
				input.Add.Json[i].Type)
		}
		klog.V(5).Infof("assignmentConfig target responseTransformer %+v", *responseTransformer)
		return nil, responseTransformer
	}
	klog.V(5).Infof("assignmentConfig target error")
	return nil, nil
}
