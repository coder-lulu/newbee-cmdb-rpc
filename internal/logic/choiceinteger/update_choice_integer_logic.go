package choiceinteger

import (
	"context"

	"gitee.com/link234/cmdb-rpc/ent/schema"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"
	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChoiceIntegerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChoiceIntegerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChoiceIntegerLogic {
	return &UpdateChoiceIntegerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateChoiceIntegerLogic) UpdateChoiceInteger(in *cmdb.ChoiceIntegerInfo) (*cmdb.BaseResp, error) {
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

	query := l.svcCtx.DB.ChoiceInteger.UpdateOneID(*in.Id).
		SetNotNilAttrID(in.AttrId).
		SetNotNilOption(&optionS)

	if in.Value != nil {
		query.SetNotNilValue(pointy.GetPointer(int(*in.Value)))
	}

	err := query.Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}

// 辅助函数：安全获取字符串值
func safeStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
