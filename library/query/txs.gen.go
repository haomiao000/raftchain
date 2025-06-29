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

func newTx(db *gorm.DB, opts ...gen.DOOption) tx {
	_tx := tx{}

	_tx.txDo.UseDB(db, opts...)
	_tx.txDo.UseModel(&model.Tx{})

	tableName := _tx.txDo.TableName()
	_tx.ALL = field.NewAsterisk(tableName)
	_tx.ID = field.NewInt64(tableName, "id")
	_tx.TxHash = field.NewString(tableName, "tx_hash")
	_tx.BlockHeight = field.NewInt64(tableName, "block_height")
	_tx.BlockHash = field.NewString(tableName, "block_hash")
	_tx.SenderPublicKey = field.NewString(tableName, "sender_public_key")
	_tx.Type = field.NewString(tableName, "type")
	_tx.Payload = field.NewBytes(tableName, "payload")
	_tx.CreatedAt = field.NewTime(tableName, "created_at")

	_tx.fillFieldMap()

	return _tx
}

// tx 交易记录表
type tx struct {
	txDo

	ALL             field.Asterisk
	ID              field.Int64
	TxHash          field.String // 交易哈希
	BlockHeight     field.Int64  // 所属区块高度
	BlockHash       field.String // 所属区块哈希
	SenderPublicKey field.String // 发送者公钥
	Type            field.String // 交易类型
	Payload         field.Bytes  // 交易载荷（具体内容）
	CreatedAt       field.Time

	fieldMap map[string]field.Expr
}

func (t tx) Table(newTableName string) *tx {
	t.txDo.UseTable(newTableName)
	return t.updateTableName(newTableName)
}

func (t tx) As(alias string) *tx {
	t.txDo.DO = *(t.txDo.As(alias).(*gen.DO))
	return t.updateTableName(alias)
}

func (t *tx) updateTableName(table string) *tx {
	t.ALL = field.NewAsterisk(table)
	t.ID = field.NewInt64(table, "id")
	t.TxHash = field.NewString(table, "tx_hash")
	t.BlockHeight = field.NewInt64(table, "block_height")
	t.BlockHash = field.NewString(table, "block_hash")
	t.SenderPublicKey = field.NewString(table, "sender_public_key")
	t.Type = field.NewString(table, "type")
	t.Payload = field.NewBytes(table, "payload")
	t.CreatedAt = field.NewTime(table, "created_at")

	t.fillFieldMap()

	return t
}

func (t *tx) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := t.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (t *tx) fillFieldMap() {
	t.fieldMap = make(map[string]field.Expr, 8)
	t.fieldMap["id"] = t.ID
	t.fieldMap["tx_hash"] = t.TxHash
	t.fieldMap["block_height"] = t.BlockHeight
	t.fieldMap["block_hash"] = t.BlockHash
	t.fieldMap["sender_public_key"] = t.SenderPublicKey
	t.fieldMap["type"] = t.Type
	t.fieldMap["payload"] = t.Payload
	t.fieldMap["created_at"] = t.CreatedAt
}

func (t tx) clone(db *gorm.DB) tx {
	t.txDo.ReplaceConnPool(db.Statement.ConnPool)
	return t
}

func (t tx) replaceDB(db *gorm.DB) tx {
	t.txDo.ReplaceDB(db)
	return t
}

type txDo struct{ gen.DO }

type ITxDo interface {
	gen.SubQuery
	Debug() ITxDo
	WithContext(ctx context.Context) ITxDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() ITxDo
	WriteDB() ITxDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) ITxDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) ITxDo
	Not(conds ...gen.Condition) ITxDo
	Or(conds ...gen.Condition) ITxDo
	Select(conds ...field.Expr) ITxDo
	Where(conds ...gen.Condition) ITxDo
	Order(conds ...field.Expr) ITxDo
	Distinct(cols ...field.Expr) ITxDo
	Omit(cols ...field.Expr) ITxDo
	Join(table schema.Tabler, on ...field.Expr) ITxDo
	LeftJoin(table schema.Tabler, on ...field.Expr) ITxDo
	RightJoin(table schema.Tabler, on ...field.Expr) ITxDo
	Group(cols ...field.Expr) ITxDo
	Having(conds ...gen.Condition) ITxDo
	Limit(limit int) ITxDo
	Offset(offset int) ITxDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) ITxDo
	Unscoped() ITxDo
	Create(values ...*model.Tx) error
	CreateInBatches(values []*model.Tx, batchSize int) error
	Save(values ...*model.Tx) error
	First() (*model.Tx, error)
	Take() (*model.Tx, error)
	Last() (*model.Tx, error)
	Find() ([]*model.Tx, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.Tx, err error)
	FindInBatches(result *[]*model.Tx, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.Tx) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) ITxDo
	Assign(attrs ...field.AssignExpr) ITxDo
	Joins(fields ...field.RelationField) ITxDo
	Preload(fields ...field.RelationField) ITxDo
	FirstOrInit() (*model.Tx, error)
	FirstOrCreate() (*model.Tx, error)
	FindByPage(offset int, limit int) (result []*model.Tx, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Rows() (*sql.Rows, error)
	Row() *sql.Row
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) ITxDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (t txDo) Debug() ITxDo {
	return t.withDO(t.DO.Debug())
}

func (t txDo) WithContext(ctx context.Context) ITxDo {
	return t.withDO(t.DO.WithContext(ctx))
}

