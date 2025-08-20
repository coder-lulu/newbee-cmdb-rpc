package attribute

import (
	"context"
	"fmt"

	"gitee.com/link234/cmdb-rpc/ent/choicefloat"
	"gitee.com/link234/cmdb-rpc/ent/choiceinteger"
	"gitee.com/link234/cmdb-rpc/ent/choicetext"
	"gitee.com/link234/cmdb-rpc/ent/schema"
	"gitee.com/link234/cmdb-rpc/internal/consts"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetAttributeByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAttributeByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAttributeByIdLogic {
	return &GetAttributeByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAttributeByIdLogic) GetAttributeById(in *cmdb.IDReq) (*cmdb.AttributeInfo, error) {
	// 1. 查询 Attribute 记录
	result, err := l.svcCtx.DB.Attribute.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 2. 查询 Choices（根据 valueType 判断查哪个表）
	var choices []*cmdb.AttributeChoiceItem
	switch result.ValueType {
	case consts.ValueTypeShortText, consts.ValueTypeLongText,
		consts.ValueTypeDateTime, consts.ValueTypeDate, consts.ValueTypeTime, consts.ValueTypeLink, consts.ValueTypeImage:
		choiceList, err := l.svcCtx.DB.ChoiceText.Query().Where(choicetext.AttrID(result.ID)).All(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		for _, c := range choiceList {
			choices = append(choices, l.parseChoiceOption(c.Option, c.Value))
		}
	case consts.ValueTypeInt:
		choiceList, err := l.svcCtx.DB.ChoiceInteger.Query().Where(choiceinteger.AttrID(result.ID)).All(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		for _, c := range choiceList {
			choices = append(choices, l.parseChoiceOption(c.Option, c.Value))
		}
	case consts.ValueTypeFloat:
		choiceList, err := l.svcCtx.DB.ChoiceFloat.Query().Where(choicefloat.AttrID(result.ID)).All(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		for _, c := range choiceList {
			choices = append(choices, l.parseChoiceOption(c.Option, c.Value))
		}
	}

	// 3. 组装 AttributeInfo
	attrInfo := AttributeEntToProto(result, choices)
	return attrInfo, nil
}

func (l *GetAttributeByIdLogic) parseChoiceOption(option schema.ChoiceItemMetaS, value interface{}) *cmdb.AttributeChoiceItem {
	meta := &cmdb.AttributeChoiceItemMeta{
		Label: option.Label,
		Icon:  option.Icon,
		Style: &cmdb.AttributeFontOption{
			Color:          &option.Style.Color,
			BgColor:        &option.Style.BgColor,
			FontStyle:      &option.Style.FontStyle,
			FontWeight:     &option.Style.FontWeight,
			TextDecoration: &option.Style.TextDecoration,
		},
	}

	return &cmdb.AttributeChoiceItem{
		Value: fmt.Sprintf("%v", value),
		Meta:  meta,
	}
}
