package dao

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type CommonModel struct {
	CreateTime time.Time `json:"createTime,omitempty" gorm:"column:create_time;type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateTime time.Time `json:"updateTime,omitempty" gorm:"column:update_time;type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
	DelTime    time.Time `json:"delTime,omitempty" gorm:"column:del_time;type:datetime;default:'9999-12-31 00:00:00';comment:逻辑删除时间"`
	UpdateBy   string    `json:"updateBy,omitempty" gorm:"column:update_by;type:varchar(36);comment:更新人"`
	CreateBy   string    `json:"createBy,omitempty" gorm:"column:create_by;type:varchar(36);comment:创建人"`
	DelFlag    int       `json:"delFlag,omitempty" gorm:"column:del_flag;type:int;size:32;default:0;comment:逻辑删除标志【0 ： 未删除 1： 已删除】"`
}

type TaskMetaData struct {
	CommonModel
	ID           uint      `json:"id" gorm:"column:id;primaryKey;autoIncrement;comment:任务元数据ID"`
	ExecuteAt    time.Time `json:"executeAt" gorm:"column:execute_at;type:datetime;not null;index:idx_execute_at;comment:任务执行时间点"`
	Name         string    `json:"name" gorm:"column:name;type:varchar(255);not null;index:idx_name;comment:任务名称"`
	DataType     int       `json:"dataType" gorm:"column:data_type;type:int;not null;index:idx_data_type;comment:数据类型"`
	LedgerPath   string    `json:"ledgerPath" gorm:"column:ledger_path;type:varchar(512);not null;comment:账本存储路径"`
	MailReceiver []string  `json:"mailReceiver" gorm:"column:mail_receiver;type:json;comment:邮件接收者列表"`
	MailType     int       `json:"mailType" gorm:"column:mail_type;type:int;comment:邮件类型"`
	MailHeader   string    `json:"mailHeader" gorm:"column:mail_header;type:varchar(255);comment:邮件标题"`
}

func (*TaskMetaData) TableName() string {
	return "tasks"
}

func init() {
	registerInjector(func(d *daoInit) {
		setupTableModel(d, &TaskMetaData{})
	})
}

type TaskDao struct {
	DB *gorm.DB
}

func NewTokenDao(db *gorm.DB) ITaskDao {
	if db == nil {
		db = GetDB()
	}
	return &TaskDao{DB: db}
}

// ITaskDao 接口扩展
type ITaskDao interface {
	// 获取任务列表（分页）
	GetTaskList(page, pageSize int) ([]TaskMetaData, int64, error)

	// 根据名称查找任务（支持模糊查询）
	GetTasksByName(name string, page, pageSize int) ([]TaskMetaData, int64, error)

	// 根据ID获取任务详情
	GetTaskByID(id uint) (*TaskMetaData, error)

	// 创建任务
	CreateTask(task *TaskMetaData) error

	// 更新任务
	UpdateTask(task *TaskMetaData) error

	// 删除任务（逻辑删除）
	DeleteTask(id uint, updateBy string) error
}

var _ ITaskDao = (*TaskDao)(nil)

// GetTaskList 获取任务列表（分页）

func (dao *TaskDao) GetTaskList(page, pageSize int) ([]TaskMetaData, int64, error) {
	var tasks []TaskMetaData
	var total int64

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询未删除的任务
	query := dao.DB.Model(&TaskMetaData{}).
		Where("del_flag = ?", 0).
		Order("execute_at DESC")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取任务总数失败: %w", err)
	}

	// 分页查询
	if err := query.Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("查询任务列表失败: %w", err)
	}

	return tasks, total, nil
}

// GetTasksByName 根据名称查找任务（支持模糊查询）
func (dao *TaskDao) GetTasksByName(name string, page, pageSize int) ([]TaskMetaData, int64, error) {
	var tasks []TaskMetaData
	var total int64

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 构建查询条件
	query := dao.DB.Model(&TaskMetaData{}).
		Where("del_flag = ?", 0).
		Where("name LIKE ?", "%"+name+"%").
		Order("execute_at DESC")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取任务总数失败: %w", err)
	}

	// 分页查询
	if err := query.Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("查询任务列表失败: %w", err)
	}

	return tasks, total, nil
}

// GetTaskByID 根据ID获取任务详情
func (dao *TaskDao) GetTaskByID(id uint) (*TaskMetaData, error) {
	var task TaskMetaData
	result := dao.DB.Where("id = ? AND del_flag = ?", id, 0).First(&task)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("任务不存在")
	}

	if result.Error != nil {
		return nil, fmt.Errorf("查询任务失败: %w", result.Error)
	}

	return &task, nil
}

// CreateTask 创建任务
func (dao *TaskDao) CreateTask(task *TaskMetaData) error {
	// 设置创建时间和更新时间
	now := time.Now()
	task.CreateTime = now
	task.UpdateTime = now
	task.DelFlag = 0
	task.DelTime = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)

	// 保存到数据库
	if err := dao.DB.Create(task).Error; err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	return nil
}

// UpdateTask 更新任务
func (dao *TaskDao) UpdateTask(task *TaskMetaData) error {
	// 设置更新时间
	task.UpdateTime = time.Now()

	// 更新数据库
	result := dao.DB.Model(task).Updates(map[string]interface{}{
		"execute_at":    task.ExecuteAt,
		"name":          task.Name,
		"data_type":     task.DataType,
		"ledger_path":   task.LedgerPath,
		"mail_receiver": task.MailReceiver,
		"mail_type":     task.MailType,
		"mail_header":   task.MailHeader,
		"update_time":   task.UpdateTime,
		"update_by":     task.UpdateBy,
	})

	if result.Error != nil {
		return fmt.Errorf("更新任务失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("任务不存在或未更新")
	}

	return nil
}

// DeleteTask 删除任务（逻辑删除）
func (dao *TaskDao) DeleteTask(id uint, updateBy string) error {
	// 设置删除标志和删除时间
	result := dao.DB.Model(&TaskMetaData{}).
		Where("id = ? AND del_flag = ?", id, 0).
		Updates(map[string]interface{}{
			"del_flag":    1,
			"del_time":    time.Now(),
			"update_by":   updateBy,
			"update_time": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("删除任务失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("任务不存在或已被删除")
	}

	return nil
}
