package model

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/mumushuiding/util"
)

// WeixinFlowApprovaldataTable WeixinFlowApprovaldataTable
var WeixinFlowApprovaldataTable = "weixin_flow_approvaldata"

// WeixinFlowApprovaldata 流程数据
type WeixinFlowApprovaldata struct {
	ID int `gorm:"primary_key" json:"id,omitempty"`
	// 客户端id
	Agentid int `json:"agentid"`
	// 审批单号20位
	ThirdNo string `gorm:"size:20;column:thirdNo" json:"thirdNo"`
	// 执行流数据
	Data string `gorm:"type:text" json:"data"`
	// 当前步骤
	Step int `json:"step"`
	// 1-审批中；2-已通过；3-已驳回；4-已取消
	Status int `json:"status"`
	// 0-提交申请时，1-审批通过时，3-提交和审批者抄送
	NotifyAttr int `gorm:"column:notifyAttr" json:"notifyAttr"`
	Createtime int `json:"createtime,omitempty"`
}

// Approvaldata 审批流数据
type Approvaldata struct {
	Errcode int            `json:"errcode"`
	Errmsg  string         `json:"errmsg"`
	Data    *ExecutionData `json:"data"`
}

// Approvaldata1 Approvaldata1
type Approvaldata1 struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Data    string `json:"data"`
}

// ToString ToString
func (a *Approvaldata) ToString() string {
	str, _ := util.ToJSONStr(a)
	return str
}

// SaveOrUpdate SaveOrUpdate
func (w *WeixinFlowApprovaldata) SaveOrUpdate() error {
	if len(w.ThirdNo) == 0 {
		return errors.New("流水号 ThirdNo 不能为空")
	}
	return wxdb.Table(WeixinFlowApprovaldataTable).Where("thirdNo=?", w.ThirdNo).Assign(*w).FirstOrCreate(w).Error
}

// InsertTX InsertTX
func (w *WeixinFlowApprovaldata) InsertTX(tx *gorm.DB) error {
	return tx.Create(w).Error
}

// UpdateDataTX UpdateDataTX
func (w *WeixinFlowApprovaldata) UpdateDataTX(tx *gorm.DB, e *ExecutionData) error {

	ad := &Approvaldata{
		Data:    e,
		Errmsg:  "ok",
		Errcode: 0,
	}
	data, err := util.ToJSONStr(ad)
	print(data)
	if err != nil {
		return err
	}
	w.Data = data
	return tx.Model(&WeixinFlowApprovaldata{}).Save(w).Error
}

// FindWeixinFlowApprovaldata FindWeixinFlowApprovaldata
func FindWeixinFlowApprovaldata(thirdNo string) (*WeixinFlowApprovaldata, error) {
	var data WeixinFlowApprovaldata
	err := wxdb.Where("thirdNo=?", thirdNo).Find(&data).Error
	if err != nil {
		err = fmt.Errorf("查询流程【%s】的直行流:%v", thirdNo, err)
	}
	return &data, err
}

//GetExecutionData 获取执行数据
func (w *WeixinFlowApprovaldata) GetExecutionData() (*ExecutionData, error) {
	ad := &Approvaldata{}
	err := util.Str2Struct(w.Data, ad)
	return ad.Data, err
}
