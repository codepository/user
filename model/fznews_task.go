package model

// FznewsTask 任务
type FznewsTask struct {
	Model
	// 任务名称
	Name string `json:"name,omitempty"`
	// 任务需要执行的方法
	Method string `json:"method,omitemtpy"`
	// 需要完成该任务的用户角色
	Roles string `json:"roles"`
}

// FindAllTask FindAllTask
func FindAllTask(query interface{}, values ...interface{}) ([]*FznewsTask, error) {
	var datas []*FznewsTask
	err := db.Where(query, values...).Find(&datas).Error
	return datas, err
}

// FindTaskRolesByTaskName 根据任务名称查询标签
func FindTaskRolesByTaskName(name string) (string, error) {
	datas, err := FindAllTask("name = ?", name)
	if err != nil {
		return "", nil
	}
	return datas[0].Roles, nil
}
