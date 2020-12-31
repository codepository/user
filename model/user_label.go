package model

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/mumushuiding/util"
)

// UserLabelTable UserLabelTable
var UserLabelTable = "weixin_oauser_taguser"

// WeixinOauserTaguser 用户标签
type WeixinOauserTaguser struct {
	ID      int    `gorm:"primary_key" json:"id,omitempty"`
	UserID  int    `gorm:"column:uId" json:"uId,omitemtpy"`
	TagID   int    `gorm:"column:tagId" json:"tagId,omitemtpy"`
	TagName string `gorm:"column:tagName" json:"tagName"`
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
	err := wxdb.Where(WeixinOauserTaguser{UserID: u.UserID, TagID: u.TagID}).Assign(u).Omit("tagName").FirstOrCreate(u).Error
	if err != nil {
		return fmt.Errorf("保存用户标签失败：%s", err.Error())
	}
	return nil
}

// DelSingle 删除一个
func (u *WeixinOauserTaguser) DelSingle() error {
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
	err := wxdb.Table(fmt.Sprintf("%s u", UserLabelTable)).Select("u.uId as uId,l.id as tagId,l.tagName as tagName").
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

// FindUsersByLabelNames 根据标签名称查找用户
func FindUsersByLabelNames(names []string) ([]*Userinfo, error) {
	var result []*Userinfo
	err := wxdb.Raw("select * from weixin_leave_userinfo where id in (select uId from weixin_oauser_taguser where tagId in (select id from weixin_oauser_tag where tagName in (?)))", names).Scan(&result).Error

	return result, err
}

// IsUserHasLabel 查询用户是否包含某个标签
func IsUserHasLabel(userID int, tagName string) (bool, error) {
	var datas []*WeixinOauserTaguser
	err := wxdb.Table(UserLabelTable+" ut").Select("ut.*").Joins("join "+LabelTable+" t on t.id=ut.tagId and tagName='"+tagName+"'").Where("uId=?", userID).
		Find(&datas).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("用户是否存在标签【%s】:%s", tagName, err.Error())
	}
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if len(datas) == 0 {
		return false, nil
	}
	return true, nil
}

// AddUserLabelByLabelName 根据标签名为用户添加标签
func AddUserLabelByLabelName(userID int, tagName string) error {
	var tu WeixinOauserTaguser
	tags, err := FindAllTags("", fmt.Sprintf("tagName='%s'", tagName))
	if err != nil {
		return fmt.Errorf("查询标签【%s】失败:%s", tagName, err.Error())
	}
	if len(tags) == 0 {
		return fmt.Errorf("不存在标签【%s】", tagName)
	}
	tu.TagID = tags[0].ID
	tu.TagName = tags[0].TagName
	tu.UserID = userID
	return tu.SaveOrUpdate()
}

// CountUserByTagID 查询标签对应用户人数
func CountUserByTagID(id int) (int, error) {
	var count int
	err := wxdb.Table(UserLabelTable).Where("tagId=?", id).Count(&count).Error
	return count, err

}
