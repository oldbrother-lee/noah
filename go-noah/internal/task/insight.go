package task

import (
	"context"
	"fmt"
	"go-noah/internal/das/dao"
	"go-noah/internal/model/insight"
	insightRepo "go-noah/internal/repository/insight"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InsightTask 数据库元数据同步任务
type InsightTask struct {
	insightRepo *insightRepo.InsightRepository
	*Task
}

func NewInsightTask(
	task *Task,
	insightRepo *insightRepo.InsightRepository,
) *InsightTask {
	return &InsightTask{
		insightRepo: insightRepo,
		Task:        task,
	}
}

// 忽略的系统库
var ignoredSchemas = []string{
	"'PERFORMANCE_SCHEMA'",
	"'INFORMATION_SCHEMA'",
	"'performance_schema'",
	"'information_schema'",
	"'MYSQL'",
	"'mysql'",
	"'sys'",
	"'SYS'",
}

// MySQL/TiDB 查询语句
var mysqlQuery = fmt.Sprintf(`
	SELECT 
		SCHEMA_NAME AS TABLE_SCHEMA
	FROM 
		INFORMATION_SCHEMA.SCHEMATA
	WHERE 
		SCHEMA_NAME NOT IN (%s)
	`, strings.Join(ignoredSchemas, ","))

// ClickHouse 查询语句
var clickhouseQuery = fmt.Sprintf(`
	SELECT 
		name AS TABLE_SCHEMA
	FROM 
		system.databases
	WHERE 
		name NOT IN (%s)
`, strings.Join(ignoredSchemas, ","))

// createSchemaRecord 创建或更新 schema 记录
func (t *InsightTask) createSchemaRecord(ctx context.Context, instanceID uuid.UUID, schemaName string) error {
	var existing insight.DBSchema
	result := t.insightRepo.DB(ctx).Where("instance_id = ? AND `schema` = ?", instanceID, schemaName).First(&existing)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 创建新记录
			schema := insight.DBSchema{
				InstanceID: instanceID,
				Schema:     schemaName,
				IsDeleted:  false,
			}
			if err := t.insightRepo.DB(ctx).Create(&schema).Error; err != nil {
				return err
			}
			t.logger.Debug("创建 schema 记录",
				zap.String("instance_id", instanceID.String()),
				zap.String("schema", schemaName),
			)
		} else {
			return result.Error
		}
	} else {
		// 如果 schema 删除后又被新建，更新 is_deleted 状态
		if existing.IsDeleted {
			if err := t.insightRepo.DB(ctx).Model(&existing).Update("is_deleted", false).Error; err != nil {
				return err
			}
			t.logger.Debug("恢复 schema 记录",
				zap.String("instance_id", instanceID.String()),
				zap.String("schema", schemaName),
			)
		}
		// 如果记录已存在且未删除，不需要做任何操作
	}
	return nil
}

// updateSchemaRecordAsSoftDel 将 schema 记录标记为软删除
func (t *InsightTask) updateSchemaRecordAsSoftDel(ctx context.Context, instanceID uuid.UUID, schemaName string) error {
	return t.insightRepo.DB(ctx).Model(&insight.DBSchema{}).
		Where("instance_id = ? AND `schema` = ?", instanceID, schemaName).
		Update("is_deleted", true).Error
}

// checkSourceSchemasIsDeleted 检查源数据库中已删除的 schema
func (t *InsightTask) checkSourceSchemasIsDeleted(ctx context.Context, instanceID uuid.UUID, sourceSchemas []string) error {
	// 获取本地所有 schema（包括已删除的）
	var localSchemas []insight.DBSchema
	if err := t.insightRepo.DB(ctx).Where("instance_id = ?", instanceID).Find(&localSchemas).Error; err != nil {
		return err
	}

	// 找出源数据库中已删除的 schema
	for _, localSchema := range localSchemas {
		found := false
		for _, sourceSchema := range sourceSchemas {
			if localSchema.Schema == sourceSchema {
				found = true
				break
			}
		}
		if !found && !localSchema.IsDeleted {
			// 源数据库中已删除，标记为软删除
			if err := t.updateSchemaRecordAsSoftDel(ctx, instanceID, localSchema.Schema); err != nil {
				return err
			}
			t.logger.Info("标记 schema 为已删除",
				zap.String("instance_id", instanceID.String()),
				zap.String("schema", localSchema.Schema),
			)
		}
	}
	return nil
}

