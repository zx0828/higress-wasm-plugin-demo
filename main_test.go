package main

import (
	"testing"

	"github.com/higress-group/wasm-go/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

// mockLog 实现了 log.Log 接口，用于测试
type mockLog struct{}

func (m *mockLog) Trace(msg string)                          {}
func (m *mockLog) Tracef(format string, args ...interface{}) {}
func (m *mockLog) Debug(msg string)                          {}
func (m *mockLog) Debugf(format string, args ...interface{}) {}
func (m *mockLog) Info(msg string)                           {}
func (m *mockLog) Infof(format string, args ...interface{})  {}
func (m *mockLog) Warn(msg string)                           {}
func (m *mockLog) Warnf(format string, args ...interface{})  {}
func (m *mockLog) Error(msg string)                          {}
func (m *mockLog) Errorf(format string, args ...interface{})  {}
func (m *mockLog) Critical(msg string)                       {}
func (m *mockLog) Criticalf(format string, args ...interface{}) {}
func (m *mockLog) ResetID(pluginID string)                  {}

func TestParseConfig(t *testing.T) {
	var testLogger log.Log = &mockLog{}

	t.Run("default config", func(t *testing.T) {
		config := &PluginConfig{}
		json := gjson.Parse(`{}`)
		err := parseConfig(json, config, testLogger)
		assert.NoError(t, err)
		assert.Equal(t, "default-client", config.ClientId)
		assert.Equal(t, "default-secret", config.SecretId)
	})

	t.Run("custom client", func(t *testing.T) {
		config := &PluginConfig{}
		json := gjson.Parse(`{"clientId": "my-client", "secretId": "my-secret"}`)
		err := parseConfig(json, config, testLogger)
		assert.NoError(t, err)
		assert.Equal(t, "my-client", config.ClientId)
		assert.Equal(t, "my-secret", config.SecretId)
	})
}
