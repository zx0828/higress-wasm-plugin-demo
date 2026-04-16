// Copyright (c) 2022 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"strings"

	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/tidwall/gjson"

	"github.com/higress-group/wasm-go/pkg/log"
	"github.com/higress-group/wasm-go/pkg/wrapper"
	"github.com/zmap/go-iptree/iptree"

	"higress-wasm-plugin-demo/pkg/sm3" // 使用内建高性能 SM3 包
)

func main() {}

// calculateSign 使用国密 SM3 算法计算签名
func calculateSign(clientId, body, secretId string) string {
	// 拼接规则保持一致：clientId + body + secretId
	signStr := fmt.Sprintf("%s%s%s", clientId, body, secretId)
	hash := sm3.Sum([]byte(signStr))
	return fmt.Sprintf("%x", hash)
}

// --- 插件初始化 ---
func init() {
	wrapper.SetCtx(
		"higress-wasm-plugin-demo",
		wrapper.ParseConfigBy(parseConfig),
		wrapper.ProcessRequestHeadersBy(onHttpRequestHeaders),
		wrapper.ProcessRequestBodyBy(onHttpRequestBody),
		wrapper.ProcessResponseHeadersBy(onHttpResponseHeaders),
	)
}

// --- 配置定义 ---
type PluginConfig struct {
	ClientId  string         `json:"clientId"`
	SecretId  string         `json:"secretId"`
	WhiteList *iptree.IPTree `json:"whiteList"`
	Message   string         `json:"message"`
}

func parseConfig(json gjson.Result, config *PluginConfig, log log.Log) error {
	config.ClientId = json.Get("clientId").String()
	config.SecretId = json.Get("secretId").String()
	config.Message = json.Get("message").String()

	ips := json.Get("whiteList").Array()
	if len(ips) > 0 {
		tree := iptree.New()
		for _, ipStr := range ips {
			str := ipStr.String()
			if err := tree.AddByString(str, struct{}{}); err != nil {
				log.Errorf("添加 IP 到白名单失败 %s: %v", str, err)
			}
		}
		config.WhiteList = tree
	}

	if config.ClientId == "" {
		config.ClientId = "default-client"
	}
	if config.SecretId == "" {
		config.SecretId = "default-secret"
	}

	log.Infof("插件配置加载成功: clientId=%s, 签名算法=SM3, 是否开启白名单=%v", config.ClientId, config.WhiteList != nil)
	return nil
}

// --- 逻辑处理 ---

func onHttpRequestHeaders(ctx wrapper.HttpContext, config PluginConfig, log log.Log) types.Action {
	remoteAddr, err := proxywasm.GetProperty([]string{"source", "address"})
	if err == nil {
		clientIP := string(remoteAddr)
		if idx := strings.LastIndex(clientIP, ":"); idx != -1 {
			clientIP = clientIP[:idx]
		}

		if config.WhiteList != nil {
			if _, matched, errMatch := config.WhiteList.GetByString(clientIP); errMatch == nil && matched {
				log.Infof("[白名单] 匹配成功: IP %s 在白名单中，将跳过请求体签名", clientIP)
				_ = proxywasm.SetProperty([]string{"wasm", "skip_sign"}, []byte("true"))
			}
		}
	}

	proxywasm.ReplaceHttpRequestHeader("X-Client-Id", config.ClientId)
	proxywasm.ReplaceHttpRequestHeader("X-Secret-Id", config.SecretId)

	// 初始 SM3 签名（空Body）
	sign := calculateSign(config.ClientId, "", config.SecretId)
	proxywasm.ReplaceHttpRequestHeader("X-Sign", sign)
	_ = proxywasm.SetProperty([]string{"wasm", "final_sign"}, []byte(sign))

	return types.ActionContinue
}

func onHttpRequestBody(ctx wrapper.HttpContext, config PluginConfig, body []byte, log log.Log) types.Action {
	if skip, err := proxywasm.GetProperty([]string{"wasm", "skip_sign"}); err == nil && string(skip) == "true" {
		return types.ActionContinue
	}

	if len(body) > 0 {
		sign := calculateSign(config.ClientId, string(body), config.SecretId)
		proxywasm.ReplaceHttpRequestHeader("X-Sign", sign)
		_ = proxywasm.SetProperty([]string{"wasm", "final_sign"}, []byte(sign))
	}
	return types.ActionContinue
}

func onHttpResponseHeaders(ctx wrapper.HttpContext, config PluginConfig, log log.Log) types.Action {
	if sign, err := proxywasm.GetProperty([]string{"wasm", "final_sign"}); err == nil {
		proxywasm.AddHttpResponseHeader("X-Sign", string(sign))
	}
	proxywasm.AddHttpResponseHeader("X-Client-Id", config.ClientId)
	if config.Message != "" {
		proxywasm.AddHttpResponseHeader("X-Wasm-Message", config.Message)
	}
	return types.ActionContinue
}