func (t txDo) ReadDB() ITxDo {
	return t.Clauses(dbresolver.Read)
}

func (t txDo) WriteDB() ITxDo {
	return t.Clauses(dbresolver.Write)
}

func (t txDo) Session(config *gorm.Session) ITxDo {
	return t.withDO(t.DO.Session(config))
}

func (t txDo) Clauses(conds ...clause.Expression) ITxDo {
	return t.withDO(t.DO.Clauses(conds...))
}

func (t txDo) Returning(value interface{}, columns ...string) ITxDo {
	return t.withDO(t.DO.Returning(value, columns...))
}

func (t txDo) Not(conds ...gen.Condition) ITxDo {
	return t.withDO(t.DO.Not(conds...))
}

func (t txDo) Or(conds ...gen.Condition) ITxDo {
	return t.withDO(t.DO.Or(conds...))
}

func (t txDo) Select(conds ...field.Expr) ITxDo {
	return t.withDO(t.DO.Select(conds...))
}

func (t txDo) Where(conds ...gen.Condition) ITxDo {
	return t.withDO(t.DO.Where(conds...))
}

func (t txDo) Order(conds ...field.Expr) ITxDo {
	return t.withDO(t.DO.Order(conds...))
}

func (t txDo) Distinct(cols ...field.Expr) ITxDo {
	return t.withDO(t.DO.Distinct(cols...))
}

func (t txDo) Omit(cols ...field.Expr) ITxDo {
	return t.withDO(t.DO.Omit(cols...))
}

func (t txDo) Join(table schema.Tabler, on ...field.Expr) ITxDo {
	return t.withDO(t.DO.Join(table, on...))
}

func (t txDo) LeftJoin(table schema.Tabler, on ...field.Expr) ITxDo {
	return t.withDO(t.DO.LeftJoin(table, on...))
}

func (t txDo) RightJoin(table schema.Tabler, on ...field.Expr) ITxDo {
	return t.withDO(t.DO.RightJoin(table, on...))
}

func (t txDo) Group(cols ...field.Expr) ITxDo {
	return t.withDO(t.DO.Group(cols...))
}

func (t txDo) Having(conds ...gen.Condition) ITxDo {
	return t.withDO(t.DO.Having(conds...))
}

func (t txDo) Limit(limit int) ITxDo {
	return t.withDO(t.DO.Limit(limit))
}

func (t txDo) Offset(offset int) ITxDo {
	return t.withDO(t.DO.Offset(offset))
}

func (t txDo) Scopes(funcs ...func(gen.Dao) gen.Dao) ITxDo {
	return t.withDO(t.DO.Scopes(funcs...))
}

func (t txDo) Unscoped() ITxDo {
	return t.withDO(t.DO.Unscoped())
}

func (t txDo) Create(values ...*model.Tx) error {
	if len(values) == 0 {
		return nil
	}
	return t.DO.Create(values)
}

func (t txDo) CreateInBatches(values []*model.Tx, batchSize int) error {
	return t.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (t txDo) Save(values ...*model.Tx) error {
	if len(values) == 0 {
		return nil
	}
	return t.DO.Save(values)
}

func (t txDo) First() (*model.Tx, error) {
	if result, err := t.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.Tx), nil
	}
}

func (t txDo) Take() (*model.Tx, error) {
	if result, err := t.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.Tx), nil
	}
}

func (t txDo) Last() (*model.Tx, error) {
	if result, err := t.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.Tx), nil
	}
}

func (t txDo) Find() ([]*model.Tx, error) {
	result, err := t.DO.Find()
	return result.([]*model.Tx), err
}

func (t txDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.Tx, err error) {
	buf := make([]*model.Tx, 0, batchSize)
	err = t.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (t txDo) FindInBatches(result *[]*model.Tx, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return t.DO.FindInBatches(result, batchSize, fc)
}

func (t txDo) Attrs(attrs ...field.AssignExpr) ITxDo {
	return t.withDO(t.DO.Attrs(attrs...))
}

func (t txDo) Assign(attrs ...field.AssignExpr) ITxDo {
	return t.withDO(t.DO.Assign(attrs...))
}

func (t txDo) Joins(fields ...field.RelationField) ITxDo {
	for _, _f := range fields {
		t = *t.withDO(t.DO.Joins(_f))
	}
	return &t
}

func (t txDo) Preload(fields ...field.RelationField) ITxDo {
	for _, _f := range fields {
		t = *t.withDO(t.DO.Preload(_f))
	}
	return &t
}

func (t txDo) FirstOrInit() (*model.Tx, error) {
	if result, err := t.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.Tx), nil
	}
}

func (t txDo) FirstOrCreate() (*model.Tx, error) {
	if result, err := t.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.Tx), nil
	}
}

func (t txDo) FindByPage(offset int, limit int) (result []*model.Tx, count int64, err error) {
	result, err = t.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = t.Offset(-1).Limit(-1).Count()
	return
}

func (t txDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = t.Count()
	if err != nil {
		return
	}

	err = t.Offset(offset).Limit(limit).Scan(result)
	return
}

func (t txDo) Scan(result interface{}) (err error) {
	return t.DO.Scan(result)
}

func (t txDo) Delete(models ...*model.Tx) (result gen.ResultInfo, err error) {
	return t.DO.Delete(models)
}

func (t *txDo) withDO(do gen.Dao) *txDo {
	t.DO = *do.(*gen.DO)
	return t
}
