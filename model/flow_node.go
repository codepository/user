package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mumushuiding/util"
)

// Node 流程数据
type Node struct {
	// 节点类型: route 表示条件节点，condition 表示条件节点，approver 审批节点,notifier 审批节点
	Type   string `json:"type,omitempty"`
	NodeID string `json:"nodeId,omitempty"`
	PrevID string `json:"prevId,omitempty"`
	// ConditionNodes 和 ChildNode 同时出现时先执行 ConditionNodes 再执行 ChildNode
	ConditionNodes []*Node         `json:"conditionNodes,omitempty"`
	ChildNode      *Node           `json:"childNode,omitempty"`
	Properties     *NodeProperties `json:"properties,omitempty"`
}

// NodeProperties 参数
type NodeProperties struct {
	Conditions    []*NodeCondition `json:"conditions,omitempty"`
	ActionerRules []*ActionerRule  `json:"actionerRules,omitempty"`
}

// NodeCondition NodeCondition
type NodeCondition struct {
	// type: 0表示值，1表示范围
	Type int `json:"type"`
	// paramKey：要匹配的值的key
	ParamKey string `json:"paramKey,omitempty"`
	// 参数类型名称
	ParamLabel string `json:"paramLabel,omitempty"`
	IsEmpty    bool   `json:"isEmpty,omitempty"`
	// 类型为range
	LowerBound      string `json:"lowerBound,omitempty"`
	LowerBoundEqual string `json:"lowerBoundEqual,omitempty"`
	UpperBoundEqual string `json:"upperBoundEqual,omitempty"`
	UpperBound      string `json:"upperBound,omitempty"`
	BoundEqual      string `json:"boundEqual,omitempty"`
	Unit            string `json:"unit,omitempty"`
	// 类型为值
	// 当前条件值，可以多个，只要包含其中一个，这个条件就满足
	ParamValues []interface{} `json:"paramValues,omitempty"`
	OriValue    []string      `json:"oriValue,omitempty"`
	Conds       []*NodeCond   `json:"conds,omitempty"`
}

// NodeCond NodeCond
type NodeCond struct {
	Type  string    `json:"type,omitempty"`
	Value string    `json:"value,omitempty"`
	Attrs *Approver `json:"attrs,omitempty"`
}

// ActionerRule 执行规则
type ActionerRule struct {
	// attr：1-或签；2-会签；
	Attr int `json:"attr"`
	// type：3上级，2标签，1单个成员
	Type int `json:"type"`
	// 标签名
	LabelName string `json:"labelName,omitempty"`
	// 用户ID
	UserID string `json:"user_id"`
	Labels int    `json:"labels,omitempty"`
	Level  int8   `json:"level,omitempty"`
}

// ExecutionData 执行流数据
type ExecutionData struct {
	// 审批单号20位
	ThirdNo string `json:"ThirdNo"`
	// 审批模板id
	OpenTemplateID string `json:"OpenTemplateId"`
	// 审批模板名称
	OpenSpName string `json:"OpenSpName"`
	// 申请单当前审批状态：1-审批中；2-已通过；3-已驳回；4-已取消
	OpenSpstatus int `json:"OpenSpstatus"`
	// 申请日期
	ApplyTime int64 `json:"ApplyTime"`
	// 申请人
	ApplyUserID   string `json:"ApplyUserId"`
	ApplyUsername string `json:"ApplyUsername"`
	// 申请人部门
	ApplyUserParty string `json:"ApplyUserParty"`
	// 申请人头像
	ApplyUserImage string `json:"ApplyUserImage"`
	// 审批节点
	ApprovalNodes []*ApprovalNode `json:"ApprovalNodes"`
	// 抄送节点
	NotifyNodes []*ApprovalNode `json:"NotifyNodes"`
	// 当前执行步骤,0是第一个审批节点
	Approverstep int `json:"Approverstep"`
}

// ApprovalNodes 审批节点
type ApprovalNodes struct {
	ApprovalNode []*ApprovalNode `json:"ApprovalNode"`
}

// ApprovalNode 审批子节点
type ApprovalNode struct {
	// 节点审批操作状态：1-审批中；2-已同意；3-已驳回；4-已转审‘
	NodeStatus int
	// 审批节点信息，当节点为标签或上级时，一个节点可能有多个分支
	Items *Items
	// 审批节点属性：1-或签；2-会签
	NodeAttr int
	// 审批节点类型：1-固定成员；2-标签；3-上级
	NodeType int
}

// Items Items
type Items struct {
	Item []*Approver
}

// Approver 审批人
type Approver struct {
	// 审批人
	ItemName string `json:"itemName"`
	// 部门
	ItemParty string
	// 头像
	ItemImage string
	// 用户微信ID
	ItemUserID string `json:"ItemUserId"`
	// 分支审批审批操作状态：1-审批中；2-已同意；3-已驳回；4-已转审
	ItemStatus int
	// 审批意见
	ItemSpeech string
	// 审批日期
	ItemOpTime int
}

