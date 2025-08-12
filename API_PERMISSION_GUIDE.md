# API权限验证功能使用指南

## 功能概述

本功能为kuu框架增加了细粒度的API权限控制，允许通过菜单配置来控制用户对特定API接口的访问权限。

## 功能特性

1. **菜单级API权限配置**：在Menu表中新增APIPattern字段，支持配置API权限表达式
2. **用户级权限检查**：通过配置文件控制哪些用户需要进行API权限验证
3. **灵活的权限表达式**：支持HTTP方法和URL模式匹配
4. **403错误返回**：当用户无权限时返回标准的403错误码

## 配置说明

### 1. kuu.json配置

在kuu.json配置文件中添加`ApiPermissionConfig`字段：

```json
{
  "prefix": "/api",
  "cors": true,
  "gzip": true,
  "db": {
    "dialect": "postgres",
    "args": "host=127.0.0.1 port=5432 user=root dbname=kuu password=hello sslmode=disable"
  },
  "ApiPermissionConfig": {
    "enabled": true,
    "includeUsers": ["admin", "manager"],
    "excludeUsers": ["guest", "readonly"],
    "includeAll": false,
    "excludeAll": false
  },
  "redis": {
    "addr": "127.0.0.1:6379"
  }
}
```

**配置字段说明**：
- `enabled`：是否启用API权限验证功能（必填）
- `includeUsers`：必须进行API权限验证的用户名数组（可选）
- `excludeUsers`：不需要进行API权限验证的用户名数组（可选）
- `includeAll`：是否对所有用户进行API权限验证（可选，默认false）
- `excludeAll`：是否对所有用户都不进行API权限验证（可选，默认false）
- root用户始终拥有所有权限，无需配置

**优先级规则**：
1. 如果`enabled`为false，则不进行任何API权限验证
2. 如果`excludeAll`为true，则所有用户都不进行验证
3. 如果用户在`excludeUsers`列表中，则不进行验证
4. 如果`includeAll`为true，则对所有用户进行验证（除了已被排除的）
5. 如果用户在`includeUsers`列表中，则进行验证
6. 其他情况默认不进行验证

### 配置示例

#### 示例1：只对特定用户进行验证
```json
{
  "ApiPermissionConfig": {
    "enabled": true,
    "includeUsers": ["admin", "manager"]
  }
}
```

#### 示例2：对所有用户验证，但排除特定用户
```json
{
  "ApiPermissionConfig": {
    "enabled": true,
    "includeAll": true,
    "excludeUsers": ["guest", "readonly"]
  }
}
```

#### 示例3：关闭API权限验证
```json
{
  "ApiPermissionConfig": {
    "enabled": false
  }
}
```

#### 示例4：对所有用户都不进行验证
```json
{
  "ApiPermissionConfig": {
    "enabled": true,
    "excludeAll": true
  }
}
```



### 2. 菜单APIPattern配置

在Menu表的APIPattern字段中配置API权限表达式，格式为JSON数组：

```json
[
  {
    "method": ["GET", "POST"],
    "pattern": "/api/users.*"
  },
  {
    "method": ["DELETE"],
    "pattern": "/api/users/[0-9]+"
  },
  {
    "method": ["PUT", "PATCH"],
    "pattern": "/api/users/[0-9]+/profile"
  }
]
```

**字段说明**：
- `method`：允许的HTTP方法数组，支持GET、POST、PUT、DELETE、PATCH等
- `pattern`：URL匹配模式，支持正则表达式

## 使用示例

### 示例1：用户管理权限

假设有一个"用户管理"菜单，需要控制用户对用户相关API的访问：

```sql
-- 菜单记录
INSERT INTO menu (code, name, api_pattern) VALUES (
  'user_management',
  '用户管理',
  '[
    {
      "method": ["GET"],
      "pattern": "/api/users.*"
    },
    {
      "method": ["POST"],
      "pattern": "/api/users"
    },
    {
      "method": ["PUT", "PATCH"],
      "pattern": "/api/users/[0-9]+"
    },
    {
      "method": ["DELETE"],
      "pattern": "/api/users/[0-9]+"
    }
  ]'
);
```

### 示例2：文件管理权限

```sql
-- 文件管理菜单
INSERT INTO menu (code, name, api_pattern) VALUES (
  'file_management',
  '文件管理',
  '[
    {
      "method": ["GET"],
      "pattern": "/api/files.*"
    },
    {
      "method": ["POST"],
      "pattern": "/api/upload"
    },
    {
      "method": ["DELETE"],
      "pattern": "/api/files/[0-9]+"
    }
  ]'
);
```

## 权限检查流程

1. **用户身份验证**：首先进行标准的JWT令牌验证
2. **Root用户检查**：Root用户始终拥有所有权限，直接允许访问
3. **API权限开关检查**：检查`ApiPermissionConfig.enabled`是否为true
4. **用户验证范围判断**：根据配置规则判断当前用户是否需要进行API权限验证
   - 检查`excludeAll`标志
   - 检查用户是否在`excludeUsers`列表中
   - 检查`includeAll`标志
   - 检查用户是否在`includeUsers`列表中
5. **菜单权限查询**：如果需要验证，获取用户有权限的菜单列表
6. **API模式匹配**：遍历菜单的APIPattern配置，检查当前请求是否匹配
7. **权限决策**：如果找到匹配的权限配置则允许访问，否则返回557错误

## 特殊情况处理

### Root用户
- Root用户始终拥有所有API权限，无需配置

### 白名单接口
- 在白名单中的接口不进行权限检查

### 配置错误处理
- 如果APIPattern配置格式错误，该菜单的API权限配置将被忽略
- 如果`ApiPermissionConfig`配置解析失败，默认不启用API权限验证
- 如果查询用户名失败，默认允许访问
- 配置冲突时按优先级规则处理：排除规则优先于包含规则

## 错误响应

当用户无权限访问API时，系统返回：

```json
{
  "code": 403,
  "msg": "API access denied",
  "data": null
}
```

支持多语言：
- 英文："API access denied"
- 简体中文："API访问被拒绝"
- 繁体中文："API訪問被拒絕"

## 最佳实践

1. **权限最小化原则**：只给用户分配必要的API权限
2. **模式精确匹配**：使用精确的正则表达式避免权限泄露
3. **定期审查**：定期检查和更新API权限配置
4. **测试验证**：在生产环境部署前充分测试权限配置

## 注意事项

1. APIPattern字段使用JSON格式，注意转义字符的正确使用
2. 正则表达式模式匹配区分大小写
3. HTTP方法匹配不区分大小写
4. 权限检查在身份验证之后进行，确保用户已登录
5. 修改APIPattern配置后立即生效，无需重启服务