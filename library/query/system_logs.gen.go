// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"github.com/haomiao000/raftchain/library/model"
)

func newSystemLog(db *gorm.DB, opts ...gen.DOOption) systemLog {
	_systemLog := systemLog{}

	_systemLog.systemLogDo.UseDB(db, opts...)
	_systemLog.systemLogDo.UseModel(&model.SystemLog{})

	tableName := _systemLog.systemLogDo.TableName()
	_systemLog.ALL = field.NewAsterisk(tableName)
	_systemLog.ID = field.NewInt64(tableName, "id")
	_systemLog.Timestamp = field.NewTime(tableName, "timestamp")
	_systemLog.NodeID = field.NewInt32(tableName, "node_id")
	_systemLog.Level = field.NewString(tableName, "level")
	_systemLog.Message = field.NewString(tableName, "message")
	_systemLog.CreatedAt = field.NewTime(tableName, "created_at")

	_systemLog.fillFieldMap()

	return _systemLog
}

type systemLog struct {
	systemLogDo

	ALL       field.Asterisk
	ID        field.Int64
	Timestamp field.Time
	NodeID    field.Int32
	Level     field.String
	Message   field.String
	CreatedAt field.Time

	fieldMap map[string]field.Expr
}

func (s systemLog) Table(newTableName string) *systemLog {
	s.systemLogDo.UseTable(newTableName)
	return s.updateTableName(newTableName)
}

func (s systemLog) As(alias string) *systemLog {
	s.systemLogDo.DO = *(s.systemLogDo.As(alias).(*gen.DO))
	return s.updateTableName(alias)
}

func (s *systemLog) updateTableName(table string) *systemLog {
	s.ALL = field.NewAsterisk(table)
	s.ID = field.NewInt64(table, "id")
	s.Timestamp = field.NewTime(table, "timestamp")
	s.NodeID = field.NewInt32(table, "node_id")
	s.Level = field.NewString(table, "level")
	s.Message = field.NewString(table, "message")
	s.CreatedAt = field.NewTime(table, "created_at")

	s.fillFieldMap()

	return s
}

func (s *systemLog) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := s.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (s *systemLog) fillFieldMap() {
	s.fieldMap = make(map[string]field.Expr, 6)
	s.fieldMap["id"] = s.ID
	s.fieldMap["timestamp"] = s.Timestamp
	s.fieldMap["node_id"] = s.NodeID
	s.fieldMap["level"] = s.Level
	s.fieldMap["message"] = s.Message
	s.fieldMap["created_at"] = s.CreatedAt
}

func (s systemLog) clone(db *gorm.DB) systemLog {
	s.systemLogDo.ReplaceConnPool(db.Statement.ConnPool)
	return s
}

func (s systemLog) replaceDB(db *gorm.DB) systemLog {
	s.systemLogDo.ReplaceDB(db)
	return s
}

type systemLogDo struct{ gen.DO }

type ISystemLogDo interface {
	gen.SubQuery
	Debug() ISystemLogDo
	WithContext(ctx context.Context) ISystemLogDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() ISystemLogDo
	WriteDB() ISystemLogDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) ISystemLogDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) ISystemLogDo
	Not(conds ...gen.Condition) ISystemLogDo
	Or(conds ...gen.Condition) ISystemLogDo
	Select(conds ...field.Expr) ISystemLogDo
	Where(conds ...gen.Condition) ISystemLogDo
	Order(conds ...field.Expr) ISystemLogDo
	Distinct(cols ...field.Expr) ISystemLogDo
	Omit(cols ...field.Expr) ISystemLogDo
	Join(table schema.Tabler, on ...field.Expr) ISystemLogDo
	LeftJoin(table schema.Tabler, on ...field.Expr) ISystemLogDo
	RightJoin(table schema.Tabler, on ...field.Expr) ISystemLogDo
	Group(cols ...field.Expr) ISystemLogDo
	Having(conds ...gen.Condition) ISystemLogDo
	Limit(limit int) ISystemLogDo
	Offset(offset int) ISystemLogDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) ISystemLogDo
	Unscoped() ISystemLogDo
	Create(values ...*model.SystemLog) error
	CreateInBatches(values []*model.SystemLog, batchSize int) error
	Save(values ...*model.SystemLog) error
	First() (*model.SystemLog, error)
	Take() (*model.SystemLog, error)
	Last() (*model.SystemLog, error)
	Find() ([]*model.SystemLog, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.SystemLog, err error)
	FindInBatches(result *[]*model.SystemLog, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.SystemLog) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) ISystemLogDo
	Assign(attrs ...field.AssignExpr) ISystemLogDo
	Joins(fields ...field.RelationField) ISystemLogDo
	Preload(fields ...field.RelationField) ISystemLogDo
	FirstOrInit() (*model.SystemLog, error)
	FirstOrCreate() (*model.SystemLog, error)
	FindByPage(offset int, limit int) (result []*model.SystemLog, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Rows() (*sql.Rows, error)
	Row() *sql.Row
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) ISystemLogDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (s systemLogDo) Debug() ISystemLogDo {
	return s.withDO(s.DO.Debug())
}

func (s systemLogDo) WithContext(ctx context.Context) ISystemLogDo {
	return s.withDO(s.DO.WithContext(ctx))
}

func (s systemLogDo) ReadDB() ISystemLogDo {
	return s.Clauses(dbresolver.Read)
}

func (s systemLogDo) WriteDB() ISystemLogDo {
	return s.Clauses(dbresolver.Write)
}

func (s systemLogDo) Session(config *gorm.Session) ISystemLogDo {
	return s.withDO(s.DO.Session(config))
}