// SyncDBMeta 从用户定义的远程数据库实例同步库信息
func (t *InsightTask) SyncDBMeta(ctx context.Context) error {
	t.logger.Info("开始同步数据库元数据")

	// 获取所有数据库配置
	configs, err := t.insightRepo.GetDBConfigs(ctx, "", 0)
	if err != nil {
		t.logger.Error("获取数据库配置失败", zap.Error(err))
		return err
	}

	if len(configs) == 0 {
		t.logger.Info("没有需要同步的数据库配置")
		return nil
	}

	// 使用 4 个并发 goroutine 同步
	var wg sync.WaitGroup
	ch := make(chan struct{}, 4)

	for _, config := range configs {
		ch <- struct{}{}
		wg.Add(1)

		go func(cfg insight.DBConfig) {
			defer func() {
				<-ch
				wg.Done()
			}()

			// 执行 SQL 超时：10 秒
			queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			var data []map[string]interface{}
			var err error

			switch strings.ToLower(string(cfg.DbType)) {
			case "mysql", "tidb":
				db := dao.MySQLDB{
					User:     cfg.UserName,
					Password: cfg.Password,
					Host:     cfg.Hostname,
					Port:     cfg.Port,
					Params:   map[string]string{"group_concat_max_len": "67108864"},
					Ctx:      queryCtx,
				}
				_, data, err = db.Query(mysqlQuery)
			case "clickhouse":
				// TODO: 实现 ClickHouse 支持
				t.logger.Warn("ClickHouse 暂不支持同步元数据",
					zap.String("hostname", cfg.Hostname),
					zap.Int("port", cfg.Port),
				)
				return
			default:
				t.logger.Warn("不支持的数据库类型",
					zap.String("db_type", string(cfg.DbType)),
					zap.String("hostname", cfg.Hostname),
					zap.Int("port", cfg.Port),
				)
				return
			}

			if err != nil {
				t.logger.Error("同步元数据失败",
					zap.String("hostname", cfg.Hostname),
					zap.Int("port", cfg.Port),
					zap.String("instance_id", cfg.InstanceID.String()),
					zap.Error(err),
				)
				return
			}

			if len(data) == 0 {
				t.logger.Warn("未发现库记录",
					zap.String("hostname", cfg.Hostname),
					zap.Int("port", cfg.Port),
					zap.String("user_name", cfg.UserName),
				)
				return
			}

			// 创建或更新元数据记录
			var sourceSchemas []string
			for _, d := range data {
				schemaName, ok := d["TABLE_SCHEMA"].(string)
				if !ok {
					continue
				}
				sourceSchemas = append(sourceSchemas, schemaName)

				if err := t.createSchemaRecord(ctx, cfg.InstanceID, schemaName); err != nil {
					t.logger.Error("创建 schema 记录失败",
						zap.String("instance_id", cfg.InstanceID.String()),
						zap.String("schema", schemaName),
						zap.Error(err),
					)
					continue
				}

				t.logger.Debug("同步元数据成功",
					zap.String("hostname", cfg.Hostname),
					zap.Int("port", cfg.Port),
					zap.String("instance_id", cfg.InstanceID.String()),
					zap.String("schema", schemaName),
				)
			}

			// 检查源数据库中已删除的 schema
			if err := t.checkSourceSchemasIsDeleted(ctx, cfg.InstanceID, sourceSchemas); err != nil {
				t.logger.Error("检查已删除 schema 失败",
					zap.String("instance_id", cfg.InstanceID.String()),
					zap.Error(err),
				)
			}
		}(config)
	}

	wg.Wait()
	t.logger.Info("数据库元数据同步完成")
	return nil
}

