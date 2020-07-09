package model

import (
	"errors"

	"github.com/jinzhu/gorm"
)

// UserOrg 用户可以查询的行业信息，对应user_org表
type UserOrg struct {
	ID       int    `gorm:"primary_key" json:"id,omitempty"`
	Orgid    int    `json:"orgid,omitempty"`
	Org      string `json:"org,omitempty"`
	Userid   int    `json:"userid,omitempty"`
	Username string `json:"username,omitempty"`
}

// FromMap FromMap
func (u *UserOrg) FromMap(fields map[string]interface{}) error {
	if fields["orgid"] == nil {
		return errors.New("orgid 不能为空")
	}
	if fields["org"] == nil {
		return errors.New("org 不能为空")
	}
	if fields["userid"] == nil {
		return errors.New("userid 不能为空")
	}
	if fields["username"] == nil {
		return errors.New("username 不能为空")
	}
	orgid, ok := fields["orgid"].(float64)
	if !ok {
		return errors.New("orgid 必须为数字")
	}
	userid, ok := fields["userid"].(float64)
	if !ok {
		return errors.New("userid 必须为数字")
	}
	org, ok := fields["org"].(string)
	if !ok {
		return errors.New("org 必须为字符串")
	}
	username, ok := fields["username"].(string)
	if !ok {
		return errors.New("username 必须为字符串")
	}
	u.Orgid = int(orgid)
	u.Userid = int(userid)
	u.Org = org
	u.Username = username
	return nil
}

// FindAllOrg 查询
func FindAllOrg(query interface{}, fields string) ([]*UserOrg, error) {
	var result []*UserOrg
	if query == nil {
		return nil, errors.New("参数不能为空")
	}
	if len(fields) == 0 {
		fields = "*"
	}
	err := db.Select(fields).Where(query).Find(&result).Error
	if err == gorm.ErrRecordNotFound {
		return make([]*UserOrg, 0), nil
	}
	return result, err
}

// DelUserOrg 删除
func DelUserOrg(query interface{}) error {
	if query == nil {
		return errors.New("参数不能为空")
	}
	return db.Where(query).Delete(&UserOrg{}).Error
}

// Save 保存
func (u *UserOrg) Save() error {
	return db.Create(u).Error
}
