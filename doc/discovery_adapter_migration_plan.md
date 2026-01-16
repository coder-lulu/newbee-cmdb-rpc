# 自动发现适配器迁移计划

## 项目状态

**当前状态**: 🔒 **已屏蔽** - 自动发现功能已在CMDB-RPC中屏蔽  
**迁移目标**: 🎯 **Agent服务** - 功能将完整移植到Agent服务中实现  
**实施时间**: 待定

## 屏蔽原因

1. **架构优化**: 自动发现功能更适合在Agent服务中实现，便于集中管理和调度
2. **依赖管理**: Nmap等网络工具在CMDB-RPC服务中引入复杂依赖
3. **功能边界**: 网络发现属于基础设施管理，应与业务数据管理解耦
4. **性能考虑**: 避免网络扫描影响CMDB核心服务性能

## 当前实现状态

### ✅ 已完成屏蔽的功能

```go
// 主要方法已屏蔽返回错误
- PreProcess()     -> 返回错误："自动发现功能已屏蔽"
- Parse()          -> 返回空数组和错误
- PostProcess()    -> 返回错误
- Validate()       -> 返回禁用状态

// 扫描功能已屏蔽
- performNetworkScan()  -> 返回错误："待移植到Agent实现"
- performPortScan()     -> 返回错误："待移植到Agent实现" 
- performSNMPDiscovery() -> 返回错误："待移植到Agent实现"
```

### 📋 保留的工具方法（供Agent移植参考）

以下方法已保留完整实现，供Agent服务移植时参考：

#### 目标解析功能
- `parseDiscoveryTargets()` - JSON/文本格式目标解析
- `getDefaultDiscoveryTargets()` - 自动网段检测

#### 配置和规则管理
- `getPortsFromRule()` - 端口列表解析
- `determineCITypeByRule()` - CI类型推断
- `matchRule()` - 规则匹配逻辑
- `getCITypeName()` - CI类型名称映射

#### 数据处理工具
- `generateAssetID()` - 资产ID生成
- `cleanAndNormalizeData()` - 数据清洗规范化
- `testSNMPConnectivity()` - SNMP连接测试

## Agent服务移植计划

### Phase 1: 环境准备 (1周)

**依赖安装**:
```bash
# Agent服务添加网络发现依赖
go get github.com/Ullaakut/nmap/v3
go get github.com/soniah/gosnmp  # SNMP支持
```

**目录结构**:
```
agent/
├── plugins/
│   └── discovery/
│       ├── network_scanner.go    # 网络扫描实现
│       ├── port_scanner.go       # 端口扫描实现
│       ├── snmp_discovery.go     # SNMP发现实现
│       ├── rule_engine.go        # 规则引擎
│       └── data_processor.go     # 数据处理器
```

### Phase 2: 核心功能移植 (2-3周)

#### 2.1 网络扫描引擎
```go
// 真实Nmap集成实现
type NetworkScanner struct {
    config *ScanConfig
    logger logx.Logger
}

func (s *NetworkScanner) ScanHosts(targets []string) ([]*DiscoveredHost, error) {
    scanner, err := nmap.NewScanner(
        nmap.WithTargets(targets...),
        nmap.WithHostDiscovery(),
        nmap.WithTimingTemplate(nmap.TimingAggressive),
    )
    // ... 真实实现
}
```

#### 2.2 端口发现引擎
```go
// 端口和服务发现
func (s *PortScanner) ScanPorts(targets []string, ports []string) ([]*DiscoveredService, error) {
    scanner, err := nmap.NewScanner(
        nmap.WithTargets(targets...),
        nmap.WithPorts(ports...),
        nmap.WithServiceInfo(),
    )
    // ... 真实实现
}
```

#### 2.3 SNMP发现引擎
```go
// SNMP设备发现
func (s *SNMPDiscovery) DiscoverDevices(targets []string) ([]*DiscoveredDevice, error) {
    for _, target := range targets {
        conn := &gosnmp.GoSNMP{
            Target:    target,
            Port:      161,
            Community: "public",
            Version:   gosnmp.Version2c,
        }
        // ... SNMP查询实现
    }
}
```

### Phase 3: 集成和调度 (1周)

#### 3.1 发现任务调度
```go
type DiscoveryScheduler struct {
    rules     []*DiscoveryRule
    scheduler cron.Cron
}

func (s *DiscoveryScheduler) Start() {
    // 定时发现任务
    s.scheduler.AddFunc("@every 5m", s.runDiscoveryTasks)
}
```

#### 3.2 数据上报接口
```go
// 发现结果上报到CMDB
func (s *DiscoveryService) ReportAssets(assets []*DiscoveredAsset) error {
    // 调用CMDB-RPC的资产导入接口
    return s.cmdbClient.ImportAssets(assets)
}
```

### Phase 4: 测试和部署 (1周)

