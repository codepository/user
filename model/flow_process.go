package model

import "github.com/jinzhu/gorm"

// WeixinFlowProcessTable WeixinFlowProcessTable
var WeixinFlowProcessTable = "weixin_flow_process"

// WeixinFlowProcess 流程参数
type WeixinFlowProcess struct {
	ProcessInstanceID string `gorm:"primary_key;column:processInstanceId" json:"processInstanceId,omitempty"`
	UserID            string `gorm:"column:userId" json:"userId,omitempty"`
	RequestedDate     string `gorm:"column:requestedDate" json:"requestedDate,omitempty"`
	Title             string `gorm:"column:title" json:"title,omitempty"`
	BusinessType      string `gorm:"column:businessType" json:"businessType,omitempty"`
	Completed         int    `gorm:"column:completed" json:"completed,omitempty"`
	DeptName          string `gorm:"column:deptName" json:"deptName,omitempty"`
	Candidate         string `json:"candidate,omitempty"`
	Username          string `gorm:"column:username" json:"username,omitempty"`
	DeploymentID      string `gorm:"column:deploymentId" json:"deploymentId,omitempty"`
}

// FindAllFlowProcess 查询流程
func FindAllFlowProcess(query interface{}, values ...interface{}) ([]*WeixinFlowProcess, error) {
	var datas []*WeixinFlowProcess
	err := wxdb.Where(query, values...).Order("requestedDate desc").Find(&datas).Error
	return datas, err
}

// FindAllFlowProcessPaged 查询流程
func FindAllFlowProcessPaged(limit, offset int, query interface{}, values ...interface{}) ([]*WeixinFlowProcess, int, error) {
	if limit == 0 {
		limit = 20
	}
	var total int
	var datas []*WeixinFlowProcess
	err := wxdb.Table(WeixinFlowProcessTable).Where(query, values...).Count(&total).Order("requestedDate desc").Limit(limit).Offset(offset).
		Find(&datas).Error
	return datas, total, err
}

// UpdateTX UpdateTX
func (p *WeixinFlowProcess) UpdateTX(tx *gorm.DB) error {
	return tx.Model(&WeixinFlowProcess{}).Updates(p).Error
}

// InsertTX InsertTX
func (p *WeixinFlowProcess) InsertTX(tx *gorm.DB) error {
	return tx.Model(&WeixinFlowProcess{}).Create(p).Error
}
