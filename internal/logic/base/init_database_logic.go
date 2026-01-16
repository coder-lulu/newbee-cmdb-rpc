package base

import (
	"context"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"
	"github.com/coder-lulu/newbee-common/v2/msg/logmsg"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/entenum"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/attribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitDatabaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitDatabaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitDatabaseLogic {
	return &InitDatabaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *InitDatabaseLogic) InitDatabase(in *cmdb.Empty) (*cmdb.BaseResp, error) {
	if err := l.svcCtx.DB.Schema.Create(l.ctx, schema.WithForeignKeys(false)); err != nil {
		logx.Errorw(logmsg.DatabaseError, logx.Field("detail", err.Error()))
		return nil, errorx.NewInternalError(err.Error())
	}

	errHandler := func(err error) (*cmdb.BaseResp, error) {
		logx.Errorw(logmsg.DatabaseError, logx.Field("detail", err.Error()))
		return nil, errorx.NewInternalError(err.Error())
	}

	// 使用默认租户上下文（tenant_id = 1）初始化数据
	tenantCtx := hooks.SetTenantIDToContext(context.Background(), entenum.TenantDefaultId)

	// 插入初始数据
	err := l.InsertInitData(tenantCtx)
	if err != nil {
		return errHandler(err)
	}

	return &cmdb.BaseResp{Msg: errormsg.Success}, nil
}

func (l *InitDatabaseLogic) InsertInitData(ctx context.Context) error {
	// 1. 插入默认CI类型分组
	err := l.insertCiTypeGroups(ctx)
	if err != nil {
		return err
	}

	// 2. 插入关系类型
	err = l.insertRelationTypes(ctx)
	if err != nil {
		return err
	}

	// 3. 插入基础属性
	err = l.insertBaseAttributes(ctx)
	if err != nil {
		return err
	}

	// 4. 插入示例CI类型
	err = l.insertSampleCiTypes(ctx)
	if err != nil {
		return err
	}

	// 5. 为CI类型创建属性分组
	err = l.insertAttributeGroupsForCiTypes(ctx)
	if err != nil {
		return err
	}

	// 6. 为CI类型分配属性
	err = l.assignAttributesToCiTypes(ctx)
	if err != nil {
		return err
	}

	// 7. 创建CI类型关系（新增）
	err = l.insertCiTypeRelations(ctx)
	if err != nil {
		return err
	}

	return nil
}

// insertCiTypeGroups 插入CI类型分组
func (l *InitDatabaseLogic) insertCiTypeGroups(ctx context.Context) error {
	// 检查是否已存在
	existing, err := l.svcCtx.DB.CiTypeGroup.Query().Count(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}
	if existing > 0 {
		logx.Info("CI类型分组已存在，跳过创建")
		return nil
	}

	var groups []*ent.CiTypeGroupCreate
	groups = append(groups,
		l.svcCtx.DB.CiTypeGroup.Create().SetName("硬件设备").SetDescription("物理硬件设备").SetIcon("icon-hardware").SetSort(1),
		l.svcCtx.DB.CiTypeGroup.Create().SetName("软件系统").SetDescription("软件和应用系统").SetIcon("icon-software").SetSort(2),
		l.svcCtx.DB.CiTypeGroup.Create().SetName("数据库").SetDescription("数据库系统").SetIcon("icon-database").SetSort(3),
		l.svcCtx.DB.CiTypeGroup.Create().SetName("其它").SetDescription("其它类型").SetIcon("icon-other").SetSort(999),
		l.svcCtx.DB.CiTypeGroup.Create().SetName("基类").SetDescription("基础类型").SetIcon("icon-base").SetSort(998),
	)

	err = l.svcCtx.DB.CiTypeGroup.CreateBulk(groups...).Exec(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}
	logx.Info("CI类型分组创建成功")
	return nil
}

