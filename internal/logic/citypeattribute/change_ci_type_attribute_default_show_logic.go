package citypeattribute

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeCiTypeAttributeDefaultShowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeCiTypeAttributeDefaultShowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeCiTypeAttributeDefaultShowLogic {
	return &ChangeCiTypeAttributeDefaultShowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChangeCiTypeAttributeDefaultShowLogic) ChangeCiTypeAttributeDefaultShow(in *cmdb.CiTypeAttributeChangeDefaultShowReq) (*cmdb.BaseResp, error) {
	cit, err := l.svcCtx.DB.CiTypeAttribute.Get(l.ctx, in.CiTypeAttributeId)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.DB.CiTypeAttribute.UpdateOneID(cit.ID).SetListShow(in.ListShow).Save(l.ctx)
	if err != nil {
		return nil, err
	}
	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
