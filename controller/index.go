package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/codepository/user/conmgr"
	"github.com/codepository/user/model"

	"github.com/mumushuiding/util"
)

// Index 首页
func Index(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Hello world!")
}

// GetToken 获取token
func GetToken(request *http.Request) (string, error) {
	token := request.Header.Get("Authorization")
	if len(token) == 0 {
		request.ParseForm()
		if len(request.Form["token"]) == 0 {
			return "", errors.New("header Authorization 没有保存 token, url参数也不存在 token， 访问失败 ！")
		}
		token = request.Form["token"][0]
	}
	return token, nil
}

// GetParam 获取url链接中指定名称的参数值
func GetParam(parameterName string, request *http.Request) string {
	if len(request.Form[parameterName]) > 0 {
		return request.Form[parameterName][0]
	}
	return ""
}

// GetBody2Struct 获取POST参数，并转化成指定的struct对象
func GetBody2Struct(request *http.Request, pojo interface{}) error {
	s, _ := ioutil.ReadAll(request.Body)
	if len(s) == 0 {
		return nil
	}
	return json.Unmarshal(s, pojo)
}

// Logout 登出
func Logout(w http.ResponseWriter, r *http.Request) {
	// 获取token
	token, err := GetToken(r)
	if err != nil {
		util.ResponseErr(w, err)
		return
	}
	err = conmgr.Logout(token)
	if err != nil {
		util.ResponseErr(w, err)
		return
	}
	util.ResponseOk(w)
}

// Login 登陆
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		util.ResponseErr(w, "只支持POST方法")
		return
	}
	fields, err := util.Body2Map(r)
	if err != nil {
		util.ResponseErr(w, err)
		return
	}
	account, ok := fields["account"].(string)
	if !ok {
		util.ResponseErr(w, "account必须为字符串")
		return
	}
	if len(account) == 0 {
		util.ResponseErr(w, "account不能为空")
		return
	}
	password, ok := fields["password"].(string)
	if !ok {
		util.ResponseErr(w, "password必须为字符串")
		return
	}
	if len(password) == 0 {
		util.ResponseErr(w, "password不能为空")
		return
	}
	c, err := conmgr.Login(account, password)
	if err != nil {
		util.ResponseErr(w, err)
		return
	}
	util.ResponseData(w, c.ToString())
}

// CheckNewUser CheckNewUser
func CheckNewUser(w http.ResponseWriter, r *http.Request) {
	conmgr.CheckIfHaveNewUser()
}

// GetData 查询接口
func GetData(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		util.ResponseErr(w, "只支持POST方法")
		return
	}
	var par model.Container
	err := util.Body2Struct(r, &par)
	if err != nil {
		util.ResponseErr(w, err)
		return
	}
	if len(par.Body.Method) == 0 {
		util.ResponseErr(w, "method不能为空")
		return
	}
	token := par.Header.Token
	if len(token) == 0 {
		token, _ = GetToken(r)
	}
	f, err := GetRoute(par.Body.Method, par.Header.Token)
	if err != nil {
		util.ResponseErr(w, err)
		return
	}
	err = f(&par)
	if err != nil {
		util.ResponseErr(w, err)
		return
	}
	if !par.Body.Paged {
		util.ResponseData(w, par.ToString())
		return
	}
	result, err := par.ToPageJSON()
	if err != nil {
		util.ResponseErr(w, err)
		return
	}
	fmt.Fprintf(w, result)
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
