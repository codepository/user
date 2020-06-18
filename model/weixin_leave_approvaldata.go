package model

import (
	"errors"

	"github.com/mumushuiding/util"
)

// WeixinApprovaldataTable WeixinApprovaldataTable
var WeixinApprovaldataTable = "weixin_leave_approvaldata"

// WeixinApprovaldata 流程数据
type WeixinApprovaldata struct {
	ID int `gorm:"primary_key" json:"id,omitempty"`
	// 客户端id
	Agentid int `json:"agentid"`
	// 审批单号20位
	ThirdNo string `gorm:"size:20;column:thirdNo" json:"thirdNo"`
	// 执行流数据
	Data string `json:"data"`
	// 当前步骤
	Step int `json:"step"`
	// 1-审批中；2-已通过；3-已驳回；4-已取消
	Status int `json:"status"`
	// 0-提交申请时，1-审批通过时，3-提交和审批者抄送
	NotifyAttr int `gorm:"column:notifyAttr" json:"notifyAttr"`
}

// Approvaldata 审批流数据
type Approvaldata struct {
	Errcode int            `json:"errcode"`
	Errmsg  string         `json:"errmsg"`
	Data    *ExecutionData `json:"data"`
}

// ToString ToString
func (a *Approvaldata) ToString() string {
	str, _ := util.ToJSONStr(a)
	return str
}

// SaveOrUpdate SaveOrUpdate
func (w *WeixinApprovaldata) SaveOrUpdate() error {
	if len(w.ThirdNo) == 0 {
		return errors.New("流水号 ThirdNo 不能为空")
	}
	return wxdb.Table(WeixinApprovaldataTable).Where("thirdNo=?", w.ThirdNo).Assign(*w).FirstOrCreate(w).Error
}
