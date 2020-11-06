package conmgr

import (
	"fmt"

	"github.com/codepository/user/model"
	"github.com/mumushuiding/util"
)

// UpdateDepartment 更新部门
func UpdateDepartment(c *model.Container) error {
	errstr := `参数格式:{"body":"params":{"data":[{"id":12,"attribute":1},{"id":2,"attribute":2}]}},注意字段中值为"",0和false的字段不更新`
	if len(c.Body.Params) == 0 || c.Body.Params["data"] == nil {
		return fmt.Errorf(errstr)
	}
	datas, ok := c.Body.Params["data"].([]interface{})
	if !ok {
		return fmt.Errorf("更新部门:data 必须为数组")
	}
	var deptMaps []map[string]interface{}
	for _, d := range datas {
		deptMap, ok := d.(map[string]interface{})
		if !ok {
			return fmt.Errorf(errstr)
		}
		deptMaps = append(deptMaps, deptMap)
	}
	var depts []*model.WeixinLeaveDepartment
	for _, m := range deptMaps {
		var dept model.WeixinLeaveDepartment
		str, _ := util.ToJSONStr(m)
		err := util.Str2Struct(str, &dept)
		if err != nil {
			return fmt.Errorf("更新部门:%s", err.Error())
		}
		depts = append(depts, &dept)

	}
	for _, dept := range depts {
		if dept.ID == 0 {
			return fmt.Errorf("更新部门:部门的id不能为空")
		}
		err := dept.Updates()
		if err != nil {
			return fmt.Errorf("更新部门:%s", err.Error())
		}
	}
	return nil
}
