package choiceinteger

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetChoiceIntegerByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChoiceIntegerByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChoiceIntegerByIdLogic {
	return &GetChoiceIntegerByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetChoiceIntegerByIdLogic) GetChoiceIntegerById(in *cmdb.IDReq) (*cmdb.ChoiceIntegerInfo, error) {
	result, err := l.svcCtx.DB.ChoiceInteger.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 转换 option 结构体为 proto 结构体
	option := &cmdb.AttributeChoiceItemMeta{
		Label: result.Option.Label,
		Icon:  result.Option.Icon,
		Style: &cmdb.AttributeFontOption{
			Color:          &result.Option.Style.Color,
			BgColor:        &result.Option.Style.BgColor,
			FontStyle:      &result.Option.Style.FontStyle,
			FontWeight:     &result.Option.Style.FontWeight,
			TextDecoration: &result.Option.Style.TextDecoration,
		},
	}

	return &cmdb.ChoiceIntegerInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		AttrId:    &result.AttrID,
		Value:     pointy.GetPointer(int64(result.Value)),
		Option:    option,
	}, nil
}
