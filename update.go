package dba

import (
	"strings"
	"strconv"
	"bytes"
	"github.com/smartwalle/errors"
	"fmt"
	"database/sql"
)

type UpdateBuilder struct {
	prefixes     expressions
	options      expressions
	tables       expressions
	columns      []string
	values       []interface{}
	wheres       whereExpressions
	orderBys     []string
	limit        uint64
	updateLimit  bool
	offset       uint64
	updateOffset bool
	suffixes     expressions
}

func (this *UpdateBuilder) Prefix(sql string, args ...interface{}) *UpdateBuilder {
	this.prefixes = append(this.prefixes, Expression(sql, args...))
	return this
}

func (this *UpdateBuilder) Options(options ...string) *UpdateBuilder {
	for _, c := range options {
		this.options = append(this.options, Expression(c))
	}
	return this
}

func (this *UpdateBuilder) Table(table string, args ...string) *UpdateBuilder {
	var ts []string
	ts = append(ts, fmt.Sprintf("`%s`", table))
	ts = append(ts, args...)
	this.tables = append(this.tables, Expression(strings.Join(ts, " ")))
	return this
}

func (this *UpdateBuilder) SET(column string, value interface{}) *UpdateBuilder {
	this.columns = append(this.columns, column)
	this.values = append(this.values, value)
	return this
}

func (this *UpdateBuilder) SetMap(data map[string]interface{}) *UpdateBuilder {
	for k, v := range data {
		this.columns = append(this.columns, k)
		this.values = append(this.values, v)
	}
	return this
}

func (this *UpdateBuilder) Where(sql string, args ...interface{}) *UpdateBuilder {
	this.wheres = append(this.wheres, WhereExpression(sql, args...))
	return this
}

func (this *UpdateBuilder) OrderBy(sql ...string) *UpdateBuilder {
	this.orderBys = append(this.orderBys, sql...)
	return this
}

func (this *UpdateBuilder) Limit(limit uint64) *UpdateBuilder {
	this.limit = limit
	this.updateLimit = true
	return this
}

func (this *UpdateBuilder) Offset(offset uint64) *UpdateBuilder {
	this.offset = offset
	this.updateOffset = true
	return this
}

func (this *UpdateBuilder) Suffix(sql string, args ...interface{}) *UpdateBuilder {
	this.suffixes = append(this.suffixes, Expression(sql, args...))
	return this
}

func (this *UpdateBuilder) ToSQL() (sql string, args []interface{}, err error) {
	if len(this.tables) == 0 {
		return "", nil, errors.New("update statements must specify a table")
	}
	if len(this.columns) == 0 {
		return "", nil, errors.New("update statements must have at least one Set")
	}

	var sqlBuffer = &bytes.Buffer{}
	if len(this.prefixes) > 0 {
		args, _ = this.prefixes.appendToSQL(sqlBuffer, " ", args)
		sqlBuffer.WriteString(" ")
	}

	sqlBuffer.WriteString("UPDATE ")

	if len(this.options) > 0 {
		args, _ = this.options.appendToSQL(sqlBuffer, " ", args)
		sqlBuffer.WriteString(" ")
	}

	if len(this.tables) > 0 {
		args, _ = this.tables.appendToSQL(sqlBuffer, ", ", args)
	}

	sqlBuffer.WriteString(" SET ")

	if len(this.columns) > 0 {
		//args, _ = this.sets.appendToSQL(sqlBuffer, ", ", args)
		var cs []string
		for _, c := range this.columns {
			cs = append(cs, fmt.Sprintf("%s=?", c))
		}
		sqlBuffer.WriteString(strings.Join(cs, ", "))
		args = append(args, this.values...)
	}

	if len(this.wheres) == 0 {
		return "", nil, errors.New("update statements must have WHERE condition")
	}

	if len(this.wheres) > 0 {
		sqlBuffer.WriteString(" WHERE ")
		args, _ = this.wheres.appendToSQL(sqlBuffer, " ", args)
	}

	if len(this.orderBys) > 0 {
		sqlBuffer.WriteString(" ORDER BY ")
		sqlBuffer.WriteString(strings.Join(this.orderBys, ", "))
	}

	if this.updateLimit {
		sqlBuffer.WriteString(" LIMIT ")
		sqlBuffer.WriteString(strconv.FormatUint(this.limit, 10))
	}

	if this.updateOffset {
		sqlBuffer.WriteString(" OFFSET ")
		sqlBuffer.WriteString(strconv.FormatUint(this.offset, 10))
	}

	if len(this.suffixes) > 0 {
		sqlBuffer.WriteString(" ")
		args, _ = this.suffixes.appendToSQL(sqlBuffer, " ", args)
	}

	sql = sqlBuffer.String()

	return sql, args, err
}

func (this *UpdateBuilder) Exec(s StmtPrepare) (sql.Result, error) {
	sql, args, err := this.ToSQL()
	if err != nil {
		return nil, err
	}
	return Exec(s, sql, args...)
}

func NewUpdateBuilder() *UpdateBuilder {
	return &UpdateBuilder{}
}
