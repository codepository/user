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
	if c.Body.Params["businessType"] == nil || len(c.Body.Params["businessType"].(string)) == 0 {
		return errors.New("流程类型不能为空")
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
	// // 是否包含部门领导这个标签
	// var isLeader float64
	for _, l := range labels {
		if strings.HasSuffix(l, "考核组成员") {
			khz = l
		}
		// if l == model.DepartmentLeader {
		// 	isLeader = 1
		// }
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
	params["isleader"] = float64(user.IsLeader)
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
		ThirdNo:    thirdNo,
		Data:       approvalData.ToString(),
		Step:       0,
		Status:     1,
		NotifyAttr: 1,
	}
	// 更新数据库
	tx := model.GetWxTx()
	err = apd.InsertTX(tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("保存流程数据失败,%s", err.Error())
	}
	// 生成流程

	p := &model.FznewsFlowProcess{
		ProcessInstanceID: apd.ThirdNo,
		UID:               user.ID,
		UserID:            user.Userid,
		Username:          user.Name,
		RequestedDate:     util.FormatDate4(time.Now()),
		Title:             title.(string),
		DeploymentID:      w.TemplateID,
		DeptName:          user.Departmentname,
		Candidate:         user.Userid,
		BusinessType:      c.Body.Params["businessType"].(string),
	}
	err = p.InsertTX(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	// 添加数据
	c.Body.Data = append(c.Body.Data, approvalData)
	c.Body.Data = append(c.Body.Data, p)
	return nil
}

// DeleteFlow 删除流程
func DeleteFlow(c *model.Container) error {
	if len(c.Body.Params) == 0 {
		return errors.New(`参数格式:{"body":{"params":{"ThirdNo":""}}} ThirdNo 为流程ID`)
	}
	if c.Body.Params["ThirdNo"] == nil || len(c.Body.Params["ThirdNo"].(string)) == 0 {
		return errors.New(`参数格式:{"body":{"params":{"ThirdNo":""}}} ThirdNo 为流程ID`)
	}
	// 判断是否可以直接删除

	return model.DeleteFlowByID(c.Body.Params["ThirdNo"].(string))
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
	datas, total, err := model.FindAllFlowProcessPaged(c.Body.MaxResults, c.Body.StartIndex, "", sqlbuffer.String())
	if err != nil {
		return err
	}
	c.Body.Total = total
	c.Body.Data = append(c.Body.Data, datas)
	return nil
}

// FindFlowLog 查询审批纪录
func FindFlowLog(c *model.Container) error {
	if len(c.Body.Params) == 0 {
		return errors.New(`参数格式:{"body":{"start_index":1,"max_results":15,"params":{"thirdNo":"fdfdf","userId":"linting","username":"张三"}}}参数不能全为空`)
	}
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
		return errors.New(`参数格式:{"body":{"start_index":1,"max_results":15,"params":{"thirdNo":"fdfdf","userId":"linting","username":"张三"}}}参数不能全为空`)
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
	datas, total, err := model.FindAllFlowProcessPaged(c.Body.MaxResults, c.Body.StartIndex, "", sqlbuffer.String())
	if err != nil {
		return err
	}
	c.Body.Total = total
	c.Body.Data = append(c.Body.Data, datas)
	return nil
}

// CompleteFlowTask 审批
func CompleteFlowTask(c *model.Container) error {
	errstr := `参数格式:{"header":{"token":"aaaaaaaaaaa"},"body":{"params":{"thirdNo":"aaafdfdfd","perform":0,"speech":""}}} perform,2、3、4、5 分别表示:通过、驳回、转审、撤消`
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
	// 查询流程 FznewsFlowProcess
	p, err := updateProcessByID(executionData)
	if err != nil {
		return err
	}
	// 更新数据库
	tx := model.GetWxTx()
	// 更新执行流数据
	wad.Step = executionData.Approverstep
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
		OpTime:   time.Now().Unix(),
		Speech:   speech,
	}
	err = log.InsertTX(tx)
	if err != nil {
		return err
	}

	tx.Commit()
	// 返回数据
	c.Body.Data = c.Body.Data[:0]
	// 添加流程数据
	c.Body.Data = append(c.Body.Data, p)
	// 添加流程纪录
	c.Body.Data = append(c.Body.Data, log)
	return nil
}

