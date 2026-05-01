package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/higress-group/wasm-go/pkg/log"
	"github.com/higress-group/wasm-go/pkg/wrapper"
)

func main() {}

// generateSign 生成 HMAC-SHA1 签名: sign = HMAC-SHA1(sk, ak + timestamp)
func generateSign(ak, sk, timestamp string) string {
	plainText := ak + timestamp
	mac := hmac.New(sha1.New, []byte(sk))
	mac.Write([]byte(plainText))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func init() {
	wrapper.SetCtx(
		"mcp-signature",
		wrapper.ParseConfigBy(parseConfig),
		wrapper.ProcessRequestHeadersBy(onHttpRequestHeaders),
		wrapper.ProcessRequestBodyBy(onHttpRequestBody),
	)
}

type PluginConfig struct {
	PathPrefixes []string `json:"pathPrefixes"`
	AccessKey    string   `json:"accessKey"`
	AccessSecret string   `json:"accessSecret"`
	Timestamp    string   `json:"timestamp"`
}

func parseConfig(json gjson.Result, config *PluginConfig, log log.Log) error {
	// 从 YAML 配置中读取拦截路径前缀
	for _, item := range json.Get("pathPrefixes").Array() {
		config.PathPrefixes = append(config.PathPrefixes, item.String())
	}
	// 如果没配，默认只拦截 /mcp/ 开头的请求
	if len(config.PathPrefixes) == 0 {
		config.PathPrefixes = []string{"/mcp/"}
	}

	// 从 YAML 配置中读取密钥
	config.AccessKey = json.Get("accessKey").String()
	config.AccessSecret = json.Get("accessSecret").String()
	config.Timestamp = json.Get("timestamp").String()
	
	// 如果 YAML 没配，使用默认 fallback
	if config.AccessKey == "" {
		config.AccessKey = "a4918add974342fb9a4e99ddda3008ce"
	}
	if config.AccessSecret == "" {
		config.AccessSecret = "330000_yczx1958770414929416198"
	}
	
	log.Infof("[SignaturePlugin] 插件配置加载成功: 拦截路径=%v, accessKey=%s", config.PathPrefixes, config.AccessKey)
	return nil
}

func onHttpRequestHeaders(ctx wrapper.HttpContext, config PluginConfig, log log.Log) types.Action {
	// 1. 判断是否是目标 MCP 请求
	origPath, _ := proxywasm.GetHttpRequestHeader("x-envoy-original-path")
	if origPath == "" {
		origPath, _ = proxywasm.GetHttpRequestHeader(":path")
	}

	matched := false
	for _, pr := range config.PathPrefixes {
		if strings.HasPrefix(origPath, pr) {
			matched = true
			break
		}
	}
	
	// 如果不是目标路径，直接放行，完全不消耗性能，也不修改任何数据
	if !matched {
		return types.ActionContinue
	}

	// 2. 开始执行 MCP 签名逻辑
	// 优先级 1: 从请求头获取传入的 timestamp
	timestamp, _ := proxywasm.GetHttpRequestHeader("timestamp")
	
	if timestamp == "" {
		// 优先级 2: 从 YAML 配置获取 timestamp
		if config.Timestamp != "" {
			timestamp = config.Timestamp
		} else {
			// 优先级 3: 默认生成当前时间戳（毫秒）
			timestamp = fmt.Sprintf("%d", time.Now().UnixMilli())
		}
	}
	
	sign := generateSign(config.AccessKey, config.AccessSecret, timestamp)

	// 存入上下文，供 Body 阶段注入到 JSON-RPC 结构中
	ctx.SetContext("ak", config.AccessKey)
	ctx.SetContext("sk", config.AccessSecret)
	ctx.SetContext("sn", sign)
	ctx.SetContext("ts", timestamp)

	log.Infof("[SignaturePlugin] 签名计算完成: ak=%s, ts=%s, sign=%s", config.AccessKey, timestamp, sign)
	return types.ActionContinue
}

func onHttpRequestBody(ctx wrapper.HttpContext, config PluginConfig, body []byte, log log.Log) types.Action {
	// 如果 Header 阶段没有拦截（或者没计算出 ak），这里直接放行
	akInter := ctx.GetContext("ak")
	if akInter == nil {
		return types.ActionContinue
	}

	ak := akInter.(string)
	sk := ctx.GetContext("sk").(string)
	sn := ctx.GetContext("sn").(string)
	ts := ctx.GetContext("ts").(string)

	// 校验是否是 MCP JSON-RPC 结构
	res := gjson.GetBytes(body, "params")
	if !res.Exists() {
		return types.ActionContinue
	}

	// 智能探测 arguments 的 JSON 路径
	argPath := "params.arguments"
	if res.IsArray() {
		argPath = "params.0.arguments"
	}

	updatedBody := body
	var err error
	
	// 精准注入所有签名参数到 JSON-RPC arguments 中
	updatedBody, err = sjson.SetBytes(updatedBody, argPath+".accessKey", ak)
	if err == nil { updatedBody, err = sjson.SetBytes(updatedBody, argPath+".accessSecret", sk) }
	if err == nil { updatedBody, err = sjson.SetBytes(updatedBody, argPath+".sign", sn) }
	if err == nil { updatedBody, err = sjson.SetBytes(updatedBody, argPath+".timestamp", ts) }

	if err != nil {
		log.Errorf("[SignaturePlugin] 注入参数到 JSON-RPC 失败: %v", err)
		return types.ActionContinue
	}

	err = proxywasm.ReplaceHttpRequestBody(updatedBody)
	if err != nil {
		log.Errorf("[SignaturePlugin] 替换 Body 失败: %v", err)
	} else {
		log.Infof("[SignaturePlugin] 成功将签名参数注入 MCP arguments: %s", argPath)
	}

	return types.ActionContinue
}
