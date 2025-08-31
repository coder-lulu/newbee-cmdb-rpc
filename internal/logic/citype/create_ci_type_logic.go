package citype

import (
	"context"
	"errors"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/attribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citype"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/schema"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/msg/errormsg"
	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/coder-lulu/newbee-common/utils/uuidx"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCiTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiTypeLogic {
	return &CreateCiTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCiTypeLogic) CreateCiType(in *cmdb.CiTypeInfo) (*cmdb.BaseIDResp, error) {
	// 开启事务
	tx, err := l.svcCtx.DB.BeginTx(l.ctx, nil)
	if err != nil {
		l.Logger.Errorf("开启事务失败: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	// 根据唯一ID查询属性是否存在
	isExist, err := tx.Attribute.Query().Where(attribute.ID(*in.UniqueId)).Exist(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询属性失败: %v", err)
		return nil, err
	}
	if !isExist {
		l.Logger.Errorf("属性唯一ID不存在: %d", *in.UniqueId)
		return nil, errors.New("属性唯一ID不存在")
	}

	isExist, err = tx.CiType.Query().Where(citype.Or(citype.NameEQ(*in.Name), citype.AliasEQ(*in.Alias))).Exist(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询CiType是否存在失败: %v", err)
		return nil, err
	}
	if isExist {
		l.Logger.Errorf("存在名称相同或别名相同的CiType: name=%s, alias=%s", *in.Name, *in.Alias)
		return nil, errors.New("存在名称相同或别名相同的CiType,添加失败")
	}

	// 验证继承模型列表
	if len(in.InheritedModels) > 0 {
		for _, parentId := range in.InheritedModels {
			exists, err := tx.CiType.Query().Where(citype.IDEQ(parentId)).Exist(l.ctx)
			if err != nil {
				l.Logger.Errorf("查询父模型失败: %v", err)
				return nil, err
			}
			if !exists {
				l.Logger.Errorf("父模型不存在: ID=%d", parentId)
				return nil, errors.New("继承的父模型不存在")
			}
		}
	}

	if in.ShowId == nil {
		in.ShowId = in.UniqueId
	}
	// 创建ci type
	query := tx.CiType.Create().
		SetNotNilSort(in.Sort).
		SetNotNilName(in.Name).
		SetNotNilAlias(in.Alias).
		SetNotNilUniqueID(in.UniqueId).
		SetNotNilCreatedBy(uuidx.ParseUUIDStringToPointer(in.CreatedBy)).
		SetNotNilIcon(in.Icon).
		SetNotNilShowID(in.ShowId).
		SetNotNilDefaultOrderAttrID(in.DefaultOrderAttr).
		SetNotNilIsInherited(in.IsInherited)

	if in.Status != nil {
		query.SetNotNilStatus(pointy.GetPointer(uint8(*in.Status)))
	}

	// 设置是否继承标志
	if in.IsInherited != nil {
		query.SetNotNilIsInherited(in.IsInherited)
	}

	// 处理唯一性约束
	if len(in.UniqueConst) > 0 {
		uniqueConsts := []schema.CiTypeUniqueConstS{}
		for _, v := range in.UniqueConst {
			uniqueConsts = append(uniqueConsts, schema.CiTypeUniqueConstS{
				AttrIds: v.AttrIds,
			})
		}
		query.SetUniqueConst(uniqueConsts)
	}

	result, err := query.Save(l.ctx)
	if err != nil {
		l.Logger.Errorf("创建CiType失败: %v", err)
		return nil, err
	}

	// 处理继承关系
	if in.IsInherited != nil && *in.IsInherited && len(in.InheritedModels) > 0 {
		l.Logger.Infof("开始创建继承关系: 子模型ID=%d, 父模型列表=%v", result.ID, in.InheritedModels)

		for _, parentId := range in.InheritedModels {
			_, err := tx.CiTypeInheritance.Create().
				SetParentID(parentId).
				SetChildID(result.ID).
				Save(l.ctx)
			if err != nil {
				l.Logger.Errorf("创建继承关系失败: parent_id=%d, child_id=%d, error=%v", parentId, result.ID, err)
				return nil, err
			}
			l.Logger.Infof("成功创建继承关系: parent_id=%d, child_id=%d", parentId, result.ID)
		}
	}

	if in.GroupId != nil {
		groupItemQuery := tx.CiTypeGroupItem.Create().
			SetNotNilGroupID(pointy.GetPointer(*in.GroupId)).
			SetNotNilTypeID(pointy.GetPointer(result.ID))
		_, err = groupItemQuery.Save(l.ctx)
		if err != nil {
			l.Logger.Errorf("添加到模型分组失败: %v", err)
			return nil, err
		}
	} else {
		// 将citype添加到模型分组，默认分组
		group, err := tx.CiTypeGroup.Query().Where(citypegroup.NameEQ("其它")).First(l.ctx)
		if err != nil {
			l.Logger.Errorf("查询默认分组失败: %v", err)
			return nil, err
		}
		groupItemQuery := tx.CiTypeGroupItem.Create().
			SetNotNilGroupID(pointy.GetPointer(group.ID)).
			SetNotNilTypeID(pointy.GetPointer(result.ID))
		_, err = groupItemQuery.Save(l.ctx)
		if err != nil {
			l.Logger.Errorf("添加到模型分组失败: %v", err)
			return nil, err
		}
	}

	// 创建默认分组
	groupQuery := tx.CiTypeAttributeGroup.Create().
		SetNotNilName(pointy.GetPointer("其它")).
		SetNotNilTypeID(pointy.GetPointer(result.ID)).
		SetNotNilSort(pointy.GetPointer(uint32(999)))

	groupResult, err := groupQuery.Save(l.ctx)
	if err != nil {
		l.Logger.Errorf("创建默认属性分组失败: %v", err)
		return nil, err
	}

	// 创建默认分组属性
	attrQuery := tx.CiTypeAttributeGroupItem.Create().
		SetNotNilAttrID(in.UniqueId).
		SetNotNilGroupID(pointy.GetPointer(groupResult.ID))
	_, err = attrQuery.Save(l.ctx)
	if err != nil {
		l.Logger.Errorf("添加默认分组属性失败: %v", err)
		return nil, err
	}

	// 添加唯一属性到CiTypeAttribute
	ciAttrQuery := tx.CiTypeAttribute.Create().
		SetNotNilTypeID(pointy.GetPointer(result.ID)).
		SetNotNilAttrID(in.UniqueId).
		SetNotNilIsRequired(pointy.GetPointer(true)).
		SetNotNilSort(pointy.GetPointer(uint32(1)))
	_, err = ciAttrQuery.Save(l.ctx)
	if err != nil {
		l.Logger.Errorf("添加唯一属性到CiTypeAttribute失败: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		l.Logger.Errorf("提交事务失败: %v", err)
		return nil, err
	}

	l.Logger.Infof("创建CiType成功: ID=%d, Name=%s, 继承关系数量=%d", result.ID, *in.Name, len(in.InheritedModels))
	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
