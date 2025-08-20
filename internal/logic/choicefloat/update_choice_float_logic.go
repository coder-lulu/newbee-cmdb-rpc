package choicefloat

import (
	"context"

	"gitee.com/link234/cmdb-rpc/ent/schema"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChoiceFloatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChoiceFloatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChoiceFloatLogic {
	return &UpdateChoiceFloatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateChoiceFloatLogic) UpdateChoiceFloat(in *cmdb.ChoiceFloatInfo) (*cmdb.BaseResp, error) {
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

	err := l.svcCtx.DB.ChoiceFloat.UpdateOneID(*in.Id).
		SetNotNilAttrID(in.AttrId).
		SetNotNilValue(in.Value).
		SetNotNilOption(&optionS).
		Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
