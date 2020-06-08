package conmgr

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codepository/user/config"
	"github.com/codepository/user/model"
	"github.com/codepository/user/service"
)

const (
	departmentInfo = "部门信息缓存"
	labelsCache    = "标签缓存"
)

// Conmgr 程序唯一的一个连接管理器
// 处理定时任务
var Conmgr *ConnManager

// ConnManager 连接管理器
type ConnManager struct {
	start int32
	stop  int32
	quit  chan struct{}

	cacheMap     map[string]interface{}
	cacheMapLock sync.RWMutex

	userMap     map[string]interface{}
	userMapLock sync.RWMutex
}

// Start 启动连接管理器
func (cm *ConnManager) Start() {
	// 是否已经启动
	if atomic.AddInt32(&cm.start, 1) != 1 {
		return
	}
	log.Println("启动连接管理器")
	// 定时任务
	go cronTaskStart(cm)
	// 获取用户信息缓存
	refreshUserMap()

}

// Stop 停止连接管理器
func (cm *ConnManager) Stop() {
	if atomic.AddInt32(&cm.stop, 1) != 1 {
		log.Println("连接管理器已经关闭")
		return
	}
	close(cm.quit)
	log.Println("关闭连接管理器")
}

// New 新建一个连接管理器
func New() {
	cm := ConnManager{
		quit:     make(chan struct{}),
		cacheMap: make(map[string]interface{}),
		userMap:  make(map[string]interface{}),
	}
	Conmgr = &cm
	Conmgr.Start()
}

// cronTaskStart 启动定时任务
func cronTaskStart(cm *ConnManager) {
	log.Println("启动定时任务")
out:
	for {
		now := time.Now()
		next := now.Add(time.Hour * 24)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 4, 0, 0, next.Location())
		// next := now.Add(time.Second * 10)
		t := time.NewTimer(next.Sub(now))
		select {
		// 连接管理器终止时退出
		case <-cm.quit:
			break out
		case <-t.C:
			// 刷新缓存表
			go RefreshCacheMap()
			// 添加新用户
			go CheckIfHaveNewUser()
			time.Sleep(10 * time.Second)
			go refreshUserMap()
		}
	}
}

// RefreshCacheMap 刷新cacheMap中的内容
func RefreshCacheMap() {
	clearCacheMap()
}

// refreshUserMap 更新用户信息
func refreshUserMap() {
	log.Println("刷新用户信息缓存")
	clearUserMap()
	// 查询所有的用户
	users, err := service.FindAllUserInfo(map[string]interface{}{})
	if err != nil {
		sendRecord(config.Errors[config.UpdateUserMapErr], "", 0, err)
		return
	}

	for _, user := range users {
		err := CaceheUserInfo(user)
		if err != nil {
			sendRecord("查询用户信息错误", fmt.Sprintf("用户ID:%d", user.ID), 0, err)
			continue
		}
	}

}

// CaceheUserInfo 缓存用户信息
func CaceheUserInfo(user *model.Userinfo) error {
	// 再查询每个用户对应的token
	login := &model.Login{}
	login.UserID = user.ID
	login.Find()

	token := login.Password
	// 再查询每个用户的信息
	data, err := service.GetAllUserinfo(user)
	if err != nil {
		return err
	}
	// 存入userMap
	Conmgr.userMap[token] = data
	// str, _ := util.ToJSONStr(data)
	// log.Printf("%s,%s\n", token, str)
	return nil
}

// Login 登陆
func Login(account, password string) (*model.Container, error) {

	result, err := service.UserLoginAndFindUserinfo(account, password)
	if err != nil {
		return nil, err
	}
	// 缓存至userMap
	Conmgr.userMap[result.Header.Token] = result.Body.Data
	return result, nil
}

// GetUserinfo 用户数据
func GetUserinfo(c *model.Container) error {
	if len(c.Header.Token) == 0 {
		return errors.New(`token不能为空,{header:{token:""}}`)
	}
	data, _ := Conmgr.userMap[c.Header.Token].([]interface{})
	if data == nil {
		return errors.New("请重新登陆token已经失效")
	}
	c.Body.Data = data
	return nil
}

// FindAllDepartment 查询部门树形结构
func FindAllDepartment(c *model.Container) error {
	if Conmgr.cacheMap[departmentInfo] == nil {
		result, err := service.FindAllDepartment()
		if err != nil {
			return err
		}
		Conmgr.cacheMap[departmentInfo] = result
	}
	c.Body.Data = append(c.Body.Data, Conmgr.cacheMap[departmentInfo])
	return nil
}

// GetUserByID 根据用户ID从缓存中查询用户信息
func GetUserByID(c *model.Container) error {
	if len(c.Body.Metrics) == 0 {
		return errors.New(`参数如:{body:{metrics:"1"}},1为用户ID`)
	}
	id, err := strconv.ParseInt(c.Body.Metrics, 10, 64)
	if err != nil {
		return err
	}
	// 再查询每个用户对应的token
	login := &model.Login{}
	login.UserID = int(id)
	login.Find()
	token := login.Password
	if Conmgr.userMap[token] == nil {

	}
	c.Body.Data = Conmgr.userMap[token].([]interface{})
	return nil

}

