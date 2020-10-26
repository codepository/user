package controller

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codepository/user/conmgr"
	"github.com/codepository/user/model"
	"github.com/codepository/user/service"
)

const (
	// SysManagerAuthority SysManagerAuthority
	SysManagerAuthority = "系统管理员"
	// AdvertiseAuthority 广告管理权限
	AdvertiseAuthority = "广告管理"
)

// RouteFunction 根据路径指向方法
type RouteFunction func(*model.Container) error

// RouteHandler 路由
type RouteHandler struct {
	handler RouteFunction
	route   string
	meta    *RouteMeta
}

// RouteMeta 路由参数
type RouteMeta struct {
	// 可以访问该路径的所有角色的id
	authority []string
}

var routers []*RouteHandler

// SetRouters 设置路由
func SetRouters() {
	routers = []*RouteHandler{
		// =================== 用户 ===================
		{route: "visit/user/userinfoByToken", handler: conmgr.GetUserinfo},
		{route: "visit/user/getUserByID", handler: conmgr.GetUserByID},
		// 用户添加角色
		{route: "exec/user/addlabel", handler: conmgr.AddUserLabel, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		// 批量添加用户标签
		{route: "exec/user/addlabelbyDepartment", handler: service.AddlabelbyDepartment, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		// 删除用户标签
		{route: "exec/user/dellabel", handler: conmgr.DelLabel, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		{route: "visit/user/findbylabelids", handler: service.FindUsersByLabelIDs},
		//查询用户 {"body":{"username":"张三","paged":false,"method":"visit/user/getUsers"}}
		{route: "visit/user/getUsers", handler: service.GetUsers},
		{route: "exec/user/forgetPass", handler: conmgr.ForgetPass},
		{route: "exec/user/alterPass", handler: conmgr.AlterPass},
		// 根据标签和职级查询用户id,level是职级
		{route: "visit/user/getUseridsByTagAndLevel", handler: service.GetUseridsByTagAndLevel},
		// 查询所有标签
		{route: "visit/lable/all", handler: conmgr.FindAllLabel},
		// 添加新的标签
		{route: "exec/label/add", handler: conmgr.AddNewLabel, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		// 查询用户标签
		{route: "visit/label/findall", handler: conmgr.FindUserLabel},
		// 启动流程
		{route: "exec/flow/startByToken", handler: conmgr.StartFlowByToken, meta: &RouteMeta{
			authority: []string{"%考核组成员"},
		}},
		{route: "exec/flow/delete", handler: conmgr.DeleteFlow},
		// 审批流程
		{route: "exec/flow/completeFlowTask", handler: conmgr.CompleteFlowTask},
		// 待审批流程查询
		{route: "visit/flow/task", handler: conmgr.FindFlowTask},
		// 审批纪录
		{route: "visit/flow/log", handler: conmgr.FindFlowLog},
		// 流程查询
		{route: "visit/flow/myprocess", handler: conmgr.FindFlowMyProcess},
		// 查询所有流程
		{route: "visit/flow/findall", handler: conmgr.FindAllFlow},
		// 查看流程是否已经存在
		{route: "visit/flow/exist", handler: conmgr.FlowExist},
		// 查看流程进度
		{route: "visit/flow/flowStepper", handler: conmgr.FlowStepper},
		//=============== 分管领导管理===================
		{route: "exec/leader/add", handler: conmgr.AddLeadership, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		{route: "exec/leader/delbyid", handler: conmgr.DelByIDLeadership, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		{route: "visit/leader/find", handler: conmgr.FindLeadership},
		// ===============行业查询权限管理================
		{route: "visit/org/findOrgidsByUserid", handler: conmgr.FindOrgidsByUserid},
		{route: "visit/org/findUserByOrgid", handler: conmgr.FindUserByOrgid},
		{route: "visit/org/delUserOrgByIds", handler: conmgr.DelUserOrgByID},
		{route: "visit/org/saveUserOrg", handler: conmgr.SaveUserOrg, meta: &RouteMeta{authority: []string{AdvertiseAuthority}}},
		// ===============任务=====================
		{route: "exec/task/yxkh", handler: conmgr.NewYxkhTask},
		{route: "visit/task/completeRate", handler: conmgr.TaskCompleteRate},
		{route: "visit/task/uncomplete", handler: conmgr.TaskUncomplete},
		{route: "exec/task/complete", handler: conmgr.CompleteTask},
		// 任务 查询任务对应的角色组
		{route: "visit/task/taskRoles", handler: conmgr.FindTaskRoles},
		// ====================== 部门 ==========================
		// 查询部门：{"body":{"refresh":false,"paged":true,"method":"visit/department/all"}}
		{route: "visit/department/all", handler: conmgr.FindAllDepartment},
		// 同步weixin_leave_leaderrole中分管领导信息，从weixin_leave_userinfo同步部门领导信息
		{route: "exec/department/sync", handler: conmgr.SyncLeader},
	}
}

// GetRoute 获取执行函数
func GetRoute(route, token string) (func(*model.Container) error, error) {
	var f *RouteHandler
	for _, r := range routers {
		if r.route == route {
			f = r
			break
		}
	}
	if f == nil {
		return nil, errors.New("method:" + route + ",不存在")
	}
	err := checkAuthority(f, token)
	if err != nil {
		return nil, err
	}
	return f.handler, nil
}
func checkAuthority(f *RouteHandler, token string) error {
	if f.meta == nil || len(f.meta.authority) == 0 {
		return nil
	}
	// 查看token
	if len(token) == 0 {
		return errors.New(`token不能为空`)
	}
	// 查看权限
	labels, err := conmgr.GetLabelNamesByToken(token)
	if err != nil {
		return err
	}
	// 权限匹配
	for _, a := range f.meta.authority {
		for _, l := range labels {
			if l == a {
				return nil
			}
			if a[0:1] == "%" && strings.HasSuffix(l, a[1:]) {
				return nil
			}
		}
	}
	// labs, err := conmgr.GetLabelNamesByIds(f.meta.authority)
	// if err != nil {
	// 	return err
	// }

	return fmt.Errorf("需要权限:%v", strings.Join(f.meta.authority, ","))

}
