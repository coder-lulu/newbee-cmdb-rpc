package cis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cis"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type CisBatchOperationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCisBatchOperationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CisBatchOperationLogic {
	return &CisBatchOperationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CisBatchOperationLogic) CisBatchOperation(in *cmdb.CisBatchOperationReq) (*cmdb.BaseResp, error) {
	// 开启事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	switch in.Operation {
	case "delete":
		err = l.batchDelete(tx, in.CiIds)
	case "update_status":
		err = l.batchUpdateStatus(tx, in.CiIds, in.Params)
	default:
		return nil, fmt.Errorf("不支持的批量操作类型: %s", in.Operation)
	}

	if err != nil {
		tx.Rollback()
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}

// batchDelete 批量删除CI实例
func (l *CisBatchOperationLogic) batchDelete(tx *ent.Tx, ids []uint64) error {
	if len(ids) == 0 {
		return fmt.Errorf("删除操作需要提供CI实例ID列表")
	}

	_, err := tx.Cis.Delete().Where(cis.IDIn(ids...)).Exec(l.ctx)
	return err
}

// batchUpdateStatus 批量更新CI实例状态
func (l *CisBatchOperationLogic) batchUpdateStatus(tx *ent.Tx, ids []uint64, params *string) error {
	if len(ids) == 0 {
		return fmt.Errorf("状态更新操作需要提供CI实例ID列表")
	}
	if params == nil {
		return fmt.Errorf("状态更新操作需要提供参数")
	}

	// 解析参数中的状态值
	var paramMap map[string]interface{}
	if err := json.Unmarshal([]byte(*params), &paramMap); err != nil {
		return fmt.Errorf("参数格式错误: %v", err)
	}

	statusValue, ok := paramMap["status"]
	if !ok {
		return fmt.Errorf("参数中缺少status字段")
	}

	status, ok := statusValue.(float64)
	if !ok {
		return fmt.Errorf("status字段必须是数字")
	}

	_, err := tx.Cis.Update().
		Where(cis.IDIn(ids...)).
		SetStatus(uint32(status)).
		Save(l.ctx)
	return err
}
