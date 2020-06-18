package service

import (
	"errors"

	"github.com/codepository/user/model"
)

// UserLoginAndFindUserinfo 用户登陆
func UserLoginAndFindUserinfo(account, password string) (*model.Container, error) {
	// 账号匹配
	user, token, err := model.UserLogin(account, password)
	if err != nil {
		return nil, err
	}
	var c model.Container
	c.Header.Token = token
	// 根据用户查询所有的用户信息
	data, err := GetAllUserinfo(user)
	if err != nil {
		return nil, err
	}
	c.Body.Data = data
	return &c, nil
}

// GetAllUserinfo 根据用户查询所有的用户信息
func GetAllUserinfo(user *model.Userinfo) ([]interface{}, error) {
	var data []interface{}
	data = append(data, user)
	// 用户标签查询
	labels, err := FindUserLabel([]int{user.ID})
	if err != nil {
		return nil, err
	}
	data = append(data, labels)
	// 用户分管部门查询
	leaership, err := model.FindFznewsLeadership(map[string]interface{}{"user_id": user.ID})
	data = append(data, leaership)
	return data, nil
}

// FindUserLabel FindUserLabel
func FindUserLabel(query interface{}) ([]*model.WeixinOauserTaguser, error) {
	return model.FindUserLabels(query)
}

// FindUserLabelNames 用户所有的标签
func FindUserLabelNames(query interface{}) ([]string, error) {
	labels, err := model.FindUserLabels(query)
	if err != nil {
		return nil, err
	}
	if len(labels) == 0 {
		return []string{}, nil
	}
	var n []string
	for _, l := range labels {
		n = append(n, l.TagName)
	}
	return n, nil
}

// CheckIfHaveNewUser 检查是否有新用户
// 返回用户的id数组
func CheckIfHaveNewUser() ([]int, error) {
	// 查询用户表最新的id
	id1, err := model.FindLastUserinfoID()
	if err != nil {
		return nil, err
	}
	// 查询登陆表最新的id
	id2, err := model.FindLastLoginID()
	if err != nil {
		return nil, err
	}
	// 若id1>id2说明有新用户
	if id1 == id2 {
		return nil, nil
	}
	// 查询最新的用户ids

	return model.FindNewUserinfoIDs(id2)
}

// SetPassWord 设置密码
func SetPassWord(userid int, password string) error {
	login := model.FznewsLogin{
		UserID:   userid,
		Password: password,
	}
	return login.SaveIfNotExist()
}

// UpdatePassword 更新密码
func UpdatePassword(userid int, password string) error {
	login := model.FznewsLogin{
		UserID:   userid,
		Password: password,
	}
	return login.Update()
}

// FindAllUserInfo 查询用户信息
func FindAllUserInfo(query interface{}) ([]*model.Userinfo, error) {
	return model.FindAllUserInfo(query)
}

// GetUsers 分页查询用户信息
func GetUsers(c *model.Container) error {
	return c.FindUserinfoPaged()
}

// AddLabel 添加标签
func AddLabel(userID int, tagID int, tagName string) error {
	label := model.WeixinOauserTaguser{
		UserID:  userID,
		TagID:   tagID,
		TagName: tagName,
	}
	return label.SaveOrUpdate()
}

// FindAllLabel 查询所有标签
func FindAllLabel() ([]*model.WeixinOauserTag, error) {
	labels, err := model.GetLabels()
	if err != nil {
		return nil, err
	}
	return labels, nil
}

// AddlabelbyDepartment 根据部门添加标签
func AddlabelbyDepartment(c *model.Container) error {
	// 验证参数
	if c.Body.Data == nil || len(c.Body.Data) != 2 {
		return errors.New("参数类型:{'body':{'data':[[1,2],[5,6]]}},1,2为标签id,5,6为部门id")
	}
	labels, ok := c.Body.Data[0].([]interface{})
	if !ok {
		return errors.New("参数类型:{'body':{'data':[[1,2],[5,6]]}},1,2为标签id,5,6为部门id")
	}
	if len(labels) == 0 {
		return errors.New("参数类型:{'body':{'data':[[1,2],[5,6]]}},1,2为标签id,5,6为部门id")
	}
	departments, ok := c.Body.Data[1].([]interface{})
	if !ok {
		return errors.New("参数类型:{'body':{'data':[[1,2],[5,6]]}},1,2为标签id,5,6为部门id")
	}
	if len(departments) == 0 {
		return errors.New("参数类型:{'body':{'data':[[1,2],[5,6]]}},1,2为标签id,5,6为部门id")
	}
	// 根据部门id从远程查询用户userids
	userids, err := model.FindAllUseridsByDepartmentids(departments)
	if err != nil {
		return err
	}
	if len(userids) == 0 {
		return errors.New("没有查询到部门对应的员工，添加失败")
	}
	// 插入数据库
	for _, userid := range userids {
		for _, labelid := range labels {
			ul := model.WeixinOauserTaguser{
				UserID: userid,
				TagID:  int(labelid.(float64)),
			}
			ul.SaveOrUpdate()
		}
	}
	return nil

}

// FindUsersByLabelIDs 根据标签查询用户
func FindUsersByLabelIDs(c *model.Container) error {
	if c.Body.Data == nil || len(c.Body.Data) == 0 {
		return errors.New("参数类型:{'body':{'data':[1,2]}},1,2为标签id")
	}
	_, ok := c.Body.Data[0].(float64)
	if !ok {
		return errors.New("参数类型:{'body':{'data':[1,2]}},1,2为标签id")
	}
	result, err := model.FindUsersByLabelIDs(c.Body.Data, c.Body.StartIndex, c.Body.MaxResults)
	if err != nil {
		return err
	}
	c.Body.Data = append(c.Body.Data, result)
	c.Body.Data = c.Body.Data[len(c.Body.Data)-1:]
	return nil
}

// FindAllDepartment 查询所有部门
func FindAllDepartment() ([]*model.TreeNode, error) {
	d, err := model.FindAllWxDepartment()
	if err != nil {
		return nil, err
	}
	result := model.TransformWxDepartment2Tree(d)
	return result, nil
}
