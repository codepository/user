package conmgr

import (
	"errors"
	"fmt"

	"strings"
	"time"

	"github.com/codepository/user/model"
	"github.com/mumushuiding/util"
)

// StartFlowByToken 启动流程
func StartFlowByToken(c *model.Container) error {
	if len(c.Header.Token) == 0 || len(c.Body.Params) == 0 {
		return errors.New(`参数类型{"header:{"token":""},"body":{"params":{"title":"张三-6月一线考核","templateId":"002d5df2a737dd36a2e78314b07d0bb1_1591669930"}},templateId为模板id`)
	}
	templateID := c.Body.Params["templateId"]
	if templateID == nil || len(templateID.(string)) == 0 {
		return errors.New("templateId 不能为空")
	}
	// 流程名字如:张三-6月一线考核
	title := c.Body.Params["title"]
	if title == nil || len(title.(string)) == 0 {
		return errors.New("流程title不能为空")
	}
	user, err := GetUserByToken(c.Header.Token)
	if err != nil {
		return err
	}
	labels, err := GetLabelNamesByToken(c.Header.Token)
	if err != nil {
		return err
	}
	var khz string
	// 是否包含部门领导这个标签
	var isLeader float64
	for _, l := range labels {
		if strings.HasSuffix(l, "考核组成员") {
			khz = l
		}
		if l == model.DepartmentLeader {
			isLeader = 1
		}
	}
	if len(khz) == 0 {
		return errors.New("不是考核组成员不能填写考核表")
	}

	// c.Body.Data = append(c.Body.Data, params)

	// 查询流程

	w := &model.WeixinTemplates{}
	err = w.FindByTemplateID(templateID.(string))
	if err != nil {
		return fmt.Errorf("查询WeixinTemplates失败:%s", err.Error())
	}
	// 流程数据
	n, err := w.GetTemplateData()
	if err != nil {
		return err
	}
	// 解析流程
	params := make(map[string]interface{})
	// 用户所在考核组
	params["khzKey"] = khz
	// 用户微信ID
	params["userid"] = user.Userid
	// 用户的职级: 中层正职、中层副职等
	params["zjKey"] = user.Level
	// 若是部门领导，那么上级领导是分管领导
	params["isleader"] = isLeader
	// 用户部门ID
	params["departmentKey"] = user.DepartmentID
	result, notify, err := n.Parse(params)
	if err != nil {
		return err
	}
	// 流水号
	thirdNo := w.GenerateThirdNo(user.Userid)
	// 存储执行流
	approvalData := model.Approvaldata{
		Errcode: 0,
		Errmsg:  "ok",
		Data: &model.ExecutionData{
			ThirdNo:        thirdNo,
			OpenTemplateID: w.TemplateID,
			OpenSpName:     w.TemplateName,
			OpenSpstatus:   1,
			ApplyTime:      time.Now().Unix(),
			ApplyUserID:    user.Userid,
			ApplyUsername:  user.Name,
			ApplyUserParty: user.Departmentname,
			ApplyUserImage: user.Avatar,
			ApprovalNodes:  result,
			NotifyNodes:    notify,
			Approverstep:   0,
		},
	}
	apd := &model.WeixinFlowApprovaldata{
		Agentid:    w.Appid,
		ThirdNo:    w.GenerateThirdNo(user.Userid),
		Data:       approvalData.ToString(),
		Step:       0,
		Status:     1,
		NotifyAttr: 1,
		Createtime: time.Now().Nanosecond(),
	}
	// 更新数据库
	tx := model.GetWxTx()
	err = apd.InsertTX(tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("保存流程数据失败,%s", err.Error())
	}
	// 生成流程

	p := &model.WeixinFlowProcess{
		ProcessInstanceID: apd.ThirdNo,
		UserID:            user.Userid,
		Username:          user.Name,
		RequestedDate:     util.FormatDate4(time.Now()),
		Title:             title.(string),
		DeploymentID:      w.TemplateID,
		DeptName:          user.Departmentname,
		Candidate:         user.Userid,
	}
	err = p.InsertTX(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	// 启动流程
	c.Body.Data = append(c.Body.Data, approvalData)
	return nil
}

// FindFlowMyProcess 流程查询
func FindFlowMyProcess(c *model.Container) error {
	if len(c.Header.Token) == 0 {
		return errors.New(`参数格式:{"header":{"token":"aaaaaaa"},"body":{"params":{"title":"张三","completed":1}}} params 可以为空`)
	}
	user, err := GetUserByToken(c.Header.Token)
	if err != nil {
		return err
	}
	title := c.Body.Params["title"]
	completed := c.Body.Params["completed"]
	var sqlbuffer strings.Builder
	sqlbuffer.WriteString("userId='" + user.Userid + "'")
	if title != nil && len(title.(string)) != 0 {
		sqlbuffer.WriteString(" and title like '%" + title.(string) + "%'")
	}
	if completed != nil {
		c, err := util.Interface2Int(completed)
		if err != nil {
			return err
		}
		sqlbuffer.WriteString(fmt.Sprintf(" and completed=%d", c))
	}
	c.Body.Paged = true
	datas, total, err := model.FindAllFlowProcessPaged(c.Body.MaxResults, c.Body.StartIndex, sqlbuffer.String())
	if err != nil {
		return err
	}
	c.Body.Total = total
	c.Body.Data = append(c.Body.Data, datas)
	return nil
}

// FindFlowLog 查询审批纪录
func FindFlowLog(c *model.Container) error {
	// if len(c.Body.Params) == 0 {
	// 	return errors.New(`参数格式:{"body":{"start_index":1,"max_results":15,"params":{"thirdNo":"fdfdf","userId":"linting","username":"张三"}}}`)
	// }
	var sqlbuffer strings.Builder
	thirdNo := c.Body.Params["thirdNo"]
	userID := c.Body.Params["userId"]
	username := c.Body.Params["username"]

	if thirdNo != nil && len(thirdNo.(string)) != 0 {
		sqlbuffer.WriteString(" and thirdNo='" + thirdNo.(string) + "'")
	}
	if userID != nil && len(userID.(string)) != 0 {
		sqlbuffer.WriteString(" and " + model.WeixinFlowLogTable + ".userId='" + userID.(string) + "'")
	}
	if username != nil && len(username.(string)) != 0 {
		sqlbuffer.WriteString(" and " + model.WeixinFlowLogTable + ".username='" + username.(string) + "'")
	}

	c.Body.Paged = true
	var total int
	var logs []*model.WeixinFlowLogResults
	var err error
	if sqlbuffer.Len() != 0 {
		logs, total, err = model.FindAllFlowLogPaged(c.Body.MaxResults, c.Body.StartIndex, sqlbuffer.String()[5:])
	} else {
		logs, total, err = model.FindAllFlowLogPaged(c.Body.MaxResults, c.Body.StartIndex, "")
	}
	if err != nil {
		return err
	}
	c.Body.Total = total
	c.Body.Data = append(c.Body.Data, logs)
	return nil
}

// FindFlowTask 查询我的任务
func FindFlowTask(c *model.Container) error {
	if len(c.Header.Token) == 0 {
		return errors.New(`参数格式:{"header":{"token":"aaaaaaa"},"body":{"params":{"title":"张三","deptName":"开发部,总编室","username":"张三"}}} params 可以为空`)
	}
	var err error
	user, err := GetUserByToken(c.Header.Token)
	if err != nil {
		return err
	}
	title := c.Body.Params["title"]
	deptName := c.Body.Params["deptName"]
	username := c.Body.Params["username"]
	var sqlbuffer strings.Builder
	sqlbuffer.WriteString("FIND_IN_SET('" + user.Userid + "',candidate) and completed=0 ")
	if title != nil && len(title.(string)) != 0 {
		sqlbuffer.WriteString(" and title like '%" + title.(string) + "%'")
	}
	if username != nil && len(username.(string)) != 0 {
		sqlbuffer.WriteString(" and username like '%" + username.(string) + "%'")
	}
	if deptName != nil && len(deptName.(string)) != 0 {
		var deptBuffer strings.Builder
		for _, d := range strings.Split(deptName.(string), ",") {
			deptBuffer.WriteString(",'" + d + "'")
		}
		sqlbuffer.WriteString(" and deptName in (" + deptBuffer.String()[1:] + ")")
	}
	c.Body.Paged = true
	datas, total, err := model.FindAllFlowProcessPaged(c.Body.MaxResults, c.Body.StartIndex, sqlbuffer.String())
	if err != nil {
		return err
	}
	c.Body.Total = total
	c.Body.Data = append(c.Body.Data, datas)
	return nil
}

// CompleteFlowTask 审批
func CompleteFlowTask(c *model.Container) error {
	errstr := `参数格式:{"header":{"token":"aaaaaaaaaaa"},"body":{"params":{"thirdNo":"aaafdfdfd","perform":0}}} perform,2、3、4、5 分别表示:通过、驳回、转审、撤消`
	if len(c.Body.Params) == 0 {
		return errors.New(errstr)
	}

	// 根据 thirdNo 从weixin_leave_approvaldata 获取执行流数据
	thirdNo := c.Body.Params["thirdNo"]
	if c.Body.Params["perform"] == nil {
		return errors.New(errstr)
	}
	perform, err := util.Interface2Int(c.Body.Params["perform"])
	if err != nil {
		return err
	}
	var speech string
	if c.Body.Params["speech"] != nil {
		speech = c.Body.Params["speech"].(string)
	}

	if thirdNo == nil || len(thirdNo.(string)) == 0 {
		return errors.New("thirdNo 不能为空")
	}
	// 判断是否已经审批
	wad, err := model.FindWeixinFlowApprovaldata(thirdNo.(string))
	if err != nil {
		return err
	}
	if wad.Status != 1 {
		return errors.New("当前步骤已经被审批,刷新流程查看详细情况")
	}
	// 查询当前步骤审批人
	executionData, err := wad.GetExecutionData()
	// str, _ := util.ToJSONStr(executionData)
	// log.Println("executionData:", str)
	if err != nil {
		return err
	}
	// 判断用户是否用审批权限
	if len(c.Header.Token) == 0 {
		return errors.New(errstr)
	}
	user, err := GetUserByToken(c.Header.Token)
	if err != nil {
		return err
	}

	// 以下为事物

	// 流转到下一流程或上一个流程
	err = executionData.Perform(perform, user.Userid, speech)
	if err != nil {
		return err
	}
	// 查询流程 WeixinFlowProcess
	p, err := updateProcessByID(executionData)
	if err != nil {
		return err
	}
	// 更新数据库
	tx := model.GetWxTx()
	// 更新执行流数据
	wad.Step = executionData.Approverstep
	// log.Println("step3:", wad.Step)
	err = wad.UpdateDataTX(tx, executionData)
	if err != nil {
		tx.Rollback()
		return err
	}
	// 更新流程
	err = p.UpdateTX(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	// 生成审批纪录
	log := &model.WeixinFlowLog{
		ThirdNo:  thirdNo.(string),
		Status:   perform,
		UserID:   user.Userid,
		Username: user.Name,
		OpTime:   time.Now().Nanosecond(),
		Speech:   speech,
	}
	err = log.InsertTX(tx)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func updateProcessByID(executionData *model.ExecutionData) (*model.WeixinFlowProcess, error) {
	procs, err := model.FindAllFlowProcess("processInstanceId=?", executionData.ThirdNo)
	if err != nil {
		return nil, err
	}
	if len(procs) == 0 {
		return nil, fmt.Errorf("不存在processInstanceId为[%s]的流程", executionData.ThirdNo)
	}
	p := procs[0]
	if executionData.Approverstep > len(executionData.ApprovalNodes) {
		p.Completed = 1
		p.Candidate = ""
	} else if executionData.Approverstep == 0 {
		p.Candidate = executionData.ApplyUserID
	} else {
		var userid strings.Builder
		for _, u := range executionData.ApprovalNodes[executionData.Approverstep-1].Items.Item {
			userid.WriteString("," + u.ItemUserID)
		}
		if userid.Len() == 0 {
			return nil, fmt.Errorf("第[%d]步审批人为空,联系管理员", executionData.Approverstep)
		}
		p.Candidate = userid.String()[1:]
	}
	return p, nil
}
