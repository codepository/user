package model

import (
	"errors"
	"fmt"
	"log"

	"github.com/mumushuiding/util"
)

// UserLabelTable UserLabelTable
var UserLabelTable = "weixin_oauser_taguser"

// WeixinOauserTaguser 用户标签
type WeixinOauserTaguser struct {
	ID        int    `gorm:"primary_key" json:"id,omitempty"`
	UserID    int    `gorm:"column:uId" json:"uId,omitemtpy"`
	TagID     int    `gorm:"column:tagId" json:"tagId,omitemtpy"`
	LabelName string `json:"label_name,omitempty"`
}

// FromMap 使用map进行赋值
func (u *WeixinOauserTaguser) FromMap(fields map[string]interface{}) error {
	if fields["uId"] == nil || fields["tagId"] == nil {
		return errors.New("uId、tagId不能为空")
	}

	userid, ok := fields["uId"].(float64)
	if !ok {
		return errors.New("uId 必须为整数")
	}
	labelid, ok := fields["tagId"].(float64)
	if !ok {
		return errors.New("tagId 必须为整数")
	}
	u.UserID = int(userid)
	u.TagID = int(labelid)
	return nil
}

// ToString ToString
func (u *WeixinOauserTaguser) ToString() string {
	str, _ := util.ToJSONStr(u)
	return str
}

// SaveOrUpdate 保存或更新
func (u *WeixinOauserTaguser) SaveOrUpdate() error {

	return wxdb.Where(WeixinOauserTaguser{UserID: u.UserID, TagID: u.TagID}).Assign(u).Omit("label_name").FirstOrCreate(u).Error
}

// DelSingle 删除一个
func (u *WeixinOauserTaguser) DelSingle() error {
	log.Println("label:", u.ToString())
	if u.ID == 0 {
		if u.UserID == 0 || u.TagID == 0 {
			return errors.New("label的id不能为空或者uId和tagId都不为空")
		}
	}
	return wxdb.Where(u).Delete(&WeixinOauserTaguser{}).Error
}

// FindUserLabels 查询用户标签
func FindUserLabels(query interface{}) ([]*WeixinOauserTaguser, error) {
	var labels []*WeixinOauserTaguser
	err := wxdb.Table(fmt.Sprintf("%s u", UserLabelTable)).Select("u.uId as uId,l.id as tagId,l.tagName as label_name").
		Joins(fmt.Sprintf("join %s l on l.id=u.tagId", LabelTable)).
		Where("uId in (?)", query).Find(&labels).Error
	if err != nil {
		return nil, err
	}
	return labels, nil
}

// FindUsersByLabelIDs 根据用户标签id查询用户
func FindUsersByLabelIDs(ids []interface{}, offset, limit int) ([]*Userinfo, error) {
	var result []*Userinfo
	if limit == 0 {
		limit = 20
	}
	err := wxdb.Table(UserLabelTable).Select("u.*").Joins(fmt.Sprintf("join %s u on u.id=%s.uId", UserinfoTabel, UserLabelTable)).
		Where("tagId in (?)", ids).Group("u.id").
		Offset(offset).
		Limit(limit).
		Find(&result).Error
	return result, err
}
