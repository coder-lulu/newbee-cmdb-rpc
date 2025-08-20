# CI权限管理前端对接文档

> **Version**: 1.0  
> **Last Update**: 2024-12-19  
> **Author**: @CMDB Team  

## 目录
- [概述](#概述)
- [API接口文档](#api接口文档)
- [数据模型](#数据模型)
- [枚举值定义](#枚举值定义)
- [前端组件设计](#前端组件设计)
- [错误处理](#错误处理)
- [使用示例](#使用示例)
- [最佳实践](#最佳实践)

## 概述

CI权限管理系统提供完整的权限控制功能，支持细粒度的权限配置，包括用户、角色、部门级别的权限控制，以及基于CI类型、实例、属性和字段的权限范围控制。

### 核心功能
- **权限创建**: 支持创建各种类型的权限配置
- **权限更新**: 支持权限配置的修改和调整
- **权限删除**: 支持权限的删除和批量删除
- **权限查询**: 支持单个权限查询和列表查询
- **权限继承**: 支持权限的继承和传播
- **审批流程**: 支持权限变更的审批机制

## API接口文档

### 基础URL
```
基础URL: /api/v1/cmdb/cipermission
```

### 1. 创建权限

**接口地址**: `POST /api/v1/cmdb/cipermission/create`

**接口描述**: 创建新的CI权限配置

**请求参数**:
```json
{
  "permission_id": "perm_12345",
  "department_id": 1001,
  "scope_type": "ci_type",
  "ci_type_id": 1,
  "subject_type": "user",
  "subject_name": "张三",
  "permission_type": "allow",
  "permission_level": "read",
  "operations": {
    "operations": [
      {
        "operation": "read",
        "description": "读取权限"
      }
    ]
  },
  "priority": 10,
  "effective_from": 1640995200,
  "effective_to": 1672531200,
  "is_temporary": false,
  "require_approval": true,
  "risk_level": "medium",
  "require_mfa": false,
  "description": "用户读写权限"
}
```

**响应数据**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 12345,
    "msg": "权限创建成功"
  }
}
```

### 2. 更新权限

**接口地址**: `PUT /api/v1/cmdb/cipermission/update`

**请求参数**: 
```json
{
  "id": 12345,
  "permission_level": "write",
  "effective_to": 1675123200,
  "updated_by": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 3. 删除权限

**接口地址**: `DELETE /api/v1/cmdb/cipermission/delete`

**请求参数**:
```json
{
  "ids": [12345, 12346, 12347]
}
```

### 4. 获取权限详情

**接口地址**: `GET /api/v1/cmdb/cipermission/get/{id}`

### 5. 获取权限列表

**接口地址**: `GET /api/v1/cmdb/cipermission/list`

**查询参数**:
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认10，最大1000）
- `department_id`: 部门ID过滤
- `permission_id`: 权限ID模糊搜索
- `scope_type`: 权限范围类型过滤
- `subject_type`: 权限主体类型过滤
- `permission_level`: 权限级别过滤
- `risk_level`: 风险等级过滤

## 数据模型

### CiPermissionInfo 完整数据模型

```typescript
interface CiPermissionInfo {
  id?: number;
  permission_id?: string;
  department_id?: number;
  scope_type?: string;
  ci_type_id?: number;
  ci_id?: number;
  attribute_id?: number;
  field_name?: string;
  subject_type?: string;
  subject_id?: string;
  subject_name?: string;
  subject_code?: string;
  permission_type?: string;
  operations?: CiPermissionOperations;
  conditions?: CiPermissionConditions;
  priority?: number;
  permission_level?: string;
  effective_from?: number;
  effective_to?: number;
  is_temporary?: boolean;
  require_approval?: boolean;
  granted_by?: string;
  granted_by_name?: string;
  granted_at?: number;
  usage_count?: number;
  last_used_at?: number;
  status?: number;
  risk_level?: string;
  require_mfa?: boolean;
  inheritable?: boolean;
  description?: string;
  comments?: string;
  created_by?: string;
  updated_by?: string;
  created_at?: number;
  updated_at?: number;
}
```

## 枚举值定义

### 权限范围类型 (scope_type)
```typescript
enum ScopeType {
  GLOBAL = "global",
  CI_TYPE = "ci_type",
  CI_INSTANCE = "ci_instance",
  ATTRIBUTE = "attribute",
  FIELD = "field"
}
```

### 权限主体类型 (subject_type)
```typescript
enum SubjectType {
  USER = "user",
  ROLE = "role",
  DEPARTMENT = "department",
  GROUP = "group",
  SYSTEM = "system"
}
```

### 权限级别 (permission_level)
```typescript
enum PermissionLevel {
  NONE = "none",
  READ = "read",
  WRITE = "write",
  ADMIN = "admin",
  SUPER_ADMIN = "super_admin"
}
```

### 风险等级 (risk_level)
```typescript
enum RiskLevel {
  LOW = "low",
  MEDIUM = "medium",
  HIGH = "high",
  CRITICAL = "critical"
}
```

## 前端组件设计

### 1. 权限列表组件 (PermissionList)

**组件功能**:
- 分页显示权限列表
- 支持多条件筛选
- 支持批量操作

**关键特性**:
```vue
<template>
  <div class="permission-list">
    <!-- 搜索筛选区域 -->
    <div class="filter-section">
      <el-form :model="searchForm" inline>
        <el-form-item label="权限ID">
          <el-input v-model="searchForm.permission_id" />
        </el-form-item>
        <el-form-item label="主体类型">
          <el-select v-model="searchForm.subject_type">
            <el-option label="用户" value="user" />
            <el-option label="角色" value="role" />
          </el-select>
        </el-form-item>
      </el-form>
    </div>
    
    <!-- 表格区域 -->
    <el-table :data="permissionList">
      <el-table-column type="selection" />
      <el-table-column prop="permission_id" label="权限ID" />
      <el-table-column prop="subject_name" label="权限主体" />
      <el-table-column prop="permission_level" label="权限级别" />
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button size="small" @click="editPermission(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="deletePermission(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
```

### 2. 权限表单组件 (PermissionForm)

**组件功能**:
- 创建/编辑权限配置
- 动态表单验证
- 权限范围联动

## 错误处理

### HTTP状态码说明

| 状态码 | 说明 | 处理建议 |
|--------|------|----------|
| 200 | 成功 | 正常处理响应数据 |
| 400 | 请求参数错误 | 检查请求参数格式 |
| 401 | 未授权 | 重新登录 |
| 403 | 无权限 | 提示用户权限不足 |
| 422 | 数据验证失败 | 根据错误信息提示用户 |
| 500 | 服务器错误 | 提示系统异常 |

## 使用示例

### 1. 创建用户权限

```typescript
const createUserPermission = async (userId: string, ciTypeId: number) => {
  try {
    const permissionData = {
      subject_type: "user",
      subject_id: userId,
      subject_name: "张三",
      scope_type: "ci_type",
      ci_type_id: ciTypeId,
      permission_type: "allow",
      permission_level: "write",
      risk_level: "medium",
      description: "用户CI类型读写权限"
    };
    
    const response = await api.post('/cipermission/create', permissionData);
    ElMessage.success('权限创建成功');
    return response.data;
  } catch (error) {
    console.error('创建权限失败:', error);
    throw error;
  }
};
```

### 2. 权限列表查询

```typescript
const fetchPermissionList = async (params: PermissionListParams) => {
  try {
    const response = await api.get('/cipermission/list', { params });
    return {
      list: response.data.data.data,
      total: response.data.data.total
    };
  } catch (error) {
    console.error('查询权限列表失败:', error);
    throw error;
  }
};
```

## 最佳实践

### 1. 性能优化
- 使用分页查询，避免一次性加载大量数据
- 实现搜索防抖，减少不必要的API调用
- 缓存常用的枚举数据

### 2. 用户体验
- 根据选择动态显示/隐藏相关字段
- 提供智能默认值和推荐配置
- 提供详细的错误信息和解决建议

### 3. 安全考虑
- 前端进行基础权限检查，但不依赖前端验证
- 敏感操作需要二次确认
- 敏感信息在前端显示时进行脱敏

---

*此文档为CI权限管理系统的前端对接指南，如有疑问请联系后端开发团队。* 