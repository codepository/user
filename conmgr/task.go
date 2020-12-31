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

// CompleteDescribe 任务人数，提交情况，审结情况,考核组数
func CompleteDescribe(c *model.Container) error {
	if len(c.Body.Params) == 0 || c.Body.Params["titleLike"] == nil {
		return fmt.Errorf("{'body':{'params':'titleLike':''}}titleLike 不能为空")
	}
	titleLike := c.Body.Params["titleLike"].(string)
	// 查询现有考核组
	tags, err := model.FindAllTags("", "tagName like '%考核组成员' and type='考核组'")
	if err != nil {
		return fmt.Errorf("查询考核组:%s", err.Error())
	}
	if len(tags) == 0 {
		return fmt.Errorf("不存在考核组标签,请添加考核组标签如:第二考核组成员")
	}
	var groups []string
	var personnumbers []int
	var applynumbers []int
	var completenumbers []int
	for _, tag := range tags {
		// 考核组
		groups = append(groups, tag.TagName)
		// 查询任务人数
		num, err := model.CountUserByTagID(tag.ID)
		if err != nil {
			return fmt.Errorf("查询标签为[%s]的人数:%s", tag.TagName, err.Error())
		}
		personnumbers = append(personnumbers, num)
		// 提交情况
		applynumber, err := model.CountFlowprocess(fmt.Sprintf("title like '%s' and uid in (SELECT uId from weixin_oauser_taguser where tagId=%d)", "%"+titleLike, tag.ID))
		if err != nil {
			return fmt.Errorf("查询流程提交情况:%s", err.Error())
		}
		applynumbers = append(applynumbers, applynumber)
		// 审结情况
		completenumber, err := model.CountFlowprocess(fmt.Sprintf("title like '%s' and completed=1 and uid in (SELECT uId from weixin_oauser_taguser where tagId=%d)", "%"+titleLike, tag.ID))
		if err != nil {
			return fmt.Errorf("查询流程审批结束情况:%s", err.Error())
		}
		completenumbers = append(completenumbers, completenumber)

	}
	c.Body.Data = append(c.Body.Data, personnumbers)
	c.Body.Data = append(c.Body.Data, applynumbers)
	c.Body.Data = append(c.Body.Data, completenumbers)
	c.Body.Data = append(c.Body.Data, groups)
	c.Body.Data = append(c.Body.Data, []string{"考核组人数", "提交人数", "审结人数", "考核组名称"})
	return nil

}

// PersonApplyYxkh 已经提交一线考核的用户或未提交一线考核的用户
// {"body":"params":{"apply":1,"titleLike":"2020年10月份-月度考核"}} 0表示未提交一线考核的员工，1表示已提交的员工
func PersonApplyYxkh(c *model.Container) error {
	if len(c.Body.Params) == 0 || c.Body.Params["apply"] == nil || c.Body.Params["titleLike"] == nil {
		return fmt.Errorf(`{"body":"params":{"limit":20,"offset":0,"apply":1,"titleLike":"2020年10月份-月度考核"}} 1表示已经提交一线考核的员工，0表示未提交的员工,apply和titleLike都不能为空`)
	}
	// 查询现有考核组
	tags, err := model.FindAllTags("", "tagName like '%考核组成员' and type='考核组'")
	if err != nil {
		return fmt.Errorf("查询考核组:%s", err.Error())
	}
	if len(tags) == 0 {
		return fmt.Errorf("不存在考核组标签,请添加考核组标签如:第二考核组成员")
	}
	apply, err := util.Interface2Int(c.Body.Params["apply"])
	if err != nil {
		return err
	}
	titleLike := c.Body.Params["titleLike"].(string)
	if err != nil {
		return err
	}
	limit := 20
	offset := 0
	if c.Body.Params["limit"] != nil {
		limit, err = util.Interface2Int(c.Body.Params["limit"])
	}
	if c.Body.Params["offset"] != nil {
		offset, err = util.Interface2Int(c.Body.Params["offset"])
	}
	for _, tag := range tags {
		// 考核组
		c.Body.Fields = append(c.Body.Fields, tag.TagName)
		if apply == 1 {
			// 已经提交指定流程的用户
			users, err := model.FindAllUserInfoLimit("id,name,avatar", fmt.Sprintf("id in(SELECT uId from weixin_oauser_taguser where tagId=%d) and id in(SELECT uid from fznews_flow_process where title LIKE '%s')", tag.ID, "%"+titleLike+"%"), limit, offset)
			if err != nil {
				return fmt.Errorf("查询已提交申请:%s", err.Error())
			}
			c.Body.Data = append(c.Body.Data, users)

		} else {
			// 未提交指定流程的用户
			users, err := model.FindAllUserInfoLimit("id,name,avatar", fmt.Sprintf("id in(SELECT uId from weixin_oauser_taguser where tagId=%d) and id not in(SELECT uid from fznews_flow_process where title LIKE '%s')", tag.ID, "%"+titleLike+"%"), limit, offset)
			if err != nil {
				return fmt.Errorf("查询已提交申请:%s", err.Error())
			}
			c.Body.Data = append(c.Body.Data, users)

		}
	}
	return nil

}