// NotifyNodes 抄送节点
type NotifyNodes struct {
	NotifyNode []*Notifier
}

// Notifier 抄送人
type Notifier struct {
	// 审批人
	ItemName string
	// 部门
	ItemParty string
	// 头像
	ItemImage string
	// 用户微信ID
	ItemUserID string `json:"ItemUserId"`
}

// ToString ToString
func (n *Node) ToString() string {
	str, _ := util.ToJSONStr(n)
	return str
}

// ToString ToString
func (an *ApprovalNodes) ToString() string {
	str, _ := util.ToJSONStr(an)
	return str
}

// validate 条件有效性验证
func (c *NodeCondition) validate() error {
	if c.Type == 0 {
		if c.ParamValues == nil || len(c.ParamKey) == 0 || len(c.ParamValues) == 0 {
			return errors.New("ParamKey、ParamValues不能为空")
		}
	}
	return nil
}

// FromString FromString
func (n *Node) FromString(source string) error {
	return util.Str2Struct(source, n)
}

// Parse 解析生成执行流
// 最后生成的一是串人名，因而需要配合用户模块
func (n *Node) Parse(params map[string]interface{}) ([]*ApprovalNode, []*ApprovalNode, error) {
	// 检查参数的有效性
	if err := checkParmas(params); err != nil {
		return nil, nil, err
	}
	var approvalNodes []*ApprovalNode
	var notifierNodes []*ApprovalNode
	// var notifyNodes []*Notifier
	var tempNode []interface{}
	tempNode = append(tempNode, n)
	// 生成审批结点
	for {
		a := len(tempNode)
		if a == 0 {
			break
		}
		// 取出最后一个元素
		current := tempNode[a-1]
		tempNode = tempNode[:a-1]
		cn, yes := current.(*Node)
		if yes {
			// 执行节点操作
			an, err := cn.execute(params["userid"].(string), params["approverid"], params["isleader"].(float64), params["departmentKey"].(int))
			if err != nil {
				return nil, nil, err
			}
			if an != nil {
				if cn.Type == "approver" {
					approvalNodes = append(approvalNodes, an)
				}
				if cn.Type == "notifier" {
					notifierNodes = append(notifierNodes, an)
				}
			}
			// 是否存在子节点
			if cn.ChildNode != nil {
				tempNode = append(tempNode, cn.ChildNode)
			}
			// 是否存在条件节点
			if cn.ConditionNodes != nil && len(cn.ConditionNodes) > 0 {
				tempNode = append(tempNode, cn.ConditionNodes)
			}
		} else {
			cond := current.([]*Node)
			for _, v := range cond {
				yes, err := v.check(params)
				if err != nil {
					return nil, nil, err
				}
				if yes {
					if v.ChildNode != nil {
						tempNode = append(tempNode, v.ChildNode)
					}
					break
				}
			}
		}

	}
	return approvalNodes, notifierNodes, nil
}

