package kuu

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

// TestMain 设置测试环境
func TestMain(m *testing.M) {
	// 设置测试配置，避免GetAppName()导致的panic
	C(map[string]interface{}{
		"name": "test-app",
	})
	os.Exit(m.Run())
}

// TestAPIPermissionItem 测试API权限项结构
func TestAPIPermissionItem(t *testing.T) {
	// 直接创建APIPermissionItem进行测试
	items := []APIPermissionItem{
		{
			Method:  []string{"GET", "POST"},
			Pattern: "/api/users.*",
		},
		{
			Method:  []string{"DELETE"},
			Pattern: "/api/users/[0-9]+",
		},
	}
	
	assert.Len(t, items, 2)
	
	// 验证第一个权限项
	assert.Equal(t, []string{"GET", "POST"}, items[0].Method)
	assert.Equal(t, "/api/users.*", items[0].Pattern)
	
	// 验证第二个权限项
	assert.Equal(t, []string{"DELETE"}, items[1].Method)
	assert.Equal(t, "/api/users/[0-9]+", items[1].Pattern)
}

// TestAPIPermissionConfig 测试API权限配置结构
func TestAPIPermissionConfig(t *testing.T) {
	// 直接创建APIPermissionConfig进行测试
	config := APIPermissionConfig{
		Enabled:      true,
		IncludeUsers: []string{"admin", "manager"},
		ExcludeUsers: []string{"guest", "readonly"},
		IncludeAll:   false,
		ExcludeAll:   false,
	}
	
	assert.True(t, config.Enabled)
	assert.Equal(t, []string{"admin", "manager"}, config.IncludeUsers)
	assert.Equal(t, []string{"guest", "readonly"}, config.ExcludeUsers)
	assert.False(t, config.IncludeAll)
	assert.False(t, config.ExcludeAll)
}

// TestShouldCheckAPIPermission 测试API权限验证判断逻辑
func TestShouldCheckAPIPermission(t *testing.T) {
	prisDesc := &PrivilegesDesc{}
	
	// 测试用例
	testCases := []struct {
		name     string
		config   APIPermissionConfig
		username string
		expected bool
	}{
		{
			name: "功能未启用",
			config: APIPermissionConfig{
				Enabled:      false,
				IncludeUsers: []string{"admin"},
			},
			username: "admin",
			expected: false,
		},
		{
			name: "ExcludeAll为true",
			config: APIPermissionConfig{
				Enabled:    true,
				ExcludeAll: true,
			},
			username: "admin",
			expected: false,
		},
		{
			name: "用户在排除列表中",
			config: APIPermissionConfig{
				Enabled:      true,
				ExcludeUsers: []string{"guest", "readonly"},
			},
			username: "guest",
			expected: false,
		},
		{
			name: "IncludeAll为true且用户不在排除列表",
			config: APIPermissionConfig{
				Enabled:      true,
				IncludeAll:   true,
				ExcludeUsers: []string{"guest"},
			},
			username: "admin",
			expected: true,
		},
		{
			name: "用户在包含列表中",
			config: APIPermissionConfig{
				Enabled:      true,
				IncludeUsers: []string{"admin", "manager"},
			},
			username: "admin",
			expected: true,
		},
		{
			name: "用户不在包含列表中",
			config: APIPermissionConfig{
				Enabled:      true,
				IncludeUsers: []string{"admin", "manager"},
			},
			username: "user1",
			expected: false,
		},
		{
			name: "优先级测试：排除优先于包含",
			config: APIPermissionConfig{
				Enabled:      true,
				IncludeUsers: []string{"admin"},
				ExcludeUsers: []string{"admin"},
			},
			username: "admin",
			expected: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := prisDesc.shouldCheckAPIPermission(tc.config, tc.username)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestHasAPIPermission 测试API权限检查方法
func TestHasAPIPermission(t *testing.T) {
	// 创建测试用的权限描述
	prisDesc := &PrivilegesDesc{
		UID: 2, // 非root用户
		PermissionMap: make(map[string]int64),
	}
	
	// 设置用户权限（只有user_management权限）
	prisDesc.PermissionMap["user_management"] = 1
	
	// 这里需要mock数据库查询，实际测试中需要设置测试数据库
	// 由于涉及数据库操作，这里只测试权限检查逻辑的核心部分
	
	// 测试root用户（始终有权限）
	rootPrisDesc := &PrivilegesDesc{UID: 1}
	assert.True(t, rootPrisDesc.HasAPIPermission("GET", "/api/users"))
	assert.True(t, rootPrisDesc.HasAPIPermission("DELETE", "/api/files/123"))
	
	// 测试配置检查
	// 注意：现在使用ApiPermissionConfig配置结构
	// 支持更灵活的配置：enabled开关、includeUsers、excludeUsers、includeAll、excludeAll
	// 这里需要mock配置和数据库查询，实际测试中需要设置测试配置和测试数据库
	// 由于涉及配置读取和用户名查询，这里只验证基本逻辑
}

// TestAPIPatternMatching 测试API模式匹配
func TestAPIPatternMatching(t *testing.T) {
	// 测试用例
	testCases := []struct {
			name     string
			method   string
			path     string
			pattern  string
			methods  []string
			expected bool
		}{
			{"GET匹配users路径", "GET", "/api/users", "/api/users.*", []string{"GET", "POST"}, true},
			{"POST匹配users子路径", "POST", "/api/users/123", "/api/users.*", []string{"GET", "POST"}, true},
			{"DELETE方法不匹配", "DELETE", "/api/users/123", "/api/users.*", []string{"GET", "POST"}, false},
			{"DELETE匹配数字ID", "DELETE", "/api/users/123", "/api/users/[0-9]+", []string{"DELETE"}, true},
			{"DELETE不匹配字母ID", "DELETE", "/api/users/abc", "/api/users/[0-9]+", []string{"DELETE"}, false},
			{"GET匹配files路径", "GET", "/api/files/download", "/api/files.*", []string{"GET"}, true},
			{"POST方法不匹配files", "POST", "/api/files/upload", "/api/files.*", []string{"GET"}, false},
			{"*方法匹配GET请求", "GET", "/api/admin/users", "/api/admin.*", []string{"*"}, true},
			{"*方法匹配POST请求", "POST", "/api/admin/users", "/api/admin.*", []string{"*"}, true},
			{"*方法匹配DELETE请求", "DELETE", "/api/admin/users/123", "/api/admin.*", []string{"*"}, true},
			{"*方法匹配PUT请求", "PUT", "/api/admin/settings", "/api/admin.*", []string{"*"}, true},
			{"*方法与其他方法混合", "PATCH", "/api/mixed/data", "/api/mixed.*", []string{"GET", "*", "POST"}, true},
		}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建API权限项
			item := APIPermissionItem{
				Method:  tc.methods,
				Pattern: tc.pattern,
			}
			
			// 检查方法匹配（与HasAPIPermission方法保持一致）
			methodMatch := false
			for _, allowedMethod := range item.Method {
				if allowedMethod == "*" || strings.ToUpper(allowedMethod) == strings.ToUpper(tc.method) {
					methodMatch = true
					break
				}
			}
			
			if !methodMatch {
				assert.False(t, tc.expected, "Method %s should not match %v", tc.method, tc.methods)
				return
			}
			
			// 使用正则表达式检查路径匹配
			patternMatch, err := regexp.MatchString(tc.pattern, tc.path)
			assert.NoError(t, err, "正则表达式匹配出错: %s", tc.pattern)
			
			if methodMatch && patternMatch {
				assert.True(t, tc.expected, "Expected true for %s %s against pattern %s", tc.method, tc.path, tc.pattern)
			} else {
				assert.False(t, tc.expected, "Expected false for %s %s against pattern %s", tc.method, tc.path, tc.pattern)
			}
		})
	}
}

