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

type CreateChoiceFloatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChoiceFloatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChoiceFloatLogic {
	return &CreateChoiceFloatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateChoiceFloatLogic) CreateChoiceFloat(in *cmdb.ChoiceFloatInfo) (*cmdb.BaseIDResp, error) {
	result, err := l.svcCtx.DB.ChoiceFloat.Create().
		SetNotNilAttrID(in.AttrId).
		SetNotNilValue(in.Value).
		SetNotNilOption(&schema.ChoiceItemMetaS{
			Label: in.Option.Label,
			Icon:  in.Option.Icon,
			Style: schema.FontOptionS{
				Color:          safeStringValue(in.Option.Style.Color),
				BgColor:        safeStringValue(in.Option.Style.BgColor),
				FontStyle:      safeStringValue(in.Option.Style.FontStyle),
				FontWeight:     safeStringValue(in.Option.Style.FontWeight),
				TextDecoration: safeStringValue(in.Option.Style.TextDecoration),
			},
		}).
		Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}

// 辅助函数：安全获取字符串值
func safeStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
