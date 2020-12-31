package controller

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/codepository/user/conmgr"
	"github.com/codepository/user/model"

	"github.com/mumushuiding/util"
)

// Index 首页
func Index(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Hello world!")
}

// Alive 程序是否存活
func Alive(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "1")
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

// SaveFileToDB 存储文件到数据库
func SaveFileToDB(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		util.ResponseErr(w, "只支持POST方法")
		return
	}
	// 获取文件
	file, filehead, err := req.FormFile("filename")
	if err != nil {
		util.ResponseErr(w, fmt.Sprintf("file upload fail:%s", err.Error()))
		return
	}
	defer file.Close()
	if len(filehead.Filename) == 0 {
		util.ResponseErr(w, "filename不能为空")
		return
	}
	values := req.URL.Query()
	token := values.Get("token")
	filetype := values.Get("filetype")
	if len(token) == 0 {
		token = req.PostFormValue("token")
		if len(token) == 0 {
			util.ResponseErr(w, "参数需要添加token,以get或者post方式添加")
			return
		}
	}
	if len(filetype) == 0 {
		filetype = req.PostFormValue("filetype")
		if len(filetype) == 0 {
			util.ResponseErr(w, "参数需要添加filetype,以get或者post方式添加")
			return
		}
	}
	user, err := conmgr.GetUserByToken(token)
	if user == nil {
		util.ResponseErr(w, "token 无效")
		return
	}
	// 文件转换成二进制数
	data := make([]byte, filehead.Size)
	_, err = file.Read(data)
	if err != nil {
		util.ResponseErr(w, fmt.Sprintf("文件转换成二进制:%s", err.Error()))
		return
	}
	// 存储文件
	ul := model.FznewsUploadfile{
		Filename: filehead.Filename,
		Filetype: filetype,
		UID:      user.ID,
		Username: user.Name,
		Blob:     data,
	}
	err = ul.Create()
	if err != nil {
		util.ResponseErr(w, fmt.Sprintf("存储至数据库:%s", err.Error()))
		return
	}
	util.ResponseData(w, fmt.Sprintf("%d", ul.ID))
}

// DownloadFileFromDB 从数据库获取文件并下载
func DownloadFileFromDB(w http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	id := values.Get("id")
	if len(id) == 0 {
		id = req.PostFormValue("id")
		if len(id) == 0 {
			util.ResponseErr(w, "参数需要添加id,以get或者post方式添加")
			return
		}
	}
	data, err := model.FindUploadfileByID(id)
	if err != nil {
		util.ResponseErr(w, fmt.Sprintf("查询上传文件:%s", err.Error()))
		return
	}
	// 导出
	fileName := data.Filename
	w.Header().Set("Content-Type", "multipart/form-data;charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", fileName))
	w.Write(data.Blob)
}

// Import 导入xlsx文件并按照指定的方法处理
func Import(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		util.ResponseErr(w, "只支持POST方法")
		return
	}
	//解析 form 中的file 上传名字
	file, filehead, err := req.FormFile("filename")
	if err != nil {
		fmt.Fprintf(w, "file upload fail:%s", err)
	}
	if len(filehead.Filename) == 0 {
		fmt.Fprintf(w, "filename不能为空")
		return
	}
	ss := strings.Split(filehead.Filename, ".")
	if ss[len(ss)-1] != "xlsx" {
		fmt.Fprintf(w, "只支持xlsx")
		file.Close()
		return
	}
	filesave := fmt.Sprintf("%s%d", filehead.Filename, time.Now().Nanosecond())
	//打开 已只读,文件不存在创建 方式打开  要存放的路径资源
	f, err := os.OpenFile(filesave, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		f.Close()
		file.Close()
		os.Remove(filesave)
		fmt.Fprintf(w, "file open fail:%s", err)
		return
	}
	//文件 copy
	_, err = io.Copy(f, file)
	if err != nil {
		fmt.Fprintf(w, "file copy fail:%s", err)
		f.Close()
		file.Close()
		os.Remove(filesave)
		return
	}
	//  文件导入之后，执行操作
	par := model.Container{}
	values := req.URL.Query()
	method := values.Get("method")
	token := values.Get("token")
	if len(token) == 0 {
		token = req.PostFormValue("token")
		if len(token) == 0 {
			fmt.Fprintf(w, "需要添加token,可以以get或者post方式添加")
			return
		}
	}
	if len(method) == 0 {
		method = req.PostFormValue("method")
		if len(method) == 0 {
			fmt.Fprintf(w, "需要添加method参数,可以以get或者post方式添加")
			return
		}
	}
	par.Body.Method = method
	par.Header.Token = token
	par.File = f
	funcs, err := GetRoute(par.Body.Method, par.Header.Token)
	if err != nil {
		util.ResponseErr(w, err)
		f.Close()
		file.Close()
		os.Remove(filesave)
		return
	}
	err = funcs(&par)

	// err = conmgr.GetPublicAssessFromXlsx(f)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
		f.Close()
		file.Close()
		os.Remove(filesave)
		return
	}
	//关闭对应打开的文件
	f.Close()
	file.Close()
	os.Remove(filesave)
	fmt.Fprintf(w, "成功")

}

// Export 以xlsx格式导出数据
func Export(w http.ResponseWriter, r *http.Request) {
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
	errstr := `参数格式:{"body":{"method":"export/yxkh/findAllEvalutionRank","data":[["用户名","部门"],["username","deptName"]]}} data[0]导出文件行标、data[1]表示行标对应的字段，俩者一一对应，都不能为空`
	if len(par.Body.Method) == 0 {
		util.ResponseErr(w, errstr)
		return
	}

	if par.Body.Data == nil || len(par.Body.Data) != 2 {
		util.ResponseErr(w, errstr)
		return
	}
	token := par.Header.Token
	if len(token) == 0 {
		token, _ = GetToken(r)
	}
	categoryHeader := par.Body.Data[0].([]interface{})
	var header []string
	for _, h := range categoryHeader {
		header = append(header, h.(string))
	}
	// if !ok {
	// 	util.ResponseErr(w, errstr)
	// 	return
	// }
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
	datas := par.Body.Data

	// 导出
	fileName := "export.csv"
	b := &bytes.Buffer{}
	wr := csv.NewWriter(b)

	wr.Write(header)
	for i := 0; i < len(datas); i++ {
		wr.Write(datas[i].([]string))
	}
	wr.Flush()
	w.Header().Set("Content-Type", "application/vnd.ms-excel;charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", fileName))
	w.Write(b.Bytes())
}
