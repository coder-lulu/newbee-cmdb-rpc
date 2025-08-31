package choicetext

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/choicetext"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetChoiceTextListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChoiceTextListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChoiceTextListLogic {
	return &GetChoiceTextListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetChoiceTextListLogic) GetChoiceTextList(in *cmdb.ChoiceTextListReq) (*cmdb.ChoiceTextListResp, error) {
	var predicates []predicate.ChoiceText
	if in.CreatedAt != nil {
		predicates = append(predicates, choicetext.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, choicetext.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, choicetext.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.AttrId != nil {
		predicates = append(predicates, choicetext.AttrIDEQ(*in.AttrId))
	}
	if in.Value != nil {
		predicates = append(predicates, choicetext.ValueContains(*in.Value))
	}

	result, err := l.svcCtx.DB.ChoiceText.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.ChoiceTextListResp{}
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

		resp.Data = append(resp.Data, &cmdb.ChoiceTextInfo{
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