func (s systemLogDo) Clauses(conds ...clause.Expression) ISystemLogDo {
	return s.withDO(s.DO.Clauses(conds...))
}

func (s systemLogDo) Returning(value interface{}, columns ...string) ISystemLogDo {
	return s.withDO(s.DO.Returning(value, columns...))
}

func (s systemLogDo) Not(conds ...gen.Condition) ISystemLogDo {
	return s.withDO(s.DO.Not(conds...))
}

func (s systemLogDo) Or(conds ...gen.Condition) ISystemLogDo {
	return s.withDO(s.DO.Or(conds...))
}

func (s systemLogDo) Select(conds ...field.Expr) ISystemLogDo {
	return s.withDO(s.DO.Select(conds...))
}

func (s systemLogDo) Where(conds ...gen.Condition) ISystemLogDo {
	return s.withDO(s.DO.Where(conds...))
}

func (s systemLogDo) Order(conds ...field.Expr) ISystemLogDo {
	return s.withDO(s.DO.Order(conds...))
}

func (s systemLogDo) Distinct(cols ...field.Expr) ISystemLogDo {
	return s.withDO(s.DO.Distinct(cols...))
}

func (s systemLogDo) Omit(cols ...field.Expr) ISystemLogDo {
	return s.withDO(s.DO.Omit(cols...))
}

func (s systemLogDo) Join(table schema.Tabler, on ...field.Expr) ISystemLogDo {
	return s.withDO(s.DO.Join(table, on...))
}

func (s systemLogDo) LeftJoin(table schema.Tabler, on ...field.Expr) ISystemLogDo {
	return s.withDO(s.DO.LeftJoin(table, on...))
}

func (s systemLogDo) RightJoin(table schema.Tabler, on ...field.Expr) ISystemLogDo {
	return s.withDO(s.DO.RightJoin(table, on...))
}

func (s systemLogDo) Group(cols ...field.Expr) ISystemLogDo {
	return s.withDO(s.DO.Group(cols...))
}

func (s systemLogDo) Having(conds ...gen.Condition) ISystemLogDo {
	return s.withDO(s.DO.Having(conds...))
}

func (s systemLogDo) Limit(limit int) ISystemLogDo {
	return s.withDO(s.DO.Limit(limit))
}

func (s systemLogDo) Offset(offset int) ISystemLogDo {
	return s.withDO(s.DO.Offset(offset))
}

func (s systemLogDo) Scopes(funcs ...func(gen.Dao) gen.Dao) ISystemLogDo {
	return s.withDO(s.DO.Scopes(funcs...))
}

func (s systemLogDo) Unscoped() ISystemLogDo {
	return s.withDO(s.DO.Unscoped())
}

func (s systemLogDo) Create(values ...*model.SystemLog) error {
	if len(values) == 0 {
		return nil
	}
	return s.DO.Create(values)
}

func (s systemLogDo) CreateInBatches(values []*model.SystemLog, batchSize int) error {
	return s.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (s systemLogDo) Save(values ...*model.SystemLog) error {
	if len(values) == 0 {
		return nil
	}
	return s.DO.Save(values)
}

func (s systemLogDo) First() (*model.SystemLog, error) {
	if result, err := s.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.SystemLog), nil
	}
}

func (s systemLogDo) Take() (*model.SystemLog, error) {
	if result, err := s.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.SystemLog), nil
	}
}

func (s systemLogDo) Last() (*model.SystemLog, error) {
	if result, err := s.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.SystemLog), nil
	}
}

func (s systemLogDo) Find() ([]*model.SystemLog, error) {
	result, err := s.DO.Find()
	return result.([]*model.SystemLog), err
}

func (s systemLogDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.SystemLog, err error) {
	buf := make([]*model.SystemLog, 0, batchSize)
	err = s.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (s systemLogDo) FindInBatches(result *[]*model.SystemLog, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return s.DO.FindInBatches(result, batchSize, fc)
}

func (s systemLogDo) Attrs(attrs ...field.AssignExpr) ISystemLogDo {
	return s.withDO(s.DO.Attrs(attrs...))
}

func (s systemLogDo) Assign(attrs ...field.AssignExpr) ISystemLogDo {
	return s.withDO(s.DO.Assign(attrs...))
}

func (s systemLogDo) Joins(fields ...field.RelationField) ISystemLogDo {
	for _, _f := range fields {
		s = *s.withDO(s.DO.Joins(_f))
	}
	return &s
}

func (s systemLogDo) Preload(fields ...field.RelationField) ISystemLogDo {
	for _, _f := range fields {
		s = *s.withDO(s.DO.Preload(_f))
	}
	return &s
}

func (s systemLogDo) FirstOrInit() (*model.SystemLog, error) {
	if result, err := s.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.SystemLog), nil
	}
}

func (s systemLogDo) FirstOrCreate() (*model.SystemLog, error) {
	if result, err := s.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.SystemLog), nil
	}
}

func (s systemLogDo) FindByPage(offset int, limit int) (result []*model.SystemLog, count int64, err error) {
	result, err = s.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = s.Offset(-1).Limit(-1).Count()
	return
}

func (s systemLogDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = s.Count()
	if err != nil {
		return
	}

	err = s.Offset(offset).Limit(limit).Scan(result)
	return
}

func (s systemLogDo) Scan(result interface{}) (err error) {
	return s.DO.Scan(result)
}

func (s systemLogDo) Delete(models ...*model.SystemLog) (result gen.ResultInfo, err error) {
	return s.DO.Delete(models)
}

func (s *systemLogDo) withDO(do gen.Dao) *systemLogDo {
	s.DO = *do.(*gen.DO)
	return s
}
