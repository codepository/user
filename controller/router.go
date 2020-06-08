package controller

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codepository/user/conmgr"
	"github.com/codepository/user/model"
	"github.com/codepository/user/service"
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
	authority []int
}

var routers []*RouteHandler

// SetRouters 设置路由
func SetRouters() {
	routers = []*RouteHandler{
		{route: "visit/user/userinfoByToken", handler: conmgr.GetUserinfo},
		{route: "visit/user/getUserByID", handler: conmgr.GetUserByID},
		// 用户添加角色
		{route: "exec/user/addlabel", handler: conmgr.AddUserLabel, meta: &RouteMeta{authority: []int{7}}},
		// 批量添加用户标签
		{route: "exec/user/addlabelbyDepartment", handler: service.AddlabelbyDepartment, meta: &RouteMeta{authority: []int{7}}},
		// 删除用户标签
		{route: "exec/user/dellabel", handler: conmgr.DelLabel, meta: &RouteMeta{authority: []int{7}}},
		{route: "visit/user/findbylabelids", handler: service.FindUsersByLabelIDs},
		{route: "visit/user/getUsers", handler: service.GetUsers},
		{route: "exec/user/forgetPass", handler: conmgr.ForgetPass},
		{route: "exec/user/alterPass", handler: conmgr.AlterPass},
		{route: "visit/department/all", handler: conmgr.FindAllDepartment},
		// 查询所有标签
		{route: "visit/lable/all", handler: conmgr.FindAllLabel},
		// 添加新的标签
		{route: "exec/label/add", handler: conmgr.AddNewLabel, meta: &RouteMeta{authority: []int{7}}},
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
	labels, err := conmgr.GetLabelsByToken(token)
	if err != nil {
		return err
	}
	// 权限匹配
	for _, a := range f.meta.authority {
		for _, l := range labels {
			if l.LabelID == a {
				return nil
			}
		}
	}
	labs, err := conmgr.GetLabelNamesByIds(f.meta.authority)
	if err != nil {
		return err
	}

	return fmt.Errorf("需要权限:%v", strings.Join(labs, ","))

}