// insertRelationTypes 插入关系类型
func (l *InitDatabaseLogic) insertRelationTypes(ctx context.Context) error {
	// 检查是否已存在
	existing, err := l.svcCtx.DB.RelationType.Query().Count(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}
	if existing > 0 {
		logx.Info("关系类型已存在，跳过创建")
		return nil
	}

	relationTypes := []*ent.RelationTypeCreate{
		l.svcCtx.DB.RelationType.Create().SetName("包含").SetCode("contain").SetCategory("logical").SetDirection("bidirectional"),
		l.svcCtx.DB.RelationType.Create().SetName("部署").SetCode("deploy").SetCategory("logical").SetDirection("bidirectional"),
		l.svcCtx.DB.RelationType.Create().SetName("安装").SetCode("install").SetCategory("logical").SetDirection("bidirectional"),
		l.svcCtx.DB.RelationType.Create().SetName("关联").SetCode("associate").SetCategory("logical").SetDirection("bidirectional"),
	}
	err = l.svcCtx.DB.RelationType.CreateBulk(relationTypes...).Exec(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}
	logx.Info("关系类型创建成功")
	return nil
}

// insertBaseAttributes 插入基础属性
func (l *InitDatabaseLogic) insertBaseAttributes(ctx context.Context) error {
	// 检查是否已存在
	existing, err := l.svcCtx.DB.Attribute.Query().Count(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}
	if existing > 0 {
		logx.Info("基础属性已存在，跳过创建")
		return nil
	}

	// 创建基础属性
	attributes := []*ent.AttributeCreate{
		l.svcCtx.DB.Attribute.Create().SetName("name").SetAlias("名称").SetValueType("text"),
		l.svcCtx.DB.Attribute.Create().SetName("description").SetAlias("描述").SetValueType("longtext"),
		l.svcCtx.DB.Attribute.Create().SetName("status").SetAlias("状态").SetValueType("int"),
		l.svcCtx.DB.Attribute.Create().SetName("ip_address").SetAlias("IP地址").SetValueType("text"),
		l.svcCtx.DB.Attribute.Create().SetName("port").SetAlias("端口").SetValueType("int"),
		l.svcCtx.DB.Attribute.Create().SetName("version").SetAlias("版本").SetValueType("text"),
		l.svcCtx.DB.Attribute.Create().SetName("model").SetAlias("型号").SetValueType("text"),
		l.svcCtx.DB.Attribute.Create().SetName("serial_number").SetAlias("序列号").SetValueType("text"),
		l.svcCtx.DB.Attribute.Create().SetName("configuration").SetAlias("配置信息").SetValueType("longtext"),
	}

	err = l.svcCtx.DB.Attribute.CreateBulk(attributes...).Exec(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}
	logx.Info("基础属性创建成功")
	return nil
}

// insertSampleCiTypes 插入示例CI类型
func (l *InitDatabaseLogic) insertSampleCiTypes(ctx context.Context) error {
	// 检查是否已存在
	existing, err := l.svcCtx.DB.CiType.Query().Count(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}
	if existing > 0 {
		logx.Info("CI类型已存在，跳过创建")
		return nil
	}

	// 获取name属性作为unique_id
	nameAttr, err := l.svcCtx.DB.Attribute.Query().Where(attribute.NameEQ("name")).First(ctx)
	if err != nil {
		return errorx.NewInternalError("未找到name属性: " + err.Error())
	}

	// 获取分组
	groups, err := l.svcCtx.DB.CiTypeGroup.Query().All(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}

	groupMap := make(map[string]uint64)
	for _, group := range groups {
		groupMap[group.Name] = group.ID
	}

	// 创建CI类型数据
	ciTypeData := []struct {
		name      string
		alias     string
		groupName string
		icon      string
		sort      uint32
	}{
		{"服务器", "物理服务器", "硬件设备", "mdi:server", 1},
		{"交换机", "网络交换机", "硬件设备", "mdi:switch", 2},
		{"应用系统", "业务应用系统", "软件系统", "mdi:application", 1},
		{"MySQL", "MySQL数据库", "数据库", "mdi:database", 1},
	}

	for _, data := range ciTypeData {
		// 创建CI类型
		ciType, err := l.svcCtx.DB.CiType.Create().
			SetName(data.name).
			SetAlias(data.alias).
			SetUniqueID(nameAttr.ID).
			SetIcon(data.icon).
			SetSort(data.sort).
			Save(ctx)

		if err != nil {
			return errorx.NewInternalError("创建CI类型失败: " + err.Error())
		}

		// 创建CI类型分组关联
		if groupId, exists := groupMap[data.groupName]; exists {
			_, err = l.svcCtx.DB.CiTypeGroupItem.Create().
				SetGroupID(groupId).
				SetTypeID(ciType.ID).
				Save(ctx)

			if err != nil {
				logx.Errorf("创建CI类型分组关联失败: %v", err)
			}
		}

		logx.Infof("成功创建CI类型: %s", data.name)
	}

	return nil
}