// 执行节点操作，只生成审批人和抄送人
func (n *Node) execute(userID string, approverid interface{}, isleader float64, departmentid int) (*ApprovalNode, error) {
	if len(n.Type) == 0 {
		return nil, fmt.Errorf("nodeId为'%s'的节点,type不能为空type: start 开始,approver 审批人,end 结束,notifier 抄送人", n.NodeID)
	}

	if n.Type != "approver" && n.Type != "notifier" {
		return nil, nil
	}
	// 检查执行规则
	if n.Properties == nil || n.Properties.ActionerRules == nil || len(n.Properties.ActionerRules) == 0 {
		return nil, fmt.Errorf("nodeId为'%s'的节点,properties.actionerRules不能为空", n.NodeID)
	}

	a := n.Properties.ActionerRules[0]
	result := &ApprovalNode{
		NodeAttr:   a.Attr,
		NodeStatus: 1,
		NodeType:   a.Type,
		Items: &Items{
			Item: []*Approver{},
		},
	}

	if a.Attr == 1 {
		if a.Type == 0 {
			return nil, fmt.Errorf("nodeId为'%s'的节点,actionerRule中type不能为空", n.NodeID)
		}
		switch a.Type {
		case 3:
			var users []*Userinfo
			var err error
			// 查询审批领导
			users, err = FindLeaderByDepartmentID(departmentid, isleader)
			if err != nil {
				return nil, err
			}
			for _, u := range users {
				result.Items.Item = append(result.Items.Item, &Approver{
					ItemImage:  u.Avatar,
					ItemParty:  u.Departmentname,
					ItemName:   u.Name,
					ItemUserID: u.Userid,
				})

			}
			break
		case 2:
			if len(a.LabelName) == 0 {
				return nil, fmt.Errorf("nodeId为'%s'的节点,当前actionerRule中LabelName不能为空", n.NodeID)
			}
			// 查询标签对应的用户
			users, err := FindUsersByLabelNames([]string{a.LabelName})
			if err != nil {
				return nil, err
			}
			if len(users) == 0 {
				return nil, fmt.Errorf("不存在标签为【%s】的用户，联系管理员", a.LabelName)
			}
			var app []*Approver
			for _, u := range users {
				app = append(app, &Approver{
					ItemImage:  u.Avatar,
					ItemParty:  u.Departmentname,
					ItemName:   u.Name,
					ItemUserID: u.Userid,
				})
			}
			result.Items.Item = app
			break
		case 1:
			// 根据用户userID查询用户信息
			if approverid == nil {
				return nil, errors.New("参数中approverid不能为空与微信的用户id对应")
			}
			id, yes := approverid.(string)
			if !yes || len(id) == 0 {
				return nil, errors.New("参数中approverid必须为有效字符与微信的用户id对应")
			}
			u, err := FindUserinfoByUserID(id)
			if err != nil {
				return nil, err
			}
			result.Items.Item = append(result.Items.Item, &Approver{
				ItemImage:  u.Avatar,
				ItemParty:  u.Departmentname,
				ItemName:   u.Name,
				ItemUserID: u.Userid,
			})
			break
		default:
			return nil, fmt.Errorf("nodeId为'%s'的节点,actionerRule中type只能为：3上级，2标签，1单个成员", n.NodeID)

		}
	}

	return result, nil
}

// 检查条件
func (n *Node) check(params map[string]interface{}) (bool, error) {
	if n.Properties == nil || n.Properties.Conditions == nil || len(n.Properties.Conditions) == 0 {
		return false, fmt.Errorf("nodeId为:%s的Properties.Conditions值不能为空", n.NodeID)
	}
	flag := 0
	for _, c := range n.Properties.Conditions {
		err := c.validate()
		if err != nil {
			return false, fmt.Errorf("节点%s, properties.conditions值有误:%s", n.NodeID, err.Error())
		}
		if c.Type == 0 {
			for _, v := range c.ParamValues {
				// log.Println("v:", reflect.TypeOf(v))
				// log.Println("key:", reflect.TypeOf(params[c.ParamKey]))
				// log.Printf("匹配条件:v:%vkey:%s,keyval:%v.\n", v, c.ParamKey, params[c.ParamKey])
				vstr, _ := util.ToJSONStr(v)
				kstr, _ := util.ToJSONStr(params[c.ParamKey])

				if vstr == kstr {
					// log.Printf("匹配成功:v:%v,key:%s,keyval:%v.\n", v, c.ParamKey, params[c.ParamKey])
					flag++
					break
				}
			}
		}
	}
	return flag == len(n.Properties.Conditions), nil
}

// checkParmas 验证参数的有效性
func checkParmas(params map[string]interface{}) error {
	if params == nil {
		return errors.New("参数不能为空")
	}
	if params["khzKey"] == nil || params["zjKey"] == nil || params["userid"] == nil {
		return errors.New("khzKey、zjKey、userid不能为空")
	}
	khz, yes := params["khzKey"].(string)

	if !yes {
		return errors.New(`参数类型必须是:{"body":{"data":[{"khzKey":"","zjKey":1,"userid":""}]}},khzKey表示用户所有考核组，zjKey表示用户的职级`)
	}
	if len(khz) == 0 {
		return errors.New("khzKey参数不能为空字符串")
	}
	_, yes1 := params["zjKey"].(float64)
	_, yes2 := params["zjKey"].(int)
	if !yes1 && !yes2 {
		return errors.New(`zjKey必须是数字，0是一般工作人员， 1中层正职（含主持工作的副职）, 2是中层副职，3是社领导`)
	}
	userid, yes := params["userid"].(string)

	if !yes {
		return errors.New(`参数类型必须是:{"body":{"data":[{"khzKey":"","zjKey":"","userid":""}]}},khzKey表示用户所有考核组，zjKey表示用户的职级`)
	}
	if len(userid) == 0 {
		return errors.New("userid参数不能为空字符串")
	}
	return nil
}

