package attribute

import (
	"context"
	"strings"

	"gitee.com/link234/cmdb-rpc/ent/attribute"

	"gitee.com/link234/cmdb-rpc/internal/consts"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"gitee.com/link234/newbee-backend-common/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateAttributeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAttributeLogic {
	return &CreateAttributeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateAttributeLogic) CreateAttribute(in *cmdb.AttributeInfo) (*cmdb.BaseIDResp, error) {
	// 启用事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	name := strings.TrimSpace(*in.Name)
	alias := strings.TrimSpace(*in.Alias)
	//效验参数
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "名称不能为空")
	}
	if alias == "" {
		return nil, status.Error(codes.InvalidArgument, "别名不能为空")
	}
	if in.ValueType == nil || *in.ValueType == "" {
		valueType := consts.ValueTypeShortText
		in.ValueType = &valueType
	}
	// 检查是否存在同名的属性
	exist, err := tx.Attribute.Query().Where(attribute.NameEQ(name)).Exist(l.ctx)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, status.Error(codes.AlreadyExists, "属性名称已存在")
	}
	// 使用通用构造器
	result, err := AttributeCreateBuilderSetter(tx.Attribute.Create(), in).Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 创建选项
	if in.IsChoice != nil && *in.IsChoice {
		err = CreateOrUpdateChoices(l.ctx, tx, &result.ID, in.ValueType, in.Choices)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
