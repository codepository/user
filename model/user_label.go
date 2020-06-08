package model

import (
	"errors"
	"fmt"
	"log"

	"github.com/mumushuiding/util"
)

// UserLabelTable UserLabelTable
var UserLabelTable = "fznews_user_label"

// UserLabel 用户标签
type UserLabel struct {
	Model
	UserID    int    `json:"user_id,omitemtpy"`
	LabelID   int    `json:"label_id,omitemtpy"`
	LabelName string `json:"label_name,omitemtpy"`
}

// FromMap 使用map进行赋值
func (u *UserLabel) FromMap(fields map[string]interface{}) error {
	if fields["user_id"] == nil || fields["label_id"] == nil {
		return errors.New("user_id、label_id、label_name不能为空")
	}

	userid, ok := fields["user_id"].(float64)
	if !ok {
		return errors.New("user_id 必须为整数")
	}
	labelid, ok := fields["label_id"].(float64)
	if !ok {
		return errors.New("label_id 必须为整数")
	}
	u.UserID = int(userid)
	u.LabelID = int(labelid)
	return nil
}

// ToString ToString
func (u *UserLabel) ToString() string {
	str, _ := util.ToJSONStr(u)
	return str
}

// SaveOrUpdate 保存或更新
func (u *UserLabel) SaveOrUpdate() error {

	return db.Where(UserLabel{UserID: u.UserID, LabelID: u.LabelID}).Assign(u).FirstOrCreate(u).Error
}

// DelSingle 删除一个
func (u *UserLabel) DelSingle() error {
	log.Println("label:", u.ToString())
	if u.ID == 0 {
		if u.UserID == 0 || u.LabelID == 0 {
			return errors.New("label的id不能为空或者user_id和label_id都不为空")
		}
	}
	return db.Where(u).Delete(&UserLabel{}).Error
}

// FindUserLabels 查询用户标签
func FindUserLabels(query interface{}) ([]*UserLabel, error) {
	var labels []*UserLabel
	err := db.Table(fmt.Sprintf("%s u", UserLabelTable)).Select("u.user_id,l.id as label_id,l.name as label_name").
		Joins("join fznews_label l on l.id=u.label_id").
		Where("user_id in (?)", query).Find(&labels).Error
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
	err := db.Table(UserLabelTable).Select("u.*").Joins(fmt.Sprintf("join %s u on u.id=%s.user_id", UserinfoTabel, UserLabelTable)).
		Where("label_id in (?)", ids).Group("u.id").
		Offset(offset).
		Limit(limit).
		Find(&result).Error
	return result, err
}
