package conmgr

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/codepository/user/model"
	"github.com/mumushuiding/util"
)

// SyncLeader 同步部门的分管领导和部门领导
func SyncLeader(c *model.Container) error {
	log.Println("sync leader start")
	var err1 error
	var err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	// 同步分管领导
	go func() {
		err1 = syncLeadersInCharge()
		wg.Done()
	}()
	// 同步部门领导
	go func() {
		err2 = syncDepartmentLeader()
		wg.Done()
	}()

	wg.Wait()
	log.Println("sync leader finish")
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

// syncLeadersInCharge 同步分管领导
func syncLeadersInCharge() error {

	// 从目标表查询所有分管领导及分管部门
	leaders, err := model.FindAllLeaveLeaderrole("")
	if err != nil {
		return err
	}
	// 遍历leaders
	for _, l := range leaders {
		//
		var dids []int
		if len(l.Dept) > 0 {
			depts := strings.Split(l.Dept, ",")
			for _, d := range depts {
				did, _ := strconv.Atoi(d)
				dids = append(dids, did)
			}
		}
		var err1 error
		var err2 error
		// 判断是否存在，不存在就更新
		for _, d := range dids {
			var departmentname string
			var uid int
			dept, err := GetDepartmentFromCache(d)
			if err != nil {
				return err
			}
			if dept != nil {
				departmentname = dept.Name
			}
			user, err := GetUserinfoFromCacheByWxUserid(l.Userid)
			if err != nil {
				return err
			}
			if user != nil {
				uid = user.ID
			}
			fl := model.FznewsLeadership{
				UserID:         uid,
				Userid:         l.Userid,
				Role:           l.Role,
				DepartmentID:   d,
				Username:       l.Username,
				DepartmentName: departmentname,
			}
			err1 = fl.FirstOrCreate()
			if err1 != nil {
				return err1
			}
		}

		// 删除多余的
		if len(dids) > 0 {

			err2 = model.DelFznewsLeadership("userid=? and role=? and department_id not in (?) ", l.Userid, l.Role, strings.Split(l.Dept, ","))

		} else {

			err2 = model.DelFznewsLeadership("userid=? and role=?", l.Userid, l.Role)

		}

		if err1 != nil {
			return err1
		}
		if err2 != nil {
			return err2
		}

	}
	log.Println("syncLeadersInCharge finish")
	return nil
}

// syncDepartmentLeader 同步部门领导
func syncDepartmentLeader() error {
	err := model.DelFznewsLeadership("role=5")
	if err != nil {
		return err
	}
	// 查出所有的领导
	leaders, err := model.FindAllUserInfo("", "is_leader=1 and status=1")
	if err != nil {
		return err
	}
	// 查询纪录是否存在，不存在就添加
	for _, l := range leaders {
		fl := model.FznewsLeadership{
			Userid:         l.Userid,
			UserID:         l.ID,
			Role:           5,
			DepartmentID:   l.DepartmentID,
			DepartmentName: l.Departmentname,
			Username:       l.Name,
		}
		err := fl.FirstOrCreate()
		if err != nil && strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
			return err
		}
	}

	return nil
}

// AddLeadership 添加分管部门
func AddLeadership(c *model.Container) error {
	// 参数检查
	errstr := `查询参数不能为空,如{"body":{"params":{"role":5,"userid":"linting"},"data":[[{"id":1,"name":"部门1"},{"id":2,"name":"部门2"}]]}},linting为用户微信id,role不能为空，role：5部门领导、6分管领导`
	datas := c.Body.Data
	if datas == nil || len(datas) < 1 {
		return errors.New(errstr)
	}
	ds, yes := datas[0].([]interface{})
	if !yes {
		return errors.New(errstr)
	}
	if c.Body.Params["role"] == nil || c.Body.Params["userid"] == nil {
		return errors.New(errstr)
	}
	role, _ := util.Interface2Int(c.Body.Params["role"])
	user, err := GetUserinfoFromCacheByWxUserid(c.Body.Params["userid"].(string))
	if err != nil {
		return err
	}
	// 添加分管部门
	for _, d := range ds {
		dm, yes := d.(map[string]interface{})
		// log.Println(reflect.TypeOf(d))
		if !yes {
			return errors.New(errstr)
		}
		f := model.FznewsLeadership{UserID: user.ID, Userid: user.Userid, Username: user.Name, DepartmentID: int(dm["id"].(float64)), DepartmentName: dm["name"].(string), Role: role}
		err := f.SaveOrUpdate()
		if err != nil {
			return err
		}
	}
	c.Header.Msg = "添加成功"
	// 更新用户分管部门缓存
	return cacheUserinfoByIDs([]int{user.ID})
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
	// 查询
	if len(c.Body.Params) == 0 {
		return errors.New(`查询参数不能为空,如{"body":{"params":{"department_name":"日报","department_id":6,"user_id":1}} 三个参数不能同时为空`)
	}
	var sqlbuffer strings.Builder
	dname := c.Body.Params["department_name"]
	did := c.Body.Params["department_id"]
	uid := c.Body.Params["user_id"]
	if dname != nil && len(dname.(string)) != 0 {
		sqlbuffer.WriteString(" and department_name='" + dname.(string) + "'")
	}
	if did != nil {
		id, err := util.Interface2Int(did)
		if err != nil {
			return err
		}
		sqlbuffer.WriteString(fmt.Sprintf(" and department_id=%d", id))
	}
	if uid != nil {
		id, err := util.Interface2Int(uid)
		if err != nil {
			return err
		}
		sqlbuffer.WriteString(fmt.Sprintf(" and user_id=%d", id))
	}
	if sqlbuffer.Len() == 0 {
		return errors.New("参数不能全为空或无效")
	}
	result, err := model.FindFznewsLeadership(sqlbuffer.String()[5:])
	if err != nil {
		return err
	}
	c.Body.Data = c.Body.Data[:0]
	c.Body.Data = append(c.Body.Data, result)
	return nil

}
