package citypegroup

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypegroupitem"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/coder-lulu/newbee-common/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCiTypeGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCiTypeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCiTypeGroupLogic {
	return &DeleteCiTypeGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCiTypeGroupLogic) DeleteCiTypeGroup(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	// 查询分组下的CItype，如果分组下有CItype，则不能删除
	ciTypeGroupItem, err := l.svcCtx.DB.CiTypeGroupItem.Query().Where(citypegroupitem.GroupIDIn(in.Ids...)).Count(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if ciTypeGroupItem > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "分组下有CItype，不能删除")
	}

	// 如果分组名称为 基类 其他 则不允许删除
	ciTypeGroup, err := l.svcCtx.DB.CiTypeGroup.Query().Where(citypegroup.IDIn(in.Ids...)).First(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if ciTypeGroup.Name == "基类" || ciTypeGroup.Name == "其它" {
		return nil, status.Errorf(codes.InvalidArgument, "基类和其它不能删除")
	}

	_, err = l.svcCtx.DB.CiTypeGroup.Delete().Where(citypegroup.IDIn(in.Ids...)).Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.DeleteSuccess}, nil
}
