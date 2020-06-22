package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

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
	Approverstep int `json:"approverstep"`
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
	ItemName string
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
			// 查询用户的上级，若是普通员工他的上级即是部门领导，
			var users []*Userinfo
			var err error
			if isleader == 1 {
				// 若是部门领导他的上级是该部门的分管领导
				users, err = FindLeaderByUserID(userID, isleader)
				if err != nil {
					return nil, err
				}
			} else {
				// 若是普通员工且ActionerRules[0].Level==1则他的上级即是部门领导
				// 若是普通员工且ActionerRules[0].Level==2，则他的上级则是上级部门领导
				// 以此类推
				switch n.Properties.ActionerRules[0].Level {
				case 1:
					users, err = FindLeaderByUserID(userID, isleader)
					if err != nil {
						return nil, err
					}
					break
				case 2:

					users, err = FindSecondLeaderByDepartmentid(departmentid)
					if err != nil {
						return nil, err
					}
					break
				default:

				}
			}

			if len(users) == 0 {
				return nil, fmt.Errorf("部门[%d]未设置领导,无法提交", departmentid)
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
				log.Printf("匹配条件:v:%vkey:%s,keyval:%v.\n", v, c.ParamKey, params[c.ParamKey])
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
