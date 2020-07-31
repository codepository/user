package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codepository/user/model"
	"github.com/mumushuiding/util"
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
	c.Body.Fields = []string{"用户信息", "用户标签", "分管部门", "可访问的受限url路径"}
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
	// 可访问的受限url路径
	var role []string
	for _, label := range labels {
		role = append(role, label.TagName)
	}
	var list []string
	permissionlist, err := model.FindAllPermissionByRoles(role)
	if err != nil {
		return nil, err
	}

	for _, p := range permissionlist {
		list = append(list, p.URL)
	}
	if list == nil {
		list = make([]string, 0)
	}
	data = append(data, list)
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

// GetUseridsByTagAndLevel GetUseridsByTagAndLevel
func GetUseridsByTagAndLevel(c *model.Container) error {
	errstr := `参数格式:{"body":{"data":[{"tags":["第一考核组成员","项目舞台"],"tag_method":"and","level":1}]}}, and表示同时拥有"第一考核组成员","项目舞台"标签的用户`
	var err error
	// 参数解析
	if c.Body.Data == nil || len(c.Body.Data) == 0 {
		return errors.New(errstr)
	}
	par, ok := c.Body.Data[0].(map[string]interface{})
	if len(par) == 0 {
		return errors.New(errstr)
	}
	if !ok {
		return errors.New(errstr)
	}
	// 标签判断
	tags, ok := par["tags"].([]interface{})
	if !ok {
		return errors.New(errstr)
	}
	// 根据标签名称获取标签集合
	tagids, err := FindTagidsByTagName(tags)
	if err != nil {
		return err
	}
	method, ok := par["tag_method"].(string)
	if !ok {
		return errors.New(errstr)
	}
	fields := "id"
	var users []*model.Userinfo
	if par["level"] != nil {
		level, err := util.Interface2Int(par["level"])
		if err != nil {
			return errors.New(errstr)
		}
		users, err = FindUsersByTagidsAndLevel(tagids, level, method, fields)
		if err != nil {
			return nil
		}
	} else {
		users, err = FindUsersByTagids(tagids, method, fields)
		if err != nil {
			return nil
		}
	}
	var userids []int
	for _, user := range users {
		userids = append(userids, user.ID)
	}
	c.Body.Data = c.Body.Data[:0]
	c.Body.Data = append(c.Body.Data, userids)
	return nil
}
func generateSQL(tagids []int, method, fields string) *strings.Builder {
	size := len(tagids)
	if len(fields) == 0 {
		fields = "*"
	}
	var sqlbuff strings.Builder
	sqlbuff.WriteString("select " + fields + " from " + model.UserinfoTabel + " where id in (select uId from " + model.UserLabelTable + " where ")
	if method == "and" {
		var tagidsbuff strings.Builder
		for _, tag := range tagids {
			tagidsbuff.WriteString(fmt.Sprintf(" or tagId=%d", tag))
		}
		sqlbuff.WriteString(fmt.Sprintf("(%s) group by uId having count(uId)=%d)", tagidsbuff.String()[3:], size))
	} else {
		var tagidsbuff strings.Builder
		for _, tag := range tagids {
			tagidsbuff.WriteString(fmt.Sprintf(",%d", tag))
		}
		sqlbuff.WriteString("tagId in (" + tagidsbuff.String()[1:] + "))")
	}
	return &sqlbuff
}

// FindUsersByTagids 根据用户标查询
// fields 表示查询的字段
// method 有两个值 and 表示与查询，or表示或查询
// 多标签查询 select id,name from weixin_leave_userinfo where  id in (select uId from weixin_oauser_taguser where (tagId=29 or tagId=36) group by uId having count(uId)=2);
func FindUsersByTagids(tagids []int, method, fields string) ([]*model.Userinfo, error) {
	if len(tagids) == 0 {
		return nil, errors.New("tagids 不能为空")
	}
	sqlbuff := generateSQL(tagids, method, fields)
	return model.FindAllUserinfoByRawSQL(sqlbuff.String())
}

// FindUsersByTagidsAndLevel FindUsersByTagidsAndLevel
func FindUsersByTagidsAndLevel(tagids []int, level int, method, fields string) ([]*model.Userinfo, error) {
	if len(tagids) == 0 {
		return nil, errors.New("tagids 不能为空")
	}
	sqlbuff := generateSQL(tagids, method, fields)
	sqlbuff.WriteString(fmt.Sprintf(" and level=%d", level))
	return model.FindAllUserinfoByRawSQL(sqlbuff.String())
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
