package attribute

import (
	"context"
	"strings"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/attribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAttributeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAttributeLogic {
	return &UpdateAttributeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAttributeLogic) UpdateAttribute(in *cmdb.AttributeInfo) (*cmdb.BaseIDResp, error) {
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	alias := strings.TrimSpace(*in.Alias)
	if alias == "" {
		return nil, status.Error(codes.InvalidArgument, "别名不能为空")
	}

	// 查询原始数据
	_, err = tx.Attribute.Query().Where(attribute.IDEQ(*in.Id)).First(l.ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, "属性不存在")
	}

	// 禁止更新name、valueType、isList
	// 其他字段可更新
	update := AttributeUpdateBuilderSetter(tx.Attribute.UpdateOneID(*in.Id), in, &alias)

	_, err = update.Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	// 如果属性时choice类型，则最少choices数量是1个，如果少于一个则拒绝更新
	if *in.IsChoice {
		if len(in.Choices) < 1 {
			return nil, status.Error(codes.InvalidArgument, "属性是choice类型，最少需要1个选项")
		}
		// 查询所有choice 进行比对，如果存在则更新
		// 根据attr的value判断是哪种choice类型
		attr, err := tx.Attribute.Get(l.ctx, *in.Id)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		err = CreateOrUpdateChoices(l.ctx, tx, &attr.ID, in.ValueType, in.Choices)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &cmdb.BaseIDResp{Id: *in.Id, Msg: errormsg.UpdateSuccess}, nil
}
