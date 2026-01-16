package choicetext

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/schema"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChoiceTextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChoiceTextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChoiceTextLogic {
	return &UpdateChoiceTextLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateChoiceTextLogic) UpdateChoiceText(in *cmdb.ChoiceTextInfo) (*cmdb.BaseResp, error) {
	// 构造 option 结构体
	var optionS schema.ChoiceItemMetaS
	if in.Option != nil {
		optionS = schema.ChoiceItemMetaS{
			Label: in.Option.Label,
			Icon:  in.Option.Icon,
		}
		if in.Option.Style != nil {
			optionS.Style = schema.FontOptionS{
				Color:          safeStringValue(in.Option.Style.Color),
				BgColor:        safeStringValue(in.Option.Style.BgColor),
				FontStyle:      safeStringValue(in.Option.Style.FontStyle),
				FontWeight:     safeStringValue(in.Option.Style.FontWeight),
				TextDecoration: safeStringValue(in.Option.Style.TextDecoration),
			}
		}
	}

	err := l.svcCtx.DB.ChoiceText.UpdateOneID(*in.Id).
		SetNotNilAttrID(in.AttrId).
		SetNotNilValue(in.Value).
		SetNotNilOption(&optionS).
		Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
