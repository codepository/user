package conmgr

import (
	"errors"
	"log"
	"reflect"

	"github.com/codepository/user/model"
)

// AddLeadership 添加分管部门
func AddLeadership(c *model.Container) error {
	// 参数检查
	datas := c.Body.Data
	if datas == nil || len(datas) < 2 {
		return errors.New(`查询参数不能为空,如{"body":{"data":[[{"id":1,"name":"部门1"},{"id":2,"name":"部门2"}],[3]]}},3为用户id`)
	}

	// 参数提取
	// ds, yes := datas[0].([]float64)
	// if !yes {
	// 	return errors.New("部门id应该为数组")
	// }
	ds, yes := datas[0].([]interface{})
	if !yes {
		return errors.New(`查询参数不能为空,如{"body":{"data":[[{"id":1,"name":"部门1"},{"id":2,"name":"部门2"}],[3]]}},3为用户id`)
	}
	uid, yes := datas[1].(float64)
	if !yes {
		return errors.New("用户id必须为数字")
	}
	// 添加
	for _, d := range ds {
		dm, yes := d.(map[string]interface{})
		log.Println(reflect.TypeOf(d))
		if !yes {
			return errors.New(`查询参数不能为空,如{"body":{"data":[[{"id":1,"name":"部门1"},{"id":2,"name":"部门2"}],[3]]}},3为用户id`)
		}
		f := model.FznewsLeadership{UserID: int(uid), DepartmentID: int(dm["id"].(float64)), DepartmentName: dm["name"].(string)}
		err := f.SaveOrUpdate()
		if err != nil {
			return err
		}
	}
	c.Header.Msg = "添加成功"
	// 更新用户分管部门缓存
	return cacheUserinfoByIDs([]int{int(uid)})
}

// DelByIDLeadership 根据id来删除分管部门
func DelByIDLeadership(c *model.Container) error {

	// 参数检查
	if c.Body.Data == nil || c.Body.Data[0] == nil {
		return errors.New(`查询参数不能为空,如{"body":{"data":[[1,2]]}},删除id为1，2的条目`)
	}
	// 参数提取
	params, yes := c.Body.Data[0].([]interface{})
	if !yes {
		return errors.New(`查询参数不能为空,如{"body":{"data":[[1,2]]}},删除id为1，2的条目`)
	}
	// 删除
	return model.DelFznewsLeadership(params)

}

// FindLeadership 查询
func FindLeadership(c *model.Container) error {
	// 参数检查
	if c.Body.Data == nil || c.Body.Data[0] == nil {
		return errors.New(`查询参数不能为空,如{"body":{"data":[{"user_id",2}]}},查询用户2的分管部门`)
	}
	// 查询
	params, yes := c.Body.Data[0].(map[string]interface{})
	if !yes {
		return errors.New(`查询参数必须为json`)
	}
	result, err := model.FindFznewsLeadership(params)
	if err != nil {
		return err
	}
	c.Body.Data = c.Body.Data[:0]
	c.Body.Data = append(c.Body.Data, result)
	return nil

}
