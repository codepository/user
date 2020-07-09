package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/codepository/user/conmgr"
	"github.com/codepository/user/model"
	"github.com/codepository/user/service"
	"github.com/mumushuiding/util"
)

const (
	// SysManagerAuthority SysManagerAuthority
	SysManagerAuthority = "系统管理员"
	// AdvertiseAuthority 广告管理权限
	AdvertiseAuthority = "广告管理"
)

// RouteFunction 根据路径指向方法
type RouteFunction func(*model.Container) error

// RouterMap 路由
var RouterMap map[string]*RouteHandler

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
		{route: "visit/user/userinfoByToken", handler: conmgr.GetUserinfo},
		{route: "visit/user/getUserByID", handler: conmgr.GetUserByID},
		// 用户添加角色
		{route: "exec/user/addlabel", handler: conmgr.AddUserLabel, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		// 批量添加用户标签
		{route: "exec/user/addlabelbyDepartment", handler: service.AddlabelbyDepartment, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		// 删除用户标签
		{route: "exec/user/dellabel", handler: conmgr.DelLabel, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		{route: "visit/user/findbylabelids", handler: service.FindUsersByLabelIDs},
		{route: "visit/user/getUsers", handler: service.GetUsers},
		{route: "exec/user/forgetPass", handler: conmgr.ForgetPass},
		{route: "exec/user/alterPass", handler: conmgr.AlterPass},
		{route: "visit/department/all", handler: conmgr.FindAllDepartment},
		// 查询所有标签
		{route: "visit/lable/all", handler: conmgr.FindAllLabel},
		// 添加新的标签
		{route: "exec/label/add", handler: conmgr.AddNewLabel, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		// 启动流程
		{route: "exec/flow/startByToken", handler: conmgr.StartFlowByToken, meta: &RouteMeta{
			authority: []string{"%考核组成员"},
		}},
		// 分管领导管理
		{route: "exec/leader/add", handler: conmgr.AddLeadership, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		{route: "exec/leader/delbyid", handler: conmgr.DelByIDLeadership, meta: &RouteMeta{authority: []string{SysManagerAuthority}}},
		{route: "exec/leader/find", handler: conmgr.FindLeadership},
		// 行业查询权限管理
		{route: "visit/org/findOrgidsByUserid", handler: conmgr.FindOrgidsByUserid},
		{route: "visit/org/findUserByOrgid", handler: conmgr.FindUserByOrgid},
		{route: "visit/org/delUserOrgByIds", handler: conmgr.DelUserOrgByID},
		{route: "visit/org/saveUserOrg", handler: conmgr.SaveUserOrg, meta: &RouteMeta{authority: []string{AdvertiseAuthority}}},
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
		return errors.New("token为空,可能没登陆")
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

// HasPermission 指定角色是否有访问指定路径的权限
func HasPermission(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if len(r.Form["token"]) == 0 {
		util.ResponseErr(w, "参数token不能为空")
		return
	}
	token := r.Form["token"][0]
	if len(r.Form["url"]) == 0 {
		util.ResponseErr(w, "参数url不能为空")
	}
	url := r.Form["url"][0]
	result, err := conmgr.HasPermission(token, url)
	if err != nil {
		fmt.Fprintf(w, "{\"message\":\"%s\",\"flag\":%t,\"status\":400}", err.Error(), false)
		return
	}
	if result {
		fmt.Fprintf(w, "{\"message\":\"%s\",\"flag\":%t,\"status\":200}", "允许访问", result)
		return
	}
	fmt.Fprintf(w, "{\"message\":\"%s\",\"ok\":%t,\"status\":200}", "无访问"+url+"权限", result)
}
