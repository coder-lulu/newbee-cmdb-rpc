package relationtype

import (
	"context"
	"encoding/json"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/relationtype"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRelationTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRelationTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRelationTypeLogic {
	return &CreateRelationTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateRelationTypeLogic) CreateRelationType(in *cmdb.RelationTypeInfo) (*cmdb.BaseIDResp, error) {
	creator := l.svcCtx.DB.RelationType.Create().
		SetNotNilName(in.Name).
		SetNotNilCode(in.Code)

	// 处理category枚举
	if in.Category != nil {
		creator.SetCategory(relationtype.Category(*in.Category))
	}

	// 处理direction枚举
	if in.Direction != nil {
		creator.SetDirection(relationtype.Direction(*in.Direction))
	}

	// 处理新增字段
	if in.Description != nil {
		creator.SetDescription(*in.Description)
	}

	if in.IsStandard != nil {
		creator.SetIsStandard(*in.IsStandard)
	}

	if in.SortOrder != nil {
		creator.SetSortOrder(int(*in.SortOrder))
	}

	if in.IsEnabled != nil {
		creator.SetIsEnabled(*in.IsEnabled)
	}

	// 处理拓扑显示字段
	if in.DisplayColor != nil {
		creator.SetDisplayColor(*in.DisplayColor)
	}

	if in.LineType != nil {
		creator.SetLineType(relationtype.LineType(*in.LineType))
	}

	if in.Icon != nil {
		creator.SetIcon(*in.Icon)
	}

	if in.Weight != nil {
		creator.SetWeight(int(*in.Weight))
	}

	if in.DisplayLabel != nil {
		creator.SetDisplayLabel(*in.DisplayLabel)
	}

	if in.TooltipTemplate != nil {
		creator.SetTooltipTemplate(*in.TooltipTemplate)
	}

	if in.DisplayStyle != nil {
		var styleMap map[string]interface{}
		if err := json.Unmarshal([]byte(*in.DisplayStyle), &styleMap); err != nil {
			l.Errorf("解析DisplayStyle JSON失败: %v", err)
			styleMap = make(map[string]interface{})
		}
		creator.SetDisplayStyle(styleMap)
	}

	result, err := creator.Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
