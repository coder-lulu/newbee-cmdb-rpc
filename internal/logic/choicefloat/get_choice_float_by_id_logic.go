package choicefloat

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetChoiceFloatByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChoiceFloatByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChoiceFloatByIdLogic {
	return &GetChoiceFloatByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetChoiceFloatByIdLogic) GetChoiceFloatById(in *cmdb.IDReq) (*cmdb.ChoiceFloatInfo, error) {
	result, err := l.svcCtx.DB.ChoiceFloat.Get(l.ctx, in.Id)
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

	return &cmdb.ChoiceFloatInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		AttrId:    &result.AttrID,
		Value:     &result.Value,
		Option:    option,
	}, nil
}
