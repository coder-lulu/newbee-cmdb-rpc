# CI增强搜索引擎设计文档

> **Version**: 1.0  
> **Last Update**: 2025-01-20  
> **Author**: @assistant  

## 一、项目框架

### 1.1 数据模型架构

基于完整的schema分析，CI系统采用了以下数据模型架构：

#### 核心实体关系
- **Cis (CI实例)**: 核心业务实体，记录具体的配置项实例
- **CiType (CI类型)**: 定义CI的类型模板，如服务器、数据库等
- **Attribute (属性)**: 定义可附加到CI类型的属性元数据

#### 属性值存储设计
采用分表存储不同类型的属性值，提高查询性能：
- `ValueText` - 文本类型属性值
- `ValueIndexText` - 索引文本属性值（优化搜索性能）
- `ValueInteger` - 整数类型属性值
- `ValueFloat` - 浮点数类型属性值
- `ValueDatetime` - 日期时间类型属性值
- `ValueJSON` - JSON类型属性值

#### 选项管理
- `ChoiceText`、`ChoiceInteger`、`ChoiceFloat` - 存储属性的可选值

#### 类型管理
- `CiTypeAttribute` - CI类型包含的属性及其配置
- `CiTypeAttributeGroup` - 属性分组管理
- `CiTypeGroup` - CI类型分组管理
- `CiTypeInheritance` - CI类型继承关系
- `CiTypeRelation` - CI类型间关系约束

#### 关系管理
- `CiRelation` - CI实例间的关系
- `RelationType` - 关系类型定义

## 二、核心功能

### 2.1 搜索引擎功能特性

#### 基础搜索功能
1. **基础字段过滤**
   - 按创建时间、更新时间、删除时间过滤
   - 按CI类型、状态过滤
   - 按创建者过滤

2. **属性搜索**
   - 支持所有属性数据类型的过滤
   - 支持复杂的操作符（eq, ne, gt, lt, like, in等）
   - 支持复合过滤组（AND/OR逻辑）

3. **全文搜索**
   - 优先使用索引文本表进行快速搜索
   - 支持跨属性的关键词搜索
   - 支持搜索字段限定

#### 高级搜索功能
1. **关系搜索**
   - 基于CI间关系进行搜索
   - 支持关系方向控制（incoming/outgoing/both）
   - 支持多层关系深度搜索
   - 支持间接关系查找

2. **继承搜索**
   - 基于CI类型继承关系扩展搜索范围
   - 支持包含子类型/父类型
   - 支持指定类型层级

3. **标签搜索**
   - 基于CI的tags字段进行搜索
   - 支持标签存在性检查
   - 支持标签值匹配

4. **元数据搜索**
   - 基于CI的metadata和custom_fields进行搜索
   - 支持JSON路径查询
   - 支持多种操作符

5. **时间范围搜索**
   - 支持任意时间字段的范围查询
   - 支持时间间隔统计

6. **分面搜索**
   - 提供搜索结果的分面统计
   - 支持多字段分面
   - 可配置返回数量

#### 结果增强功能
1. **智能排序**
   - 支持基础字段排序
   - 支持属性字段排序（规划中）
   - 支持多字段组合排序

2. **聚合统计**
   - 按状态、类型统计
   - 按时间维度统计
   - 自定义聚合字段

3. **相关推荐**
   - 基于关系自动推荐相关CI
   - 支持关系链分析

### 2.2 搜索请求结构

```go
type SearchRequest struct {
    *cmdb.CisListReq
    // 扩展搜索条件
    RelationFilters   []*RelationFilter   // 关系过滤
    InheritanceSearch *InheritanceSearch  // 继承搜索
    TagFilters        []*TagFilter        // 标签过滤
    MetadataFilters   []*MetadataFilter   // 元数据过滤
    SpatialSearch     *SpatialSearch      // 空间搜索
    TimeRangeSearch   *TimeRangeSearch    // 时间范围搜索
    FacetSearch       *FacetSearch        // 分面搜索
}
```

#### 关系过滤示例
```go
RelationFilter{
    RelationType:    "connect",          // 关系类型
    Direction:       "outgoing",         // 关系方向
    TargetCiIds:     []uint64{1,2,3},   // 目标CI
    TargetTypeIds:   []uint64{10,20},   // 目标类型
    RelationDepth:   2,                 // 关系深度
    IncludeIndirect: true,              // 包含间接关系
}
```

#### 继承搜索示例
```go
InheritanceSearch{
    IncludeChildren: true,               // 包含子类型
    IncludeParents:  false,              // 不包含父类型
    TypeHierarchy:   []uint64{5,6,7},   // 指定类型层级
}
```

#### 标签过滤示例
```go
TagFilter{
    Key:      "environment",            // 标签键
    Values:   []string{"prod", "test"}, // 标签值
    Operator: "in",                     // 操作符
}
```