// insertAttributeGroupsForCiTypes 为CI类型创建属性分组
func (l *InitDatabaseLogic) insertAttributeGroupsForCiTypes(ctx context.Context) error {
	// 获取所有CI类型
	ciTypes, err := l.svcCtx.DB.CiType.Query().All(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}

	// 为每个CI类型创建属性分组
	for _, ciType := range ciTypes {
		// 检查是否已存在属性分组
		existing, err := l.svcCtx.DB.CiTypeAttributeGroup.Query().Where(citypeattributegroup.TypeIDEQ(ciType.ID)).Count(ctx)

		if err != nil {
			return errorx.NewInternalError(err.Error())
		}

		if existing > 0 {
			continue // 已存在，跳过
		}

		// 创建标准属性分组
		attributeGroups := []struct {
			name string
			sort uint32
		}{
			{"基本信息", 1},
			{"技术规格", 2},
			{"管理信息", 3},
		}

		for _, group := range attributeGroups {
			_, err := l.svcCtx.DB.CiTypeAttributeGroup.Create().
				SetName(group.name).
				SetTypeID(ciType.ID).
				SetSort(group.sort).
				Save(ctx)

			if err != nil {
				return errorx.NewInternalError("创建属性分组失败: " + err.Error())
			}
		}

		logx.Infof("为CI类型 %s 创建属性分组成功", ciType.Name)
	}

	return nil
}

// assignAttributesToCiTypes 为CI类型分配属性
func (l *InitDatabaseLogic) assignAttributesToCiTypes(ctx context.Context) error {
	// 检查是否已存在CI类型属性
	existing, err := l.svcCtx.DB.CiTypeAttribute.Query().Count(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}
	if existing > 0 {
		logx.Info("CI类型属性已存在，跳过创建")
		return nil
	}

	// 获取所有CI类型
	ciTypes, err := l.svcCtx.DB.CiType.Query().All(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}

	// 获取所有属性
	attributes, err := l.svcCtx.DB.Attribute.Query().All(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}

	// 为每个CI类型分配属性
	for _, ciType := range ciTypes {
		// 获取该CI类型的属性分组
		attributeGroups, err := l.svcCtx.DB.CiTypeAttributeGroup.Query().
			Where(citypeattributegroup.TypeIDEQ(ciType.ID)).
			All(ctx)
		if err != nil {
			return errorx.NewInternalError(err.Error())
		}

		if len(attributeGroups) == 0 {
			continue
		}

		// 为不同CI类型分配不同的属性
		var attributesToAssign []*ent.Attribute
		switch ciType.Name {
		case "服务器":
			// 服务器分配所有属性
			attributesToAssign = attributes
		case "交换机":
			// 交换机分配网络相关属性
			for _, attr := range attributes {
				if attr.Name == "name" || attr.Name == "description" || attr.Name == "status" ||
					attr.Name == "ip_address" || attr.Name == "port" || attr.Name == "model" ||
					attr.Name == "serial_number" {
					attributesToAssign = append(attributesToAssign, attr)
				}
			}
		case "应用系统":
			// 应用系统分配软件相关属性
			for _, attr := range attributes {
				if attr.Name == "name" || attr.Name == "description" || attr.Name == "status" ||
					attr.Name == "version" || attr.Name == "configuration" || attr.Name == "port" {
					attributesToAssign = append(attributesToAssign, attr)
				}
			}
		case "MySQL":
			// MySQL分配数据库相关属性
			for _, attr := range attributes {
				if attr.Name == "name" || attr.Name == "description" || attr.Name == "status" ||
					attr.Name == "ip_address" || attr.Name == "port" || attr.Name == "version" ||
					attr.Name == "configuration" {
					attributesToAssign = append(attributesToAssign, attr)
				}
			}
		default:
			// 默认分配基础属性
			for _, attr := range attributes {
				if attr.Name == "name" || attr.Name == "description" || attr.Name == "status" {
					attributesToAssign = append(attributesToAssign, attr)
				}
			}
		}

		// 为每个分配的属性创建CI类型属性记录
		for i, attr := range attributesToAssign {
			// 确定属性分组 (简单分配策略)
			var groupID uint64
			if len(attributeGroups) > 0 {
				if attr.Name == "name" || attr.Name == "description" || attr.Name == "status" {
					groupID = attributeGroups[0].ID // 基本信息
				} else if attr.Name == "ip_address" || attr.Name == "port" || attr.Name == "model" || attr.Name == "serial_number" {
					if len(attributeGroups) > 1 {
						groupID = attributeGroups[1].ID // 技术规格
					} else {
						groupID = attributeGroups[0].ID
					}
				} else {
					if len(attributeGroups) > 2 {
						groupID = attributeGroups[2].ID // 管理信息
					} else {
						groupID = attributeGroups[0].ID
					}
				}
			}

			// 创建CI类型属性记录
			_, err := l.svcCtx.DB.CiTypeAttribute.Create().
				SetTypeID(ciType.ID).
				SetAttrID(attr.ID).
				SetIsRequired(attr.Name == "name"). // 名称必填
				SetSort(uint32(i + 1)).
				Save(ctx)

			if err != nil {
				return errorx.NewInternalError("创建CI类型属性失败: " + err.Error())
			}

			// 创建属性分组项记录
			_, err = l.svcCtx.DB.CiTypeAttributeGroupItem.Create().
				SetGroupID(groupID).
				SetAttributeID(attr.ID).
				SetSort(uint32(i + 1)).
				Save(ctx)

			if err != nil {
				return errorx.NewInternalError("创建属性分组项失败: " + err.Error())
			}
		}

		logx.Infof("为CI类型 %s 分配了 %d 个属性", ciType.Name, len(attributesToAssign))
	}

	return nil
}

