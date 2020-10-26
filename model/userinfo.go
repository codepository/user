package model

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/mumushuiding/util"
)

// Userinfo 同weixin_leave_userinfo对应
type Userinfo struct {
	ID int `json:"id,omitempty"`
	// 对应微信id
	Userid         string `json:"userid"`
	Name           string `json:"name,omitempty"`
	DepartmentID   int    `gorm:"column:departmentid" json:"departmentid,omitempty"`
	Departmentname string `json:"departmentname,omitempty"`
	Position       string `json:"position,omitempty"`
	Mobile         string `json:"mobile,omitempty"`
	Gender         int    `json:"gender,omitempty"`
	Email          string `json:"email,omitempty"`
	Avatar         string `json:"avatar,omitempty"`
	Status         int    `json:"status,omitempty"`
	//0是一般工作人员， 1中层正职（含主持工作的副职）, 2是中层副职，3是社领导
	Level int `json:"level"`
	// 1表示是部门领导，0表示非部门领导
	IsLeader int `json:"is_leader"`
}

// UserinfoTabel 对应weixin_leave_userinfo表
var UserinfoTabel = "weixin_leave_userinfo"

// ToString ToString
func (u *Userinfo) ToString() string {
	str, _ := util.ToJSONStr(u)
	return str
}

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

// FindAllUserinfoByRawSQL 根据rawsql查询
func FindAllUserinfoByRawSQL(rawSQL string, values ...interface{}) ([]*Userinfo, error) {
	var datas []*Userinfo
	err := wxdb.Raw(rawSQL, values...).Scan(&datas).Error
	return datas, err
}

// FindAllUserInfo 查询所有用户
func FindAllUserInfo(query interface{}, values ...interface{}) ([]*Userinfo, error) {
	var users []*Userinfo
	err := wxdb.Table(UserinfoTabel).Where(query, values...).Find(&users).Error
	return users, err
}

// FindAllUseridsByDepartmentids 根据部门id查询所有用户id
func FindAllUseridsByDepartmentids(departmentids []interface{}) ([]int, error) {
	var ids []int
	err := wxdb.Table(UserinfoTabel).Model(&Userinfo{}).Select("id").Where("departmentid in (?)", departmentids).Pluck("id", &ids).Error
	return ids, err
}

// FindSecondLeaderByDepartmentid 上级的上级的领导
func FindSecondLeaderByDepartmentid(departmentid int) ([]*Userinfo, error) {
	var result []*Userinfo
	err := wxdb.Raw("select * from weixin_leave_userinfo where is_leader=1 and departmentid in (select parentid from weixin_leave_department where id=?)", departmentid).Scan(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("该部门不存在二级部门领导,请联系管理员")
	}
	return result, nil
}

// FindLeaderByDepartmentID 若isLeader是0,用户上级为部门领导;若isLeader是1上级为分管部门的领导
// 根据用户部门找出分管该部门的users
// 若isLeader=0,则从users筛选出包含"部门领导"标签的用户，否则筛选出包含"分管领导"标签的用户
// func FindLeaderByDepartmentID(departmentID int, isLeader float64) ([]*Userinfo, error) {
// 	var result []*Userinfo
// 	var err error
// 	var yes bool
// 	// 查询管理该部门的用户
// 	err = wxdb.Raw("select * from weixin_leave_userinfo u join fznews_leadership l on l.user_id=u.id and l.department_id=?", departmentID).Scan(&result).Error
// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		return nil, err
// 	}
// 	if len(result) == 0 {
// 		return nil, fmt.Errorf("部门[%d]未设置分管领导,联系管理员", departmentID)
// 	}
// 	// 查询用户是否拥有"部门领导"或者"分管领导"标签
// 	var users []*Userinfo
// 	var ustr []string
// 	for _, u := range result {
// 		ustr = append(ustr, u.Name)
// 		if isLeader == 0 {
// 			// 找部门领导，取
// 			yes, err = IsUserHasLabel(u.ID, DepartmentLeader)
// 			// log.Printf("用户:%s,包含标签:%s\n,是真的:%v", u.Name, DepartmentLeader, yes)
// 		} else {
// 			yes, err = IsUserHasLabel(u.ID, LeadersInCharge)
// 			// log.Printf("用户:%s,包含标签:%s\n,是真的:%v", u.Name, LeadersInCharge, yes)
// 		}
// 		if err != nil {
// 			return nil, err
// 		}
// 		if yes {
// 			users = append(users, u)
// 		}
// 	}
// 	if len(users) == 0 {
// 		if isLeader == 0 {
// 			return nil, fmt.Errorf("用户%v若是部门领导则需要设置【部门领导】标签", ustr)
// 		}
// 		return nil, fmt.Errorf("用户%v若是分管领导则需要设置【分管领导】标签", ustr)
// 	}
// 	return users, nil
// }

// FindLeaderByUserID 若isLeader是0,用户上级为部门领导;若isLeader是1上级为分管部门的领导
// 根据用户部门找出分管该部门的users
// 若isLeader=0,则从users筛选出包含"部门领导"标签的用户，否则筛选出包含"分管领导"标签的用户
func FindLeaderByUserID(userID string, isLeader float64) ([]*Userinfo, error) {
	var result []*Userinfo
	if isLeader == 1 {
		//
		err := wxdb.Raw("select * from weixin_leave_userinfo u join fznews_leadership l on l.user_id=u.id and l.department_id in (select departmentid from weixin_leave_userinfo where userid=?)", userID).Scan(&result).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("该部门未设置分管领导,请联系管理员")
		}
	} else {
		err := wxdb.Raw("select * from weixin_leave_userinfo where is_leader=1 and departmentid in (select departmentid from weixin_leave_userinfo where userid=?)", userID).Scan(&result).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("该部门未设置部门领导,请联系管理员")
		}
	}
	return result, nil
}

// FindUserinfoByUserID 根据userID查询用户信息
func FindUserinfoByUserID(userID string) (*Userinfo, error) {
	var result *Userinfo
	err := wxdb.Table(UserinfoTabel).Where("userid=?", userID).Find(result).Error
	return result, err
}
