package sqllite

import (
	"aigcpanel/go/internal/component/log"
	"aigcpanel/go/internal/utils"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"aigcpanel/go/internal/domain"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func Init() {

	_, taskStoreErr := NewSQLiteStore(utils.SQLiteFile)
	if taskStoreErr != nil {
		log.Logger.Error(taskStoreErr.Error())
	}

}

func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLiteStore) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS data_task (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            createdAt INTEGER NOT NULL,
            updatedAt INTEGER NOT NULL,
            biz TEXT,
            type INTEGER DEFAULT 1,
            title TEXT,
            status TEXT,
            statusMsg TEXT,
            startTime INTEGER,
            endTime INTEGER,
            serverName TEXT,
            serverTitle TEXT,
            serverVersion TEXT,
            param TEXT,
            jobResult TEXT,
            modelConfig TEXT,
            result TEXT
        )`,
		`CREATE INDEX IF NOT EXISTS idx_data_task_biz ON data_task(biz)`,
		`CREATE INDEX IF NOT EXISTS idx_data_task_status ON data_task(status)`,
		`CREATE INDEX IF NOT EXISTS idx_data_task_type ON data_task(type)`,
	}
	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

type TaskFilters struct {
	Biz    string
	Status []string
	Type   *int
}

func (s *SQLiteStore) CreateTask(task domain.AppTask) (domain.AppTask, error) {
	now := time.Now().UnixMilli()
	if task.CreatedAt == 0 {
		task.CreatedAt = now
	}
	if task.UpdatedAt == 0 {
		task.UpdatedAt = now
	}
	if task.Type == 0 {
		task.Type = 1
	}
	res, err := s.db.Exec(
		`INSERT INTO data_task
            (biz, type, title, status, statusMsg, startTime, endTime, serverName, serverTitle, serverVersion, param, jobResult, modelConfig, result, createdAt, updatedAt)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.Biz,
		task.Type,
		task.Title,
		task.Status,
		task.StatusMsg,
		task.StartTime,
		task.EndTime,
		task.ServerName,
		task.ServerTitle,
		task.ServerVersion,
		task.Param,
		task.JobResult,
		task.ModelConfig,
		task.Result,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return domain.AppTask{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.AppTask{}, err
	}
	task.ID = id
	return task, nil
}

func (s *SQLiteStore) GetTask(id int64) (domain.AppTask, error) {
	var task domain.AppTask
	row := s.db.QueryRow(
		`SELECT id, biz, type, title, status, statusMsg, startTime, endTime, serverName, serverTitle, serverVersion, param, jobResult, modelConfig, result, createdAt, updatedAt
         FROM data_task WHERE id = ?`,
		id,
	)
	if err := row.Scan(
		&task.ID,
		&task.Biz,
		&task.Type,
		&task.Title,
		&task.Status,
		&task.StatusMsg,
		&task.StartTime,
		&task.EndTime,
		&task.ServerName,
		&task.ServerTitle,
		&task.ServerVersion,
		&task.Param,
		&task.JobResult,
		&task.ModelConfig,
		&task.Result,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return domain.AppTask{}, err
	}
	return task, nil
}

func (s *SQLiteStore) ListTasks(filters TaskFilters) ([]domain.AppTask, error) {
	query := `SELECT id, biz, type, title, status, statusMsg, startTime, endTime, serverName, serverTitle, serverVersion, param, jobResult, modelConfig, result, createdAt, updatedAt
        FROM data_task`
	where := make([]string, 0)
	args := make([]interface{}, 0)
	if filters.Biz != "" {
		where = append(where, "biz = ?")
		args = append(args, filters.Biz)
	}
	if len(filters.Status) > 0 {
		placeholders := make([]string, 0, len(filters.Status))
		for _, s := range filters.Status {
			if s == "" {
				continue
			}
			placeholders = append(placeholders, "?")
			args = append(args, s)
		}
		if len(placeholders) > 0 {
			where = append(where, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
		}
	}
	if filters.Type != nil {
		where = append(where, "type = ?")
		args = append(args, *filters.Type)
	}
	if len(where) > 0 {
		query = query + " WHERE " + strings.Join(where, " AND ")
	}
	query = query + " ORDER BY id DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domain.AppTask, 0)
	for rows.Next() {
		var task domain.AppTask
		if err := rows.Scan(
			&task.ID,
			&task.Biz,
			&task.Type,
			&task.Title,
			&task.Status,
			&task.StatusMsg,
			&task.StartTime,
			&task.EndTime,
			&task.ServerName,
			&task.ServerTitle,
			&task.ServerVersion,
			&task.Param,
			&task.JobResult,
			&task.ModelConfig,
			&task.Result,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *SQLiteStore) UpdateTask(id int64, updates map[string]any) (domain.AppTask, error) {
	sets := make([]string, 0)
	args := make([]interface{}, 0)
	addField := func(name string, value interface{}) {
		sets = append(sets, fmt.Sprintf("%s = ?", name))
		args = append(args, value)
	}
	if v, ok := updates["biz"]; ok {
		addField("biz", v)
	}
	if v, ok := updates["type"]; ok {
		addField("type", v)
	}
	if v, ok := updates["title"]; ok {
		addField("title", v)
	}
	if v, ok := updates["status"]; ok {
		addField("status", v)
	}
	if v, ok := updates["statusMsg"]; ok {
		addField("statusMsg", v)
	}
	if v, ok := updates["startTime"]; ok {
		addField("startTime", v)
	}
	if v, ok := updates["endTime"]; ok {
		addField("endTime", v)
	}
	if v, ok := updates["serverName"]; ok {
		addField("serverName", v)
	}
	if v, ok := updates["serverTitle"]; ok {
		addField("serverTitle", v)
	}
	if v, ok := updates["serverVersion"]; ok {
		addField("serverVersion", v)
	}
	if v, ok := updates["param"]; ok {
		addField("param", v)
	}
	if v, ok := updates["jobResult"]; ok {
		addField("jobResult", v)
	}
	if v, ok := updates["modelConfig"]; ok {
		addField("modelConfig", v)
	}
	if v, ok := updates["result"]; ok {
		addField("result", v)
	}
	if len(sets) == 0 {
		return s.GetTask(id)
	}
	addField("updatedAt", time.Now().UnixMilli())
	args = append(args, id)

	query := fmt.Sprintf("UPDATE data_task SET %s WHERE id = ?", strings.Join(sets, ", "))
	if _, err := s.db.Exec(query, args...); err != nil {
		return domain.AppTask{}, err
	}
	return s.GetTask(id)
}

func (s *SQLiteStore) DeleteTask(id int64) error {
	_, err := s.db.Exec("DELETE FROM data_task WHERE id = ?", id)
	return err
}
