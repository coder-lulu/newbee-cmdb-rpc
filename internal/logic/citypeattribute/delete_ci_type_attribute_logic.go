package citypeattribute

import (
	"context"

	"gitee.com/link234/cmdb-rpc/ent/citypeattribute"
	"gitee.com/link234/cmdb-rpc/ent/citypeattributegroup"
	"gitee.com/link234/cmdb-rpc/ent/citypeattributegroupitem"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCiTypeAttributeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCiTypeAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCiTypeAttributeLogic {
	return &DeleteCiTypeAttributeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCiTypeAttributeLogic) DeleteCiTypeAttribute(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. 查询所有分组下的attr_id
	ciTypeAttrs, err := tx.CiTypeAttribute.Query().
		Where(citypeattribute.IDIn(in.Ids...)).
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	ciTypeIds := make([]uint64, 0)
	for _, cta := range ciTypeAttrs {
		ciTypeIds = append(ciTypeIds, cta.TypeID)
	}

	// 查询每个typeID下的分组
	groups, err := tx.CiTypeAttributeGroup.Query().Where(citypeattributegroup.TypeIDIn(ciTypeIds...)).All(l.ctx)
	if err != nil {
		return nil, err
	}

	attrIDs := make([]uint64, 0)
	for _, cta := range ciTypeAttrs {
		attrIDs = append(attrIDs, cta.AttrID)
	}

	// 遍历每个分组，删除分组ID
	for _, group := range groups {
		_, err = tx.CiTypeAttributeGroupItem.Delete().Where(citypeattributegroupitem.GroupID(group.ID), citypeattributegroupitem.AttrIDIn(attrIDs...)).Exec(l.ctx)
		if err != nil {
			return nil, err
		}
	}

	// // 2. 查询每个attr的valueType
	// attrs, err := tx.Attribute.Query().Where(attribute.IDIn(attrIDs...)).All(l.ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// // 3. 按valueType删除value_xxx表
	// for _, attr := range attrs {
	// 	switch attr.ValueType {
	// 	case consts.ValueTypeShortText, consts.ValueTypeLongText:
	// 		_, err = tx.ValueText.Delete().Where(valuetext.AttrID(attr.ID)).Exec(l.ctx)
	// 		_, err = tx.ValueIndexText.Delete().Where(valueindextext.AttrID(attr.ID)).Exec(l.ctx)
	// 	case consts.ValueTypeInt:
	// 		_, err = tx.ValueInteger.Delete().Where(valueinteger.AttrID(attr.ID)).Exec(l.ctx)
	// 		_, err = tx.ValueIndexInteger.Delete().Where(valueindexinteger.AttrID(attr.ID)).Exec(l.ctx)
	// 	case consts.ValueTypeFloat:
	// 		_, err = tx.ValueFloat.Delete().Where(valuefloat.AttrID(attr.ID)).Exec(l.ctx)
	// 		_, err = tx.ValueIndexFloat.Delete().Where(valueindexfloat.AttrID(attr.ID)).Exec(l.ctx)
	// 	case consts.ValueTypeJSON:
	// 		_, err = tx.ValueJSON.Delete().Where(valuejson.AttrID(attr.ID)).Exec(l.ctx)
	// 	case consts.ValueTypeDateTime, consts.ValueTypeDate, consts.ValueTypeTime:
	// 		_, err = tx.ValueDatetime.Delete().Where(valuedatetime.AttrID(attr.ID)).Exec(l.ctx)
	// 		_, err = tx.ValueIndexDatetime.Delete().Where(valueindexdatetime.AttrID(attr.ID)).Exec(l.ctx)
	// 		// 其他类型可按需扩展
	// 	}
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	// 5. 删除CiTypeAttribute表
	_, err = tx.CiTypeAttribute.Delete().Where(citypeattribute.IDIn(in.Ids...)).Exec(l.ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &cmdb.BaseResp{Msg: "删除成功"}, nil
}
