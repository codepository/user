package conmgr

import (
	"errors"
	"strings"

	"github.com/codepository/user/config"
	"github.com/codepository/user/model"
	"github.com/codepository/user/service"
	"github.com/jinzhu/gorm"
	"github.com/mumushuiding/util"
)

type userinfoCacheStatus int

const (
	userCacheStatus = iota
	labelsCacheStatus
	leadershipCacheStatus
	permissionCacheStatus
)

// AlterPass 修改密码
func AlterPass(c *model.Container) error {
	if len(c.Header.Token) == 0 {
		return errors.New("token不能为空{header:{token:'fdsfs'}} fdsfs为token")
	}
	if len(c.Body.Metrics) == 0 {
		return errors.New("密码不能为空{body:{metrics:'fdsfs'}} fdsfs为新密码")
	}
	user, err := GetUserByToken(c.Header.Token)
	if err != nil {
		return err
	}
	// 更新密码
	err = service.UpdatePassword(user.ID, c.Body.Metrics)
	if err != nil {
		return err
	}
	// 更新用户缓存信息
	l := &model.FznewsLogin{UserID: user.ID, Password: c.Body.Metrics}
	token, _ := l.GetToken()
	Conmgr.cacheMap[token] = Conmgr.cacheMap[c.Header.Token]
	c.Header.Token = token
	c.Body.Data = nil
	return nil
}

// GetPermissionByToken 获取用户可受限的访问路径
func GetPermissionByToken(token string) ([]string, error) {
	userinfos := Conmgr.userMap[token]
	if userinfos == nil {
		return nil, errors.New("请重新登陆")
	}
	permission := userinfos.([]interface{})[permissionCacheStatus].([]string)
	return permission, nil

}

// GetUserByToken 根据token获取用户信息
func GetUserByToken(token string) (*model.Userinfo, error) {
	userinfos := Conmgr.userMap[token]
	if userinfos == nil {
		return nil, errors.New("请重新登陆")
	}
	user := userinfos.([]interface{})[userCacheStatus].(*model.Userinfo)
	return user, nil
}

// GetLabelsByToken 根据token获取用户标签
func GetLabelsByToken(token string) ([]*model.WeixinOauserTaguser, error) {
	userinfos := Conmgr.userMap[token]
	if userinfos == nil {
		return nil, errors.New("请重新登陆")
	}
	data := userinfos.([]interface{})
	if len(data) < 2 {
		return nil, nil
	}
	return data[labelsCacheStatus].([]*model.WeixinOauserTaguser), nil
}

// GetLabelNamesByToken GetLabelNamesByToken
func GetLabelNamesByToken(token string) ([]string, error) {
	labels, err := GetLabelsByToken(token)
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

// ForgetPass 忘记密码
func ForgetPass(c *model.Container) error {
	// 检查参数
	if len(c.Body.Metrics) == 0 {
		return errors.New("参数不能为空{body:{'metrics':''}}")
	}
	if !util.IsEmail(c.Body.Metrics) {
		return errors.New("邮箱格式不对")
	}
	// 先检测邮件是否已经注册
	users, err := service.FindAllUserInfo(map[string]interface{}{"email": c.Body.Metrics})
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if len(users) == 0 {
		return errors.New("邮箱没有注册,找管理员")
	}
	// 重置密码
	rand := util.RandomNumbers(5)

	err = service.UpdatePassword(users[0].ID, rand)
	if err != nil {
		return errors.New("重置密码失败,找管理员")
	}
	// 发送密码给用户
	err = service.SendMail([]string{c.Body.Metrics}, config.Config.EmailUser, "修改密码", "登陆密码:"+rand)
	if err != nil {
		c.Header.Msg = "修改成功，请查看邮箱"
	}
	return nil

}

// Logout 登出
func Logout(token string) error {
	// 清除用户缓存
	Conmgr.userMap[token] = nil
	return nil
}

// HasPermission HasPermission
func HasPermission(token, url string) (bool, error) {
	// 查询用户可访问的路径
	permissions, err := GetPermissionByToken(token)
	if err != nil {
		return false, err
	}
	// url去掉ip和Port
	if strings.Contains(url, "://") {
		url = strings.ReplaceAll(url, "://", "")
		url = url[strings.Index(url, "/"):]
	}
	// 判断是否包含Url
	var flag bool
	for _, p := range permissions {
		if len(p) <= len(url) && strings.HasPrefix(url, p) {
			flag = true
			break
		}
	}
	return flag, nil
}
