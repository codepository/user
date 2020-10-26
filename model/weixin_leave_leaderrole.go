package model

import "github.com/jinzhu/gorm"

// WeixinLeaveLeaderrole 分管领导信息表
type WeixinLeaveLeaderrole struct {
	ID int `json:"id"`
	// 角色
	Role int `json:"role"`
	// 用户微信帐号名
	Userid string `json:"userid"`
	// 用户姓名
	Username string `json:"username"`
	// 部门
	Dept string `json:"dept"`
}

// FindAllLeaveLeaderrole FindAllLeaveLeaderrole
func FindAllLeaveLeaderrole(query interface{}, values ...interface{}) ([]*WeixinLeaveLeaderrole, error) {
	var datas []*WeixinLeaveLeaderrole
	err := wxdb.Where(query, values...).Find(&datas).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return datas, nil
}