// PersonFinishYxkh 一线考核已经审结的用户，一线考核为审结的用户
// {"body":"params":{"finish":1,"titleLike":"2020年10月份-月度考核"}} 1表示已经审批结束的一线考核的员工，0表示未审批结束的一线考核员工
func PersonFinishYxkh(c *model.Container) error {
	if len(c.Body.Params) == 0 || c.Body.Params["finish"] == nil || c.Body.Params["titleLike"] == nil {
		return fmt.Errorf(`{"body":"params":{"finish":1,"titleLike":"2020年10月份-月度考核"}} 1表示已经审批结束的一线考核的员工，0表示未审批结束的一线考核员工,finish和titleLike都不能为空`)
	}
	// 查询现有考核组
	tags, err := model.FindAllTags("", "tagName like '%考核组成员' and type='考核组'")
	if err != nil {
		return fmt.Errorf("查询考核组:%s", err.Error())
	}
	if len(tags) == 0 {
		return fmt.Errorf("不存在考核组标签,请添加考核组标签如:第二考核组成员")
	}
	finish, err := util.Interface2Int(c.Body.Params["finish"])
	if err != nil {
		return err
	}
	titleLike := c.Body.Params["titleLike"].(string)
	if err != nil {
		return err
	}
	limit := 20
	offset := 0
	if c.Body.Params["limit"] != nil {
		limit, err = util.Interface2Int(c.Body.Params["limit"])
	}
	if c.Body.Params["offset"] != nil {
		offset, err = util.Interface2Int(c.Body.Params["offset"])
	}
	for _, tag := range tags {
		// 考核组
		c.Body.Fields = append(c.Body.Fields, tag.TagName)
		if finish == 1 {
			// 已经提交指定流程的用户
			users, err := model.FindAllUserInfoLimit("id,name,avatar", fmt.Sprintf("id in(SELECT uId from weixin_oauser_taguser where tagId=%d) and id in(SELECT uid from fznews_flow_process where title LIKE '%s'  and completed=1)", tag.ID, "%"+titleLike+"%"), limit, offset)
			if err != nil {
				return fmt.Errorf("查询已审批结束:%s", err.Error())
			}
			c.Body.Data = append(c.Body.Data, users)

		} else {
			// 未提交指定流程的用户
			users, err := model.FindAllUserInfoLimit("id,name,avatar", fmt.Sprintf("id in(SELECT uId from weixin_oauser_taguser where tagId=%d) and id  in(SELECT uid from fznews_flow_process where title LIKE '%s' and completed=0)", tag.ID, "%"+titleLike+"%"), limit, offset)
			if err != nil {
				return fmt.Errorf("查询未审批结束:%s", err.Error())
			}
			c.Body.Data = append(c.Body.Data, users)

		}
	}
	return nil
}

// UserTaskRank 剩余任务数和已经完成任务数排行
// {"body":"params":{"businessType":"月度考核","titleLike":"11月份-月度考核"}}；
func UserTaskRank(c *model.Container) error {
	errstr := `{"body":"params":{"businessType":"月度考核","titleLike":"11月份-月度考核","limit":10}}`
	if len(c.Body.Params) == 0 {
		return fmt.Errorf(errstr)
	}
	var sqlbuff strings.Builder
	if c.Body.Params["businessType"] != nil {
		sqlbuff.WriteString(fmt.Sprintf("and businessType='%s'", c.Body.Params["businessType"].(string)))
	}
	if c.Body.Params["titleLike"] != nil {
		sqlbuff.WriteString(fmt.Sprintf("and title like '%s'", "%"+c.Body.Params["titleLike"].(string)+"%"))
	}
	limit := 10
	if c.Body.Params["limit"] != nil {
		l, err := util.Interface2Int(c.Body.Params["limit"])
		if err != nil {
			return err
		}
		limit = l
	}
	if sqlbuff.Len() == 0 {
		return fmt.Errorf(errstr)
	}
	sql := sqlbuff.String()[4:]
	// 查询未完成任务排行
	uncomp, err := model.FindUserTaskRank(limit, fmt.Sprintf("%s and completed=%d", sql, 0))
	if err != nil {
		return fmt.Errorf("用户未完成任务排行:%s", err.Error())
	}
	// 已经完成任务数排行
	comp, err := model.FindUserTaskRank(limit, fmt.Sprintf("%s and completed=%d", sql, 1))
	if err != nil {
		return fmt.Errorf("用户已完成任务排行:%s", err.Error())
	}
	c.Body.Data = append(c.Body.Data, uncomp)
	c.Body.Data = append(c.Body.Data, comp)
	c.Body.Data = append(c.Body.Data, []string{"未完成任务数排行", "已完成任务数排行"})
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
