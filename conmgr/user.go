package conmgr

import (
	"errors"

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
func GetLabelsByToken(token string) ([]string, error) {
	userinfos := Conmgr.userMap[token]
	if userinfos == nil {
		return nil, errors.New("请重新登陆")
	}
	data := userinfos.([]interface{})
	if len(data) < 2 {
		return nil, nil
	}
	return data[labelsCacheStatus].([]string), nil
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