func updateProcessByID(executionData *model.ExecutionData) (*model.FznewsFlowProcess, error) {
	procs, err := model.FindAllFlowProcess("", fmt.Sprintf("processInstanceId='%s'", executionData.ThirdNo))
	if err != nil {
		return nil, err
	}
	if len(procs) == 0 {
		return nil, fmt.Errorf("不存在processInstanceId为[%s]的流程", executionData.ThirdNo)
	}
	p := procs[0]
	p.Step = executionData.Approverstep
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

// FlowStepper 查询流程进度
func FlowStepper(c *model.Container) error {
	errstr := `参数格式：{"body":{"params":{"processInstanceId":""}}} processInstanceId 为流程ID不能为空`
	if len(c.Body.Params) == 0 || c.Body.Params["processInstanceId"] == nil {
		return fmt.Errorf(errstr)
	}
	// 查询
	wad, err := model.FindWeixinFlowApprovaldata(c.Body.Params["processInstanceId"].(string))
	if err != nil {
		return err
	}
	d, err := wad.GetExecutionData()
	if err != nil {
		return err
	}
	c.Body.Data = append(c.Body.Data, d)
	return nil
}

// FindAllFlow 查询所有流程
func FindAllFlow(c *model.Container) error {

	// errstr := `参数格式: {"body":{"params":{"processInstanceId":"xxxxx","userId":"wanyu","uid":33,"titleLike":"","title":"sss",
	// "businessType":"dxxx","deptName":"","candidate":"linting","username":"张三","completed":0,"posts":[0,1]}}}
	// userId 流程申请人对应的微信号, uid 申请人的ID, title 流程名,titleLike 流程名近似查询, businessType 流程类型
	// deptName 部门名称, candidate 审批人对应微信号, username 申请人姓名, posts对应申请人的职级：0是一般工作人员，
	// 1中层正职（含主持工作的副职）, 2是中层副职，3是社领导
	// `
	// if len(c.Body.Params) == 0 {
	// 	return fmt.Errorf(errstr)
	// }
	var buff strings.Builder
	if !util.InterfaceIsEmpty(c.Body.Params["processInstanceId"]) {
		buff.WriteString(fmt.Sprintf(" and processInstanceId='%v'", c.Body.Params["processInstanceId"]))
	}
	if !util.InterfaceIsEmpty(c.Body.Params["userId"]) {
		buff.WriteString(fmt.Sprintf(" and userId='%v'", c.Body.Params["userId"]))
	}
	if !util.InterfaceIsEmpty(c.Body.Params["uid"]) {
		buff.WriteString(fmt.Sprintf(" and uid=%v", c.Body.Params["uid"]))
	}
	if !util.InterfaceIsEmpty(c.Body.Params["title"]) {
		buff.WriteString(fmt.Sprintf(" and title='%v'", c.Body.Params["title"]))
	}
	if !util.InterfaceIsEmpty(c.Body.Params["titleLike"]) {
		titleLike, ok := c.Body.Params["titleLike"].(string)
		if !ok {
			return fmt.Errorf("titleLike 必须是字符串")
		}
		buff.WriteString(fmt.Sprintf(" and title like '%s'", "%"+titleLike+"%"))
	}
	if !util.InterfaceIsEmpty(c.Body.Params["businessType"]) {
		buff.WriteString(fmt.Sprintf(" and businessType='%v'", c.Body.Params["businessType"]))
	}
	if !util.InterfaceIsEmpty(c.Body.Params["deptName"]) {
		dept, ok := c.Body.Params["deptName"].(string)
		if !ok {
			return errors.New("deptName 参数必须是使用逗号分隔的字符串")
		}
		var deptBuffer strings.Builder
		for _, d := range strings.Split(dept, ",") {

			deptBuffer.WriteString(fmt.Sprintf(",'%s'", d))
		}
		buff.WriteString(fmt.Sprintf(" and deptName in (%v)", deptBuffer.String()[1:]))
	}
	if !util.InterfaceIsEmpty(c.Body.Params["candidate"]) {
		buff.WriteString(fmt.Sprintf(" and FIND_IN_SET('%v',candidate)", c.Body.Params["candidate"]))
	}
	if !util.InterfaceIsEmpty(c.Body.Params["username"]) {
		username, ok := c.Body.Params["username"].(string)
		if !ok {
			return fmt.Errorf("username 必须是字符串")
		}
		buff.WriteString(fmt.Sprintf(" and username like '%s'", "%"+username+"%"))
	}
	if c.Body.Params["completed"] != nil {
		c, err := util.Interface2Int(c.Body.Params["completed"])
		if err != nil {
			return fmt.Errorf("completed:%s", err.Error())
		}
		buff.WriteString(fmt.Sprintf(" and completed=%d", c))
	}
	if c.Body.Params["posts"] != nil {
		levels, ok := c.Body.Params["posts"].([]interface{})
		if !ok {
			return fmt.Errorf("posts 格式 [0,1,2] 0是一般工作人员,1中层正职（含主持工作的副职）, 2是中层副职, 3是社领导")
		}
		var lbuff strings.Builder
		for _, l := range levels {
			lbuff.WriteString(fmt.Sprintf(",%v", l))
		}
		if lbuff.Len() > 0 {
			buff.WriteString(fmt.Sprintf(" and uid in (select  id from %s where level in (%s))", model.UserinfoTabel, lbuff.String()[1:]))
		}
	}
	var fields string
	if c.Body.Params["fields"] != nil {
		fields = c.Body.Params["fields"].(string)
	}
	if buff.Len() != 0 {
		datas, _, err := model.FindAllFlowProcessPaged(0, 0, fields, buff.String()[4:])
		if err != nil {
			return err
		}
		c.Body.Data = append(c.Body.Data, datas)
	} else {
		datas, _, err := model.FindAllFlowProcessPaged(0, 0, fields, map[string]interface{}{})
		if err != nil {
			return err
		}
		c.Body.Data = append(c.Body.Data, datas)
	}
	return nil

}

// FlowExist 查询流程是否已经存在
func FlowExist(c *model.Container) error {
	err := FindAllFlow(c)
	if err != nil {
		return err
	}
	ok := false
	if c.Body.Data[0] != nil {
		x := c.Body.Data[0].([]*model.FznewsFlowProcess)
		if len(x) > 0 {
			ok = true
		}
	}
	c.Body.Data = c.Body.Data[:0]
	c.Body.Data = append(c.Body.Data, ok)
	return nil
}
