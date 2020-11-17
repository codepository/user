package model

import (
	"errors"
)

// FznewsLeadershipTable 分管部门表格名
var FznewsLeadershipTable = "fznews_leadership"

// FznewsLeadership 设置用户分管的部门
type FznewsLeadership struct {
	Model
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	// 对应微信ID
	Userid         string `json:"userid"`
	DepartmentID   int    `json:"department_id"`
	DepartmentName string `json:"department_name"`
	// 对应该部门的角色
	Role int `json:"role"`
}

// FznewsLeadershipUser FznewsLeadershipUser
type FznewsLeadershipUser struct {
	FznewsLeadership
	Leader string `json:"leader,omitempty"`
	Avatar string `json:"avatar,omitemtpy"`
}

// SaveOrUpdate 不存在保存，存在就更新
func (f *FznewsLeadership) SaveOrUpdate() error {
	if f.UserID == 0 || f.DepartmentID == 0 || len(f.DepartmentName) == 0 || f.Role == 0 {
		return errors.New("user_id、department_id、department_name、role 不能为空")
	}
	return db.Where(FznewsLeadership{UserID: f.UserID, DepartmentID: f.DepartmentID}).Assign(*f).FirstOrCreate(f).Error
}

// FirstOrCreate 存在就更新，否则就创建
func (f *FznewsLeadership) FirstOrCreate() error {
	return db.Where(FznewsLeadership{Userid: f.Userid, Role: f.Role, DepartmentID: f.DepartmentID}).Assign(f).FirstOrCreate(f).Error
}

// DelFznewsLeadership 删除
func DelFznewsLeadership(query interface{}, values ...interface{}) error {
	if query == nil {
		return errors.New("参数不能为空")
	}
	return db.Where(query, values...).Delete(&FznewsLeadership{}).Error
}

// FindFznewsLeadership FindFznewsLeadership
func FindFznewsLeadership(query interface{}, values ...interface{}) ([]*FznewsLeadershipUser, error) {
	var result []*FznewsLeadershipUser
	// err := db.Table(FznewsLeadershipTable + " l").Select("l.*,u.name as leader,u.avatar").
	// 	Joins("join " + UserinfoTabel + " u on u.id=l.user_id").
	// 	Where(query).Find(&result).Error
	err := db.Table(FznewsLeadershipTable).Select("*").
		Where(query, values...).Find(&result).Error
	if len(result) > 0 {
		var userid []int
		for _, u := range result {
			userid = append(userid, u.UserID)
		}
		// 查询用户信息
		users, err := FindAllUserInfo("*", userid)
		if err != nil {
			return nil, err
		}
		umap := make(map[int]*Userinfo)
		for _, u := range users {
			umap[u.ID] = u
		}
		for _, r := range result {
			r.Avatar = umap[r.UserID].Avatar
			r.Leader = umap[r.UserID].Name
			// // 查询用户领导组标签
			// tags, err := FindAllUserTags(r.UserID, "type=?", "领导组")
			// if err != nil {
			// 	return nil, fmt.Errorf("查询用户领导组标签失败:%s", err.Error())
			// }
			// if len(tags) > 0 {
			// 	var s string
			// 	for _, t := range tags {
			// 		s = s + "," + t.TagName
			// 	}
			// 	r.LeaderType = s[1:]
			// }

		}
	}
	return result, err
}

// FindLeadershipByDepartmentIDAndRole 根据部门、角色及部门层级level查询上级领导
// 若当前部门的层级小于level,那么上级领导要上溯至层级为level的部门领导
// 如level=3，领导为甲；当前部门层级为5，领导为乙，则上级领导为甲
// 如level=3,领导为丙;当前部门层级为2，领导为乙，则上级领导领导为乙
func FindLeadershipByDepartmentIDAndRole(departmentID int, role, level int) ([]*FznewsLeadershipUser, error) {
	var did int
	var err error
	cur, err := FindDepartmentByID(departmentID)
	if err != nil {
		return nil, err
	}
	if cur.Level == 0 {
		l, err := SetDepartmentLevelByID(cur.ID)
		if err != nil {
			return nil, err
		}
		cur.Level = l
	}
	// 部门层级<=level,当前部门的领导为直接上级
	if cur.Level <= level {
		did = cur.ID
	} else {
		// 若为子部门上级领导角色为部门领导
		role = 5
		var parent = cur
		var curlevel = cur.Level
		// 部门上溯找到层级为level的上级部门
		for {
			// 结束条件
			if parent.Parentid == 0 || parent.Level <= level || curlevel <= level {
				did = parent.ID
				break
			}
			parent, err = FindDepartmentByID(parent.Parentid)
			if err != nil {
				return nil, err
			}
			curlevel--

		}
	}
	return FindFznewsLeadership("department_id=? and role=?", did, role)
}