// insertCiTypeRelations 创建CI类型关系
func (l *InitDatabaseLogic) insertCiTypeRelations(ctx context.Context) error {
	// 检查是否已存在CI类型关系
	existing, err := l.svcCtx.DB.CiTypeRelation.Query().Count(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}

	if existing > 0 {
		logx.Infof("CI类型关系已存在，跳过创建")
		return nil
	}

	// 获取CI类型映射
	ciTypes, err := l.svcCtx.DB.CiType.Query().All(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}

	ciTypeMap := make(map[string]uint64)
	for _, ciType := range ciTypes {
		ciTypeMap[ciType.Name] = ciType.ID
	}

	// 获取关系类型映射
	relationTypes, err := l.svcCtx.DB.RelationType.Query().All(ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}

	relationTypeMap := make(map[string]uint64)
	for _, relationType := range relationTypes {
		relationTypeMap[relationType.Code] = relationType.ID
	}

	// 定义CI类型关系数据
	relationData := []struct {
		parentName   string
		childName    string
		relationCode string
		constraint   string
	}{
		{"服务器", "应用系统", "hosted_on", "ONE_TO_MANY"},    // 服务器承载应用系统
		{"应用系统", "MySQL", "depends_on", "MANY_TO_ONE"}, // 应用系统依赖MySQL
		{"交换机", "服务器", "connect", "ONE_TO_MANY"},       // 交换机连接服务器
		{"MySQL", "服务器", "hosted_on", "MANY_TO_ONE"},   // MySQL部署在服务器上
	}

	for _, data := range relationData {
		parentID, parentExists := ciTypeMap[data.parentName]
		childID, childExists := ciTypeMap[data.childName]
		relationTypeID, relationExists := relationTypeMap[data.relationCode]

		if !parentExists {
			logx.Errorf("未找到父CI类型: %s", data.parentName)
			continue
		}
		if !childExists {
			logx.Errorf("未找到子CI类型: %s", data.childName)
			continue
		}
		if !relationExists {
			logx.Errorf("未找到关系类型: %s", data.relationCode)
			continue
		}

		// 创建CI类型关系
		_, err := l.svcCtx.DB.CiTypeRelation.Create().
			SetParentID(parentID).
			SetChildID(childID).
			SetRelationTypeID(relationTypeID).
			SetConstraint(data.constraint).
			Save(ctx)

		if err != nil {
			logx.Errorf("创建CI类型关系失败: %v", err)
			continue
		}

		logx.Infof("成功创建CI类型关系: %s -[%s]-> %s (%s)",
			data.parentName, data.relationCode, data.childName, data.constraint)
	}

	return nil
}