// GetProcessConfigFromJSONFile test
func (n *Node) GetProcessConfigFromJSONFile() {
	file, err := os.Open("D:/Workspaces/go/src/github.com/codepository/user/process/一线考核流程.json")
	if err != nil {
		log.Printf("cannot open file processConfig.json:%v", err)
		panic(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(n)
	if err != nil {
		log.Printf("decode processConfig.json failed:%v", err)
	}
}

//   ================================ 执行流 ========================

// ToString ToString
func (e *ExecutionData) ToString() (string, error) {
	return util.ToJSONStr(e)
}
func (e *ExecutionData) execute(perform int, userid string) error {
	// str, _ := e.ToString()
	// log.Println("ExecutionData.execute:", str)
	if e.OpenSpstatus != 1 {
		return errors.New("非审批状态无法审批")
	}
	if userid != e.ApplyUserID {
		return errors.New("只允许本人操作")
	}
	// 执行流程
	if perform != 2 {
		return errors.New("当前步骤只允许通过")
	}
	e.OpenSpstatus = perform
	return nil
}
func (a *ApprovalNode) execute(perform int, userid, speech string) error {
	// NodeStatus是否是处于审批中

	if a.NodeStatus != 1 {
		return errors.New("已审批,刷新查看详情")
	}
	canIPerform := false
	var us []string
	for _, item := range a.Items.Item {
		// 或签且已经有人审批了
		if item.ItemStatus != 1 && a.NodeAttr == 1 {
			return errors.New("已审批,刷新查看详情")
		}
		if item.ItemUserID == userid {
			canIPerform = true
			// 执行流程
			item.ItemSpeech = speech
			item.ItemOpTime = time.Now().Nanosecond()
			if a.NodeAttr == 1 {
				item.ItemStatus = perform
				a.NodeStatus = perform
			} else {
				return errors.New("暂不支持会签")
			}
			break
		} else {
			us = append(us, item.ItemName)
		}
	}
	if !canIPerform {
		if len(us) == 0 {
			return errors.New("当前步骤没有审批人,请联系管理员")
		}
		return errors.New("你没有审批权限,审批人只能是:" + strings.Join(us, ","))
	}
	return nil
}

// complete 通过或驳回
func (e *ExecutionData) complete(perform int, userid string, speech string) error {
	if e.Approverstep == 0 {
		err := e.execute(perform, userid)
		if err != nil {
			return err
		}
	} else if e.Approverstep > len(e.ApprovalNodes) {
		return errors.New("流程已经审批结束,无法执行")
	} else {

		approvalNode := e.ApprovalNodes[e.Approverstep-1]
		err := approvalNode.execute(perform, userid, speech)
		if err != nil {
			return err
		}

	}
	switch perform {
	case 2:
		e.Approverstep = e.Approverstep + 1

		break
	case 3:
		e.Approverstep = e.Approverstep - 1
		break
	default:
		return errors.New("只支持“通过”、“驳回”")
	}

	if e.Approverstep == 0 {
		e.OpenSpstatus = 1
	} else if len(e.ApprovalNodes) >= e.Approverstep {
		e.ApprovalNodes[e.Approverstep-1].NodeStatus = 1
		for _, item := range e.ApprovalNodes[e.Approverstep-1].Items.Item {
			item.ItemStatus = 1
			item.ItemSpeech = ""
		}

	}
	return nil
}

// withdraw 撤回
func (e *ExecutionData) withdraw(perform int, userid string, speech string) error {
	// 当前步骤的前一步是由撤消人审批的情况下才能撤回
	if e.Approverstep == 0 {
		return errors.New("处于第一步,无法撤回")
	}
	canWithdraw := false
	if e.Approverstep == 1 {
		if e.ApplyUserID == userid {
			canWithdraw = true
			e.OpenSpstatus = 1
		}
	} else if e.Approverstep > len(e.ApprovalNodes) {
		return errors.New("流程已经结束无法撤消")
	} else {
		for _, u := range e.ApprovalNodes[e.Approverstep-2].Items.Item {
			if u.ItemUserID == userid {
				canWithdraw = true
				// 修改上一步骤为正在审批状态
				e.ApprovalNodes[e.Approverstep-2].NodeStatus = 1
				u.ItemSpeech = ""
				u.ItemStatus = 1
				break
			}
		}
	}
	if !canWithdraw {
		return errors.New("只能撤回相邻步骤的流程")
	}
	for _, u := range e.ApprovalNodes[e.Approverstep-1].Items.Item {
		e.ApprovalNodes[e.Approverstep-1].NodeStatus = 0
		u.ItemStatus = 0
	}
	e.Approverstep--
	return nil

}

// Perform 执行流程
// perform 值包含2、3、4、5 分别表示:通过、驳回、转审、撤消
// userid 表示用户
func (e *ExecutionData) Perform(perform int, userid string, speech string) error {

	// log.Println("开始step=", step)
	if perform == 2 || perform == 3 { // 通过或驳回
		err := e.complete(perform, userid, speech)
		if err != nil {
			return err
		}
	} else if perform == 5 { // 撤消
		err := e.withdraw(perform, userid, speech)
		if err != nil {
			return err
		}
	}

	return nil

}