#### 4.1 功能测试
- 网络扫描准确性测试
- 端口发现完整性测试  
- SNMP设备识别测试
- 大规模网段扫描测试

#### 4.2 性能测试
- 并发扫描能力测试
- 内存使用优化测试
- 网络带宽影响测试

## 移植清单

### 📦 从CMDB-RPC中移植的代码

| 组件 | 文件位置 | 移植到Agent位置 | 状态 |
|------|---------|----------------|------|
| 目标解析 | `parseDiscoveryTargets()` | `discovery/target_parser.go` | ✅ 保留 |
| 网段检测 | `getDefaultDiscoveryTargets()` | `discovery/network_detector.go` | ✅ 保留 |
| 规则引擎 | `determineCITypeByRule()` | `discovery/rule_engine.go` | ✅ 保留 |
| 数据清洗 | `cleanAndNormalizeData()` | `discovery/data_processor.go` | ✅ 保留 |
| 配置管理 | `DiscoveryConfig` | `discovery/config.go` | ✅ 保留 |

### 🔧 需要在Agent中新增的功能

| 功能 | 描述 | 优先级 |
|------|------|--------|
| **Nmap集成** | 真实网络扫描实现 | 🔴 高 |
| **SNMP库集成** | 设备信息查询 | 🟡 中 |
| **任务调度** | 定时发现任务 | 🔴 高 |
| **结果上报** | 向CMDB上报资产 | 🔴 高 |
| **性能监控** | 扫描性能指标 | 🟡 中 |

## 风险评估

### 🔴 高风险项
1. **网络权限**: Agent需要网络扫描权限
2. **工具依赖**: Nmap等外部工具安装和配置
3. **网络影响**: 大规模扫描可能影响网络性能

### 🟡 中风险项  
1. **数据同步**: Agent与CMDB数据一致性
2. **配置迁移**: 现有发现规则的迁移
3. **监控集成**: 发现任务的监控和告警

### 🟢 低风险项
1. **代码移植**: 大部分逻辑已经实现
2. **接口兼容**: 使用现有的资产导入接口
3. **配置管理**: 配置结构已经定义

## 测试策略

### 单元测试
```bash
# Agent发现功能单元测试
go test ./plugins/discovery/... -v

# 网络扫描模块测试
go test ./plugins/discovery/network_scanner_test.go

# SNMP发现模块测试  
go test ./plugins/discovery/snmp_discovery_test.go
```

### 集成测试
```bash
# 端到端发现流程测试
go test ./test/e2e/discovery_test.go

# 与CMDB集成测试
go test ./test/integration/cmdb_integration_test.go
```

### 性能测试
```bash
# 大规模网段扫描测试
go test ./test/performance/large_network_test.go -bench=.

# 并发发现能力测试
go test ./test/performance/concurrent_discovery_test.go -bench=.
```

## 配置文件示例

### Agent配置
```yaml
# agent/etc/agent.yaml
discovery:
  enabled: true
  scan_interval: 5m
  max_targets: 1000
  timeout: 30s
  
  # 发现规则
  rules:
    - id: "network_scan_rule"
      name: "网络主机发现"
      type: "network_scan"
      enabled: true
      targets: ["192.168.1.0/24", "10.0.0.0/24"]
      
    - id: "port_scan_rule"  
      name: "服务端口发现"
      type: "port_scan"
      enabled: true
      ports: ["22", "80", "443", "3306", "5432"]
      
  # CMDB上报配置
  cmdb:
    endpoint: "http://cmdb-rpc:8080"
    timeout: 10s
    batch_size: 100
```

## 成功标准

### 功能标准 ✅
- [ ] 网络主机发现准确率 > 95%
- [ ] 端口服务识别完整率 > 90%
- [ ] SNMP设备发现成功率 > 85%
- [ ] 资产数据上报成功率 > 99%

### 性能标准 📊
- [ ] 单个C类网段扫描时间 < 5分钟
- [ ] 并发支持 > 10个发现任务
- [ ] 内存使用 < 512MB
- [ ] CPU使用率 < 30%

### 可靠性标准 🛡️
- [ ] 服务可用性 > 99.9%
- [ ] 错误恢复时间 < 30秒
- [ ] 数据丢失率 < 0.1%

## 时间计划

| 阶段 | 时间 | 里程碑 |
|------|------|--------|
| **Phase 1** | 第1周 | 环境准备完成 |
| **Phase 2** | 第2-4周 | 核心功能实现 |
| **Phase 3** | 第5周 | 集成调度完成 |
| **Phase 4** | 第6周 | 测试部署完成 |

**总计**: 6周

---

## 联系信息

**技术负责人**: Agent开发团队  
**项目跟踪**: 待确定  
**文档更新**: 本文档将随迁移进展实时更新

---

**注意**: 当前CMDB-RPC中的自动发现适配器已完全屏蔽，不会影响现有业务功能。所有发现功能将在Agent服务中重新实现并提供更强大的能力。 