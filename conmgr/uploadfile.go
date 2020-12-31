package conmgr

import (
	"fmt"
	"strings"

	"github.com/codepository/user/model"
	"github.com/mumushuiding/util"
)

// DelUploadfileByID 根据ID删除
func DelUploadfileByID(c *model.Container) error {
	errstr := `参数类型:{"body":{"params":{"id":0}}} 文件id不能为空`
	if len(c.Body.Params) == 0 || c.Body.Params["id"] == nil {
		return fmt.Errorf(errstr)
	}
	id, err := util.Interface2Int(c.Body.Params["id"])
	if err != nil {
		return err
	}
	err = model.DelUploafileByID(id)
	return err
}

// FindUploadfile 查询上传至数据库的文件
func FindUploadfile(c *model.Container) error {
	errstr := `参数类型:{"body":{"params":{"fields":"id,filename","filetype":"remarks","filename":"文件名","limit":10,"offset":0}}} 参数不能全为空`
	if len(c.Body.Params) == 0 {
		return fmt.Errorf(errstr)
	}
	var sqlbuff strings.Builder
	var err error
	fields := "id,filename"
	if c.Body.Params["fields"] != nil {
		fields = c.Body.Params["fields"].(string)
	}
	if c.Body.Params["filetype"] != nil {
		sqlbuff.WriteString(fmt.Sprintf("and filetype='%s' ", c.Body.Params["filetype"].(string)))
	}
	if c.Body.Params["filename"] != nil {
		filename, ok := c.Body.Params["filename"].(string)
		if !ok {
			return fmt.Errorf("filename 必须为字符串")
		}
		if len(filename) > 0 {
			sqlbuff.WriteString(fmt.Sprintf("and filename like '%s' ", "%"+c.Body.Params["filename"].(string)+"%"))
		}
	}
	limit := 10
	offset := 0
	if c.Body.Params["limit"] != nil {
		limit, err = util.Interface2Int(c.Body.Params["limit"])
		if err != nil {
			return fmt.Errorf("limit:%s", err.Error())
		}
	}
	if c.Body.Params["offset"] != nil {
		offset, err = util.Interface2Int(c.Body.Params["offset"])
		if err != nil {
			return fmt.Errorf("offset:%s", err.Error())
		}
	}
	sql := ""
	if sqlbuff.Len() != 0 {
		sql = sqlbuff.String()[4:]
	}
	datas, err := model.FindUploadfiles(fields, "", limit, offset, sql)
	if err != nil {
		return fmt.Errorf("查询上传文件:%s", err.Error())
	}
	c.Body.Data = append(c.Body.Data, datas)
	return nil

}