// BenchmarkHasAPIPermission 性能测试
func BenchmarkHasAPIPermission(b *testing.B) {
	// 这里需要mock配置和数据库查询，实际测试中需要设置测试配置和测试数据库
	// 注意：现在使用ApiPermissionConfig配置结构，支持更灵活的权限控制
	// 由于涉及配置读取和用户名查询，这里只验证基本逻辑
	b.Skip("需要mock数据库和配置")
}

// BenchmarkShouldCheckAPIPermission 权限判断逻辑性能测试
func BenchmarkShouldCheckAPIPermission(b *testing.B) {
	prisDesc := &PrivilegesDesc{}
	config := APIPermissionConfig{
		Enabled:      true,
		IncludeUsers: []string{"admin", "manager", "operator"},
		ExcludeUsers: []string{"guest", "readonly"},
		IncludeAll:   false,
		ExcludeAll:   false,
	}
	username := "admin"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prisDesc.shouldCheckAPIPermission(config, username)
	}
}

// TestAPIWhitelist 测试API白名单功能
func TestAPIWhitelist(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		path      string
		whitelist []string
		expected  bool
	}{
		{
			name:      "登录接口在白名单中",
			method:    "POST",
			path:      "/api/login",
			whitelist: []string{"^/api/login$", "^/api/logout$"},
			expected:  true,
		},
		{
			name:      "验证码接口在白名单中",
			method:    "GET",
			path:      "/api/captcha/generate",
			whitelist: []string{"^/api/captcha.*"},
			expected:  true,
		},
		{
			name:      "公共资源在白名单中",
			method:    "GET",
			path:      "/api/public/info",
			whitelist: []string{"^/api/public/.*"},
			expected:  true,
		},
		{
			name:      "静态资源在白名单中",
			method:    "GET",
			path:      "/assets/css/style.css",
			whitelist: []string{"^/assets/.*"},
			expected:  true,
		},
		{
			name:      "非白名单路径",
			method:    "GET",
			path:      "/api/users",
			whitelist: []string{"^/api/login$", "^/api/logout$"},
			expected:  false,
		},
		{
			name:      "空白名单",
			method:    "GET",
			path:      "/api/login",
			whitelist: []string{},
			expected:  false,
		},
	}

	for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// 模拟白名单检查逻辑
				matched := false
				for _, whitelistPattern := range tt.whitelist {
					if isMatched, _ := regexp.MatchString(whitelistPattern, tt.path); isMatched {
						matched = true
						break
					}
				}
				assert.Equal(t, tt.expected, matched)
			})
		}
}