### 2.3 搜索结果结构

```go
type SearchResult struct {
    *cmdb.CisListResp
    Facets       map[string][]FacetItem  // 分面统计
    Highlights   map[uint64][]string     // 搜索高亮
    SearchTime   int64                   // 搜索耗时(毫秒)
    TotalPages   int                     // 总页数
    CurrentPage  int                     // 当前页
    RelatedCis   []uint64                // 相关CI列表
}
```

## 三、使用示例

### 3.1 基础搜索

```go
engine := NewEnhancedSearchEngine(ctx, svcCtx)

req := &SearchRequest{
    CisListReq: &cmdb.CisListReq{
        Page:     1,
        PageSize: 20,
        TypeId:   &typeId,
        Search:   &searchKeyword,
    },
}

result, err := engine.Search(req)
```

### 3.2 关系搜索

查找与特定服务器直接连接的所有设备：

```go
req := &SearchRequest{
    CisListReq: &cmdb.CisListReq{
        Page:     1,
        PageSize: 20,
    },
    RelationFilters: []*RelationFilter{{
        RelationType:  "connect",
        Direction:     "both",
        TargetCiIds:   []uint64{serverCiId},
        RelationDepth: 1,
    }},
}
```

### 3.3 继承搜索

查找所有服务器类型及其子类型的CI：

```go
req := &SearchRequest{
    CisListReq: &cmdb.CisListReq{
        Page:     1,
        PageSize: 20,
        TypeId:   &serverTypeId,
    },
    InheritanceSearch: &InheritanceSearch{
        IncludeChildren: true,
        IncludeParents:  false,
    },
}
```

### 3.4 复合搜索

结合多种搜索条件：

```go
req := &SearchRequest{
    CisListReq: &cmdb.CisListReq{
        Page:       1,
        PageSize:   20,
        Search:     &keyword,
        TypeId:     &typeId,
    },
    RelationFilters: []*RelationFilter{{
        RelationType: "dependency",
        Direction:    "incoming",
    }},
    TagFilters: []*TagFilter{{
        Key:      "environment",
        Values:   []string{"production"},
        Operator: "in",
    }},
    MetadataFilters: []*MetadataFilter{{
        Path:     "system.version",
        Value:    "2.0",
        Operator: "gte",
    }},
    FacetSearch: &FacetSearch{
        Fields: []string{"type_id", "status"},
        Size:   10,
    },
}
```

## 四、性能优化

### 4.1 索引策略

1. **属性值表索引**
   - 所有属性值表在 `(attr_id, value)` 上建立复合索引
   - `ValueIndexText` 表专门用于文本搜索优化

2. **关系表索引**
   - `CiRelation` 表在关系字段上建立索引
   - 支持双向关系查询优化

3. **基础字段索引**
   - CI表在常用搜索字段上建立索引
   - 支持多字段组合查询

### 4.2 查询优化

1. **分阶段查询**
   - 先通过条件筛选出CI ID集合
   - 再基于ID集合查询详细信息

2. **缓存策略**
   - 类型继承关系缓存
   - 热门搜索结果缓存

3. **并行查询**
   - 分面统计与主查询并行执行
   - 相关CI查询异步执行

## 五、扩展计划

### 5.1 搜索功能增强

1. **智能搜索**
   - 基于历史搜索的推荐
   - 搜索纠错和建议
   - 自然语言查询解析

2. **高级分析**
   - CI关系图谱分析
   - 变更影响分析
   - 配置漂移检测

3. **实时搜索**
   - 搜索结果实时更新
   - 变更通知推送

### 5.2 性能优化

1. **搜索引擎集成**
   - 集成Elasticsearch进行全文搜索
   - 支持复杂的搜索语法

2. **数据湖集成**
   - 历史数据归档和查询
   - 大数据分析支持

## 六、已知问题

### 6.1 功能限制

1. **属性排序**
   - 当前不支持按属性值排序
   - 计划在下个版本实现

2. **空间搜索**
   - 地理位置搜索功能待实现
   - 需要地理坐标数据支持

3. **搜索高亮**
   - 搜索结果高亮功能未完全实现
   - 需要前端配合显示

### 6.2 性能问题

1. **大数据量查询**
   - 超大结果集的分页性能需优化
   - 深度分页性能较差

2. **复杂关系查询**
   - 多层级关系查询性能有待提升
   - 需要增加关系路径缓存

## 七、文档索引

- [原始搜索逻辑文档](get_cis_list_logic.go)
- [数据模型Schema](../ent/schema/)
- [API接口文档](../api/)
- [测试用例](../../test/)

---

## 变更日志

- **2025-01-20**: 初版设计文档完成
- **计划中**: 属性排序功能实现
- **计划中**: Elasticsearch集成 