package model

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type labelStatus int

// LabelTable LabelTable
var LabelTable = "weixin_oauser_tag"

// DepartmentLeader DepartmentLeader
var DepartmentLeader = "部门领导"

// LeadersInCharge LeadersInCharge
var LeadersInCharge = "分管领导"

// WeixinOauserTag 标签
type WeixinOauserTag struct {
	ID       int    `gorm:"primary_key" json:"id,omitempty"`
	TagName  string `gorm:"column:tagName" json:"tagName"`
	Type     string `json:"type"`
	Describe string `json:"describe,omitempty"`
}

// FromMap FromMap
func (l *WeixinOauserTag) FromMap(json map[string]interface{}) error {
	name, yes := json["tagName"].(string)
	if !yes {
		return errors.New("tagName 须为字符串")
	}
	l.TagName = name
	type1, yes := json["type"].(string)
	if !yes {
		return errors.New("type 须为字符串")
	}
	l.Type = type1
	describe, yes := json["describe"].(string)
	if !yes {
		return errors.New("describe 须为字符串")
	}
	l.Describe = describe
	return nil

}

// SaveOrUpdate 不存在就保存
func (l *WeixinOauserTag) SaveOrUpdate() error {
	//
	if len(l.TagName) == 0 || len(l.Type) == 0 || len(l.Describe) == 0 {
		return errors.New("tagName、type、describe都不能为空;tagName是标签的名称,type是标签的类型,describe是标签的作用")
	}
	return wxdb.Where(WeixinOauserTag{TagName: l.TagName}).Assign(l).FirstOrCreate(l).Error
}

// GetLabels 获取所有标签
func GetLabels() ([]*WeixinOauserTag, error) {
	var result []*WeixinOauserTag
	err := wxdb.Find(&result).Error
	return result, err
}

// FindAllTags 查询所有标签
func FindAllTags(query interface{}, values ...interface{}) ([]*WeixinOauserTag, error) {
	var datas []*WeixinOauserTag
	err := wxdb.Where(query, values...).Find(&datas).Error
	return datas, err
}

// FindAllUserTags 查询某个用户所有的标签
func FindAllUserTags(userID int, query interface{}, values ...interface{}) ([]*WeixinOauserTag, error) {
	var datas []*WeixinOauserTag
	err := wxdb.Table(LabelTable+" t").Select("t.*").
		Joins("join "+UserLabelTable+" tu on uId=? and t.id=tu.tagId", userID).
		Where(query, values...).Find(&datas).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return datas, nil
}
