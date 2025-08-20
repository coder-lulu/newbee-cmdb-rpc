package attribute

import (
	"context"
	"fmt"

	"gitee.com/link234/cmdb-rpc/ent/attribute"
	"gitee.com/link234/cmdb-rpc/ent/choicefloat"
	"gitee.com/link234/cmdb-rpc/ent/choiceinteger"
	"gitee.com/link234/cmdb-rpc/ent/choicetext"
	"gitee.com/link234/cmdb-rpc/ent/predicate"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetAttributeListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAttributeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAttributeListLogic {
	return &GetAttributeListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAttributeListLogic) GetAttributeList(in *cmdb.AttributeListReq) (*cmdb.AttributeListResp, error) {
	var predicates []predicate.Attribute
	if in.Name != nil {
		predicates = append(predicates, attribute.NameContains(*in.Name))
	}
	if in.Alias != nil {
		predicates = append(predicates, attribute.AliasContains(*in.Alias))
	}

	result, err := l.svcCtx.DB.Attribute.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.AttributeListResp{}
	resp.Total = result.PageDetails.Total

	// 1. 批量查出所有属性ID
	attrIDs := make([]uint64, 0, len(result.List))
	for _, v := range result.List {
		attrIDs = append(attrIDs, v.ID)
	}

	// 2. 批量查出所有Choices
	choicesMap := make(map[uint64][]*cmdb.AttributeChoiceItem)
	// 查文本选项
	textChoices, _ := l.svcCtx.DB.ChoiceText.Query().Where(choicetext.AttrIDIn(attrIDs...)).All(l.ctx)
	for _, c := range textChoices {
		meta := &cmdb.AttributeChoiceItemMeta{
			Label: c.Option.Label,
			Icon:  c.Option.Icon,
			Style: &cmdb.AttributeFontOption{
				Color:          &c.Option.Style.Color,
				BgColor:        &c.Option.Style.BgColor,
				FontStyle:      &c.Option.Style.FontStyle,
				FontWeight:     &c.Option.Style.FontWeight,
				TextDecoration: &c.Option.Style.TextDecoration,
			},
		}
		choicesMap[c.AttrID] = append(choicesMap[c.AttrID], &cmdb.AttributeChoiceItem{
			Value: fmt.Sprintf("%v", c.Value),
			Meta:  meta,
		})
	}
	// 查整数选项
	intChoices, _ := l.svcCtx.DB.ChoiceInteger.Query().Where(choiceinteger.AttrIDIn(attrIDs...)).All(l.ctx)
	for _, c := range intChoices {
		meta := &cmdb.AttributeChoiceItemMeta{
			Label: c.Option.Label,
			Icon:  c.Option.Icon,
			Style: &cmdb.AttributeFontOption{
				Color:          &c.Option.Style.Color,
				BgColor:        &c.Option.Style.BgColor,
				FontStyle:      &c.Option.Style.FontStyle,
				FontWeight:     &c.Option.Style.FontWeight,
				TextDecoration: &c.Option.Style.TextDecoration,
			},
		}
		choicesMap[c.AttrID] = append(choicesMap[c.AttrID], &cmdb.AttributeChoiceItem{
			Value: fmt.Sprintf("%v", c.Value),
			Meta:  meta,
		})
	}
	// 查浮点选项
	floatChoices, _ := l.svcCtx.DB.ChoiceFloat.Query().Where(choicefloat.AttrIDIn(attrIDs...)).All(l.ctx)
	for _, c := range floatChoices {
		meta := &cmdb.AttributeChoiceItemMeta{
			Label: c.Option.Label,
			Icon:  c.Option.Icon,
			Style: &cmdb.AttributeFontOption{
				Color:          &c.Option.Style.Color,
				BgColor:        &c.Option.Style.BgColor,
				FontStyle:      &c.Option.Style.FontStyle,
				FontWeight:     &c.Option.Style.FontWeight,
				TextDecoration: &c.Option.Style.TextDecoration,
			},
		}
		choicesMap[c.AttrID] = append(choicesMap[c.AttrID], &cmdb.AttributeChoiceItem{
			Value: fmt.Sprintf("%v", c.Value),
			Meta:  meta,
		})
	}

	// 3. 组装返回
	for _, v := range result.List {
		choices := choicesMap[v.ID]
		resp.Data = append(resp.Data, AttributeEntToProto(v, choices))
	}

	return resp, nil
}
