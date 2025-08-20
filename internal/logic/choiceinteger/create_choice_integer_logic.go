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

type CreateChoiceIntegerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChoiceIntegerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChoiceIntegerLogic {
	return &CreateChoiceIntegerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateChoiceIntegerLogic) CreateChoiceInteger(in *cmdb.ChoiceIntegerInfo) (*cmdb.BaseIDResp, error) {
	// 构造 option 结构体
	var optionS schema.ChoiceItemMetaS
	if in.Option != nil {
		optionS = schema.ChoiceItemMetaS{
			Label: in.Option.Label,
			Icon:  in.Option.Icon,
		}
		if in.Option.Style != nil {
			optionS.Style = schema.FontOptionS{
				Color:          getStringValue(in.Option.Style.Color),
				BgColor:        getStringValue(in.Option.Style.BgColor),
				FontStyle:      getStringValue(in.Option.Style.FontStyle),
				FontWeight:     getStringValue(in.Option.Style.FontWeight),
				TextDecoration: getStringValue(in.Option.Style.TextDecoration),
			}
		}
	}

	query := l.svcCtx.DB.ChoiceInteger.Create().
		SetNotNilAttrID(in.AttrId).
		SetNotNilOption(&optionS)

	if in.Value != nil {
		query.SetNotNilValue(pointy.GetPointer(int(*in.Value)))
	}

	result, err := query.Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}

// 辅助函数：安全获取字符串值
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