// CheckIfHaveNewUser 是否有新用户
func CheckIfHaveNewUser() {
	// 是否有新用户
	ids, err := service.CheckIfHaveNewUser()
	if err != nil {
		sendRecord("检查新用户", "", 0, err)
		return
	}
	if len(ids) == 0 {
		return
	}
	// 为用户设置登陆密码
	for _, id := range ids {
		err := service.SetPassWord(id, "123")
		if err != nil {
			sendRecord(config.Errors[config.CheckNewUserErr], fmt.Sprintf("id为%d", id), 0, err)
			continue
		}
		// // 用户添加新用户标签
		// err = service.AddLabel(id, int(model.New), model.LabelTypes[model.New])
		// if err != nil {
		// 	sendRecord(config.Errors[config.AddLabelErr], fmt.Sprintf("id:%d,labelid:%d,labelname:%s", id, int(model.New), model.LabelTypes[model.New]), 0, err)
		// }
	}
}

// DelLabel 删除标签
func DelLabel(c *model.Container) error {
	// 参数判断
	if c.Body.Data == nil {
		return errors.New("参数类型必须是:{'body':{'data':[{'user_id':1,'label_id:1},{'user_id':2,'label_id:1}]}}")
	}
	if len(c.Body.Data) == 0 {
		return errors.New("data不能为空")
	}
	// data, ok := c.Body.Data[0].([]interface{})
	// if !ok {
	// 	return errors.New("参数类型:{'body':{'data':[{'user_id':1,'label_id:0,'label_name':'一线考核'}]}}")
	// }
	data := c.Body.Data
	var strbuffer strings.Builder
	var userids []int
	// 删除标签
	for i, v := range data {
		ul, ok := v.(map[string]interface{})
		if !ok {
			strbuffer.WriteString(fmt.Sprintf("第[%d]条删除失败,err:%s,", (i + 1), "参数类型必须是:{'body':{'data':[{'user_id':1,'label_id:1}]}}"))
			continue
		}
		userlabel := &model.UserLabel{}
		err := userlabel.FromMap(ul)
		if err != nil {
			strbuffer.WriteString(fmt.Sprintf("第[%d]条删除失败,err:%s,", (i + 1), err.Error()))
			continue
		}
		err = userlabel.DelSingle()
		if err != nil {
			strbuffer.WriteString(fmt.Sprintf("第[%d]条删除失败,err:%s,", (i + 1), err.Error()))
			continue
		}
		userids = append(userids, userlabel.UserID)

	}
	c.Body.Data = nil
	if len(strbuffer.String()) != 0 {
		return errors.New(strbuffer.String())
	}
	c.Header.Msg = "删除成功"
	// 更新相关用户的用户信息缓存
	err := cacheUserinfoByIDs(userids)
	return err
}

// AddUserLabel 添加用户标签
func AddUserLabel(c *model.Container) error {
	if c.Body.Data == nil {
		return errors.New("参数类型:{'body':{'data':[{'user_id':1,'label_id:0,'label_name':'一线考核'}]}}")
	}
	if len(c.Body.Data) == 0 {
		return errors.New("data不能为空")
	}
	// data, ok := c.Body.Data[0].([]interface{})
	data := c.Body.Data
	// if !ok {
	// 	return errors.New("参数类型:{'body':{'data':[[{'user_id':1,'label_id:0,'label_name':'一线考核'}]]}}")
	// }
	var strbuffer strings.Builder
	var userids []int
	for i, v := range data {
		ul, ok := v.(map[string]interface{})
		if !ok {
			strbuffer.WriteString(fmt.Sprintf("第[%d]条添加失败,err:%s,", (i + 1), "参数类型:{'body':{'data':[{'user_id':1,'label_id:0,'label_name':'一线考核'}]}}"))
			continue
		}
		userlabel := &model.UserLabel{}
		err := userlabel.FromMap(ul)
		if err != nil {
			strbuffer.WriteString(fmt.Sprintf("第[%d]条添加失败,err:%s,", (i + 1), err.Error()))
			continue
		}
		err = userlabel.SaveOrUpdate()
		if err != nil {
			strbuffer.WriteString(fmt.Sprintf("第[%d]条添加失败,err:%s,", (i + 1), err.Error()))
			continue
		}
		userids = append(userids, userlabel.UserID)
	}
	c.Body.Data = nil
	if len(strbuffer.String()) != 0 {
		return errors.New(strbuffer.String())
	}
	// 更新相关用户的用户信息缓存
	c.Header.Msg = "添加成功"
	err := cacheUserinfoByIDs(userids)
	return err
}
func cacheUserinfoByIDs(ids []int) error {
	// 更新相关用户的用户信息缓存
	userinfos, err := service.FindAllUserInfo(ids)
	if err != nil {
		return errors.New("添加标签成功，但是更新用户信息缓存失败,err:" + err.Error())
	}
	for _, user := range userinfos {
		err := CaceheUserInfo(user)
		if err != nil {
			return err
		}
	}
	return nil
}

// clearCacheMap 清空ClearCacheMap
func clearCacheMap() {
	Conmgr.cacheMapLock.Lock()
	defer Conmgr.cacheMapLock.Unlock()
	len := len(Conmgr.cacheMap)
	if len > 0 {
		//清空 map 的唯一办法就是重新 make 一个新的 map，不用担心垃圾回收的效率，Go语言中的并行垃圾回收效率比写一个清空函数要高效的多。
		Conmgr.cacheMap = map[string]interface{}{}
	}
}
func clearUserMap() {
	Conmgr.userMapLock.Lock()
	defer Conmgr.userMapLock.Unlock()
	len := len(Conmgr.userMap)
	if len > 0 {
		Conmgr.userMap = map[string]interface{}{}
	}
}

// sendRecord 发送纪录
func sendRecord(typename, data string, flag uint8, err error) {
	record := model.Record{
		Data: data,
		Type: typename,
		Flag: flag,
		Err:  err.Error(),
	}
	record.Save()
}
