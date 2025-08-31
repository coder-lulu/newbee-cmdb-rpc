package choiceinteger

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/choiceinteger"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetChoiceIntegerListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChoiceIntegerListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChoiceIntegerListLogic {
	return &GetChoiceIntegerListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetChoiceIntegerListLogic) GetChoiceIntegerList(in *cmdb.ChoiceIntegerListReq) (*cmdb.ChoiceIntegerListResp, error) {
	var predicates []predicate.ChoiceInteger
	if in.CreatedAt != nil {
		predicates = append(predicates, choiceinteger.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, choiceinteger.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, choiceinteger.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.AttrId != nil {
		predicates = append(predicates, choiceinteger.AttrIDEQ(*in.AttrId))
	}
	if in.Value != nil {
		predicates = append(predicates, choiceinteger.ValueEQ(int(*in.Value)))
	}

	result, err := l.svcCtx.DB.ChoiceInteger.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.ChoiceIntegerListResp{}
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

		resp.Data = append(resp.Data, &cmdb.ChoiceIntegerInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			AttrId:    &v.AttrID,
			Value:     pointy.GetPointer(int64(v.Value)),
			Option:    option,
		})
	}

	return resp, nil
}
