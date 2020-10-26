package model

import "github.com/jinzhu/gorm"

// WeixinFlowLogTable WeixinFlowLogTable
var WeixinFlowLogTable = "weixin_flow_log"

// WeixinFlowLog 审批纪录
type WeixinFlowLog struct {
	ID       int    `gorm:"primary_key" json:"id,omitempty"`
	ThirdNo  string `gorm:"column:thirdNo" json:"thirdNo,omitemtpty"`
	UserID   string `gorm:"column:userId" json:"userId,omitempty"`
	Username string `json:"username,omitempty"`
	Status   int    `json:"status"`
	Speech   string `json:"speech,omitempty"`
	OpTime   int    `gorm:"column:opTime" json:"opTime,omitempty"`
}

// WeixinFlowLogResults WeixinFlowLogResults
type WeixinFlowLogResults struct {
	WeixinFlowLog
	Processname string `json:"processname,omitempty"`
}

// InsertTX InsertTX
func (l *WeixinFlowLog) InsertTX(tx *gorm.DB) error {
	return tx.Create(l).Error
}

// FindAllFlowLogPaged 分页查询审批纪录
func FindAllFlowLogPaged(limit, offset int, query interface{}, values ...interface{}) ([]*WeixinFlowLogResults, int, error) {
	var total int
	var datas []*WeixinFlowLogResults
	if limit == 0 {
		limit = 20
	}
	err := wxdb.Table(WeixinFlowLogTable).Select(WeixinFlowLogTable+".*,"+FznewsFlowProcessTable+".title as processname").Joins("join "+FznewsFlowProcessTable+" on processInstanceId=thirdNo").
		Where(query, values...).Count(&total).Order("opTime desc").Limit(limit).Offset(offset).
		Find(&datas).Error
	if err != nil {
		return nil, 0, nil
	}
	return datas, total, nil
}
