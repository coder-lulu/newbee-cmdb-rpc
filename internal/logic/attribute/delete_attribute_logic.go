package attribute

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/attribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/choicefloat"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/choiceinteger"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/choicetext"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeleteAttributeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAttributeLogic {
	return &DeleteAttributeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteAttributeLogic) DeleteAttribute(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. 检查依赖
	count, err := tx.CiTypeAttribute.Query().Where(citypeattribute.AttrIDIn(in.Ids...)).Count(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if count > 0 {
		return nil, status.Error(codes.FailedPrecondition, "有CI类型属性依赖该属性，禁止删除")
	}

	// 2. 级联删除choice表
	_, err = tx.ChoiceText.Delete().Where(choicetext.AttrIDIn(in.Ids...)).Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	_, err = tx.ChoiceInteger.Delete().Where(choiceinteger.AttrIDIn(in.Ids...)).Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	_, err = tx.ChoiceFloat.Delete().Where(choicefloat.AttrIDIn(in.Ids...)).Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 3. 删除属性
	_, err = tx.Attribute.Delete().Where(attribute.IDIn(in.Ids...)).Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &cmdb.BaseResp{Msg: errormsg.DeleteSuccess}, nil
}
