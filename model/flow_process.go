package model

import "github.com/jinzhu/gorm"

// FznewsFlowProcessTable FznewsFlowProcessTable
var FznewsFlowProcessTable = "fznews_flow_process"

// FznewsFlowProcess 流程参数
type FznewsFlowProcess struct {
	ProcessInstanceID string `gorm:"primary_key;column:processInstanceId" json:"processInstanceId,omitempty"`
	// UID 对应用户ID
	UID int `gorm:"column:uid" json:"uid,omitempty"`
	// UserID 对应微信ID
	UserID        string `gorm:"column:userId" json:"userId,omitempty"`
	RequestedDate string `gorm:"column:requestedDate" json:"requestedDate,omitempty"`
	Title         string `gorm:"column:title" json:"title,omitempty"`
	BusinessType  string `gorm:"column:businessType" json:"businessType,omitempty"`
	Completed     int    `gorm:"column:completed" json:"completed"`
	DeptName      string `gorm:"column:deptName" json:"deptName,omitempty"`
	Candidate     string `json:"candidate,omitempty"`
	Username      string `gorm:"column:username" json:"username,omitempty"`
	DeploymentID  string `gorm:"column:deploymentId" json:"deploymentId,omitempty"`
	// step 当前执行步骤
	Step int `json:"step"`
}

// FindAllFlowProcess 查询流程
func FindAllFlowProcess(query interface{}, values ...interface{}) ([]*FznewsFlowProcess, error) {
	var datas []*FznewsFlowProcess
	err := wxdb.Where(query, values...).Order("requestedDate desc").Find(&datas).Error
	return datas, err
}

// FindAllFlowProcessPaged 查询流程
func FindAllFlowProcessPaged(limit, offset int, query interface{}, values ...interface{}) ([]*FznewsFlowProcess, int, error) {
	if limit == 0 {
		limit = 20
	}
	var total int
	var datas []*FznewsFlowProcess
	err := wxdb.Table(FznewsFlowProcessTable).Where(query, values...).Count(&total).Order("requestedDate desc").Limit(limit).Offset(offset).
		Find(&datas).Error
	return datas, total, err
}

// DeleteFlowByID 删除流程
func DeleteFlowByID(id string) error {
	err := wxdb.Table(FznewsFlowProcessTable).Where("processInstanceId=?", id).Delete(FznewsFlowProcess{}).Error
	return err
}

// UpdateTX UpdateTX
func (p *FznewsFlowProcess) UpdateTX(tx *gorm.DB) error {
	return tx.Model(&FznewsFlowProcess{}).Save(p).Error
}

// InsertTX InsertTX
func (p *FznewsFlowProcess) InsertTX(tx *gorm.DB) error {
	return tx.Model(&FznewsFlowProcess{}).Create(p).Error
}
