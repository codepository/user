package model

import (
	"errors"

	"github.com/mumushuiding/util"
)

// Userinfo 同weixin_leave_userinfo对应
type Userinfo struct {
	ID             int    `json:"id,omitempty"`
	UserID         string `json:"userid,omitempty"`
	Name           string `json:"name,omitempty"`
	DepartmentID   int    `json:"departmentid,omitempty"`
	Departmentname string `json:"departmentname,omitempty"`
	Position       string `json:"position,omitempty"`
	Mobile         string `json:"mobile,omitempty"`
	Gender         int    `json:"gender,omitempty"`
	Email          string `json:"email,omitempty"`
	Avatar         string `json:"avatar,omitempty"`
	Status         int    `json:"status,omitempty"`
	//0是一般工作人员， 1中层正职（含主持工作的副职）, 2是中层副职，3是社领导
	Level int `json:"level,omitempty"`
	// 1表示是部门领导，0表示非部门领导
	IsLeader int `json:"is_leader,omitempty"`
}

// UserinfoTabel 对应weixin_leave_userinfo表
var UserinfoTabel = "weixin_leave_userinfo"

// UserLogin 用户登陆
func UserLogin(account, password string) (*Userinfo, string, error) {
	var result []*Userinfo
	var accountType string
	// 查询用户
	yes := util.IsMobile(account)
	if yes {
		accountType = "手机"
		err := wxdb.Table(UserinfoTabel).Where("mobile=?", account).Find(&result).Error
		if err != nil {
			return nil, "", err
		}
	} else if util.IsChinese(account) {
		accountType = "姓名"
		err := wxdb.Table(UserinfoTabel).Where("name=?", account).Find(&result).Error
		if err != nil {
			return nil, "", err
		}
	} else if util.IsEmail(account) {
		accountType = "邮箱"
		err := wxdb.Table(UserinfoTabel).Where("email=?", account).Find(&result).Error
		if err != nil {
			return nil, "", err
		}
	} else {
		return nil, "", errors.New("登陆账号只能为【手机,姓名,邮箱】其中之一")
	}
	if len(result) == 0 {
		return nil, "", errors.New("账号不存在")
	}
	if len(result) > 1 {
		return nil, "", errors.New("有多个" + accountType + "为【" + account + "】的账号，无法登陆！")
	}
	// 匹配密码
	login := FznewsLogin{
		UserID:   result[0].ID,
		Password: password,
	}
	yes, token, err := login.Check()
	if err != nil {
		return nil, "", err
	}
	if !yes {
		return nil, "", errors.New("密码不正确")
	}
	return result[0], token, nil
}

// FindLastUserinfoID 查询最新的用户id
func FindLastUserinfoID() (int, error) {
	var id []int
	err := wxdb.Table(UserinfoTabel).Model(&Userinfo{}).Select("id").Order("id desc").Limit(1).Pluck("id", &id).Error
	if err != nil {
		return 0, err
	}
	return id[0], nil
}

// FindNewUserinfoIDs 查询所有新的用户id
func FindNewUserinfoIDs(last int) ([]int, error) {
	var ids []int
	err := wxdb.Table(UserinfoTabel).Model(&Userinfo{}).Select("id").Where("id>?", last).Pluck("id", &ids).Error
	return ids, err
}

// FindAllUserInfo 查询所有用户
func FindAllUserInfo(query interface{}) ([]*Userinfo, error) {
	var users []*Userinfo
	err := wxdb.Table(UserinfoTabel).Where(query).Find(&users).Error
	return users, err
}

// FindAllUseridsByDepartmentids 根据部门id查询所有用户id
func FindAllUseridsByDepartmentids(departmentids []interface{}) ([]int, error) {
	var ids []int
	err := wxdb.Table(UserinfoTabel).Model(&Userinfo{}).Select("id").Where("departmentid in (?)", departmentids).Pluck("id", &ids).Error
	return ids, err
}
