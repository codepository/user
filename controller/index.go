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
	util.ResponseData(w, par.ToString())
}