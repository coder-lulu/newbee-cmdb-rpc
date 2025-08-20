package choicefloat

import (
	"context"
	"time"

	"gitee.com/link234/cmdb-rpc/ent/choicefloat"
	"gitee.com/link234/cmdb-rpc/ent/predicate"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetChoiceFloatListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChoiceFloatListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChoiceFloatListLogic {
	return &GetChoiceFloatListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetChoiceFloatListLogic) GetChoiceFloatList(in *cmdb.ChoiceFloatListReq) (*cmdb.ChoiceFloatListResp, error) {
	var predicates []predicate.ChoiceFloat
	if in.CreatedAt != nil {
		predicates = append(predicates, choicefloat.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, choicefloat.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, choicefloat.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.AttrId != nil {
		predicates = append(predicates, choicefloat.AttrIDEQ(*in.AttrId))
	}
	if in.Value != nil {
		predicates = append(predicates, choicefloat.ValueEQ(*in.Value))
	}

	result, err := l.svcCtx.DB.ChoiceFloat.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.ChoiceFloatListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		option := &cmdb.AttributeChoiceItemMeta{
			Label: v.Option.Label,
			Icon:  v.Option.Icon,
			Style: &cmdb.AttributeFontOption{
				Color:          &v.Option.Style.Color,
				BgColor:        &v.Option.Style.BgColor,
				FontStyle:      &v.Option.Style.FontStyle,
				FontWeight:     &v.Option.Style.FontWeight,
				TextDecoration: &v.Option.Style.TextDecoration,
			},
		}

		resp.Data = append(resp.Data, &cmdb.ChoiceFloatInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			AttrId:    &v.AttrID,
			Value:     &v.Value,
			Option:    option,
		})
	}

	return resp, nil
}
