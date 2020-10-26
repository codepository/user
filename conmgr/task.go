package conmgr

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codepository/user/model"
	"github.com/mumushuiding/util"
)

// RouteFunction 根据路径指向方法
type RouteFunction func(*model.Container) error

// RouteHandler 路由
type RouteHandler struct {
	handler RouteFunction
	route   string
}

var routers = []*RouteHandler{
	{route: "exec/task/yxkh", handler: NewYxkhTask},
}

// GenerateTask 生成任务
func GenerateTask() {
	// 获取任务列表
	tasks, err := model.FindAllTask(map[string]interface{}{}, nil)
	if err != nil {
		return
	}
	// 循环生成任务
	for _, task := range tasks {
		par := model.Container{}
		par.Body.Metrics = task.Roles
		par.Body.UserName = task.Name
		var f *RouteHandler
		for _, r := range routers {
			if r.route == task.Method {
				f = r
				break
			}
		}
		if f == nil {
			SendRecord("生成任务", task.Name, 0, errors.New(task.Method+"不存在"))
			continue
		}
		err := f.handler(&par)
		if err != nil {
			SendRecord("生成任务", task.Name, 0, err)
		}

	}

}

// NewYxkhTask 生成一线考核任务
func NewYxkhTask(c *model.Container) error {
	// 今日若为1号则继续
	now := time.Now()
	if now.Day() != 1 {
		return nil
	}
	// 生成当月任务基本参数
	start := util.FormatDate3(now)
	end := util.FormatDate3(now.AddDate(0, 0, 15))
	// 获取需要完成任务的用户
	for _, role := range strings.Split(c.Body.Metrics, ",") {
		users, err := model.FindUsersByLabelNames([]string{role})
		if err != nil {
			return err
		}
		// 纪录
		for _, user := range users {
			var t model.FznewsTaskUser
			t.Userid = user.ID
			t.Username = user.Name
			t.Start = start
			t.Role = role
			t.End = end
			t.Task = c.Body.UserName
			err := t.Save()
			if err != nil && !strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
				SendRecord("保存用户任务", t.ToString(), 0, err)
				continue
			}
		}
	}
	return nil
}

// FindTaskRoles 查询指定任务对应的角色组
func FindTaskRoles(c *model.Container) error {
	if len(c.Body.Metrics) == 0 {
		return errors.New(`查询任务角色参数格式:{"body":{"metrics":"一线考核"}},metrics为任务名称`)
	}
	rolesstr, err := model.FindTaskRolesByTaskName(c.Body.Metrics)
	if err != nil {
		return err
	}
	if len(rolesstr) != 0 {
		c.Body.Data = append(c.Body.Data, strings.Split(rolesstr, ","))
	}
	return nil
}

// CompleteTask 完成任务
func CompleteTask(c *model.Container) error {
	// 参数检查
	if len(c.Body.Params) == 0 {
		return errors.New(`完成任务参数格式:{"body":{"data":[{"task":"一线考核","start":"2020-06-01","userid": 1}]}},task为任务,start为任务开始日期`)
	}
	// log.Println("data[0]:", c.Body.Data[0])
	par := c.Body.Params
	if len(par) == 0 {
		return errors.New(`完成任务参数格式:{"body":{"data":[{"task":"一线考核","start":"2020-06-01","userid": 1}]}},task为任务,start为任务开始日期`)
	}
	if par["task"] == nil || par["start"] == nil || par["userid"] == nil {
		return fmt.Errorf("完成任务参数：task 、 start 和 userid 都不能为空")
	}
	id, err := util.Interface2Int(par["userid"])
	if err != nil {
		return fmt.Errorf("完成任务报错:%s", err.Error())
	}
	err = model.CompleteTask(par["task"].(string), par["start"].(string), id)
	if err != nil {
		return fmt.Errorf("完成任务报错:%s", err.Error())
	}
	return nil
}

// TaskUncomplete 未完成任务的用户
func TaskUncomplete(c *model.Container) error {
	// 参数检查
	if len(c.Body.Params) == 0 {
		return errors.New(`查询任务完成情况参数格式:{"body":{"data":[{"task":"一线考核","start":"2020-06-01"}]}},task为任务,start为任务开始日期`)
	}
	// log.Println("data[0]:", c.Body.Data[0])
	par := c.Body.Params
	if len(par) == 0 {
		return errors.New(`查询任务完成情况参数格式:{"body":{"data":[{"task":"一线考核","start":"2020-06-01"}]}},task为任务,start为任务开始日期`)
	}
	if par["task"] == nil || par["start"] == nil {
		return fmt.Errorf("查询任务完成情况参数：task 和 start 都不能为空")
	}
	if c.Body.MaxResults == 0 {
		c.Body.MaxResults = 30
	}
	users, err := model.FindUsersUnCompleteTask(c.Body.MaxResults, c.Body.StartIndex, par)
	if err != nil {
		return fmt.Errorf("查询任务 [%s]未完成任务的用户: %s ", par["task"].(string), err.Error())
	}
	c.Body.Data = c.Body.Data[:0]
	c.Body.Data = append(c.Body.Data, users)
	// 显示名、紧急程度
	x := []string{"月度考核未交清单"}
	now := time.Now()
	day := now.Day()
	if day > 15 {
		x = append(x, "dead")
	} else if day > 10 {
		x = append(x, "danger")
	} else if day > 8 {
		x = append(x, "warn")
	} else {
		x = append(x, "normal")
	}
	c.Body.Data = append(c.Body.Data, x)
	return nil
}

// TaskCompleteRate 任务完成情况
func TaskCompleteRate(c *model.Container) error {
	// 参数检查
	if len(c.Body.Params) == 0 {
		return errors.New(`查询任务完成情况参数格式:{"body":{"data":[{"task":"一线考核","start":"2020-06-01"}]}},task为任务,start为任务开始日期`)
	}
	// log.Println("data[0]:", c.Body.Data[0])
	par := c.Body.Params
	if len(par) == 0 {
		return errors.New(`查询任务完成情况参数格式:{"body":{"data":[{"task":"一线考核","start":"2020-06-01"}]}},task为任务,start为任务开始日期`)
	}
	if par["task"] == nil || par["start"] == nil {
		return fmt.Errorf("查询任务完成情况参数：task 和 start 都不能为空")
	}
	taskName := par["task"].(string)
	// start := par["start"].(string)
	// 根据任务名称查询所有的角色
	rolesstr, err := model.FindTaskRolesByTaskName(taskName)
	if len(rolesstr) == 0 {
		return fmt.Errorf("任务[%s]对应的role为空，请完善！", taskName)
	}
	if err != nil {
		return fmt.Errorf("查询任务 [%s] 对应标签错误: %s ", taskName, err.Error())
	}
	roles := strings.Split(rolesstr, ",")
	// 根据任务名称、开始时间、角色来统计，统计对象包含完成率
	var rates []float64
	for _, role := range roles {
		par["role"] = role
		rate, err := model.FindTaskCompleteRate(par)
		if err != nil {
			return fmt.Errorf("查询任务 [%s] 对应标签错误: %s ", taskName, err.Error())
		}
		rates = append(rates, rate)
	}
	// 并根据role进行分组
	c.Body.Data = c.Body.Data[:0]
	c.Body.Data = append(c.Body.Data, roles)
	c.Body.Data = append(c.Body.Data, rates)
	return nil
}
