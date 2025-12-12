package postgres

import (
	"fmt"
	"strings"
	"time"
)

// QueryBuilder строит SQL-запросы безопасно
type QueryBuilder struct {
	baseQuery  string
	conditions []string
	args       []any
	argCounter int
}

// NewQueryBuilder создает новый builder
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		baseQuery:  baseQuery,
		conditions: []string{},
		args:       []any{},
		argCounter: 1,
	}
}

// AddCondition добавляет WHERE-условие
func (qb *QueryBuilder) AddCondition(condition string, arg any) *QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s = $%d", condition, qb.argCounter))
	qb.args = append(qb.args, arg)
	qb.argCounter++
	return qb
}

// AddLikeCondition добавляет ILIKE-условие
func (qb *QueryBuilder) AddLikeCondition(condition string, arg string) *QueryBuilder {
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s ILIKE $%d", condition, qb.argCounter))
	qb.args = append(qb.args, "%"+arg+"%")
	qb.argCounter++
	return qb
}

// AddDateRangeCondition добавляет условие для диапазона дат
func (qb *QueryBuilder) AddDateRangeCondition(startField, endField string, start, end time.Time) *QueryBuilder {
	// start_date <= $end AND (end_date IS NULL OR end_date >= $start)
	condition := fmt.Sprintf(
		"%s <= $%d AND (%s IS NULL OR %s >= $%d)",
		startField, qb.argCounter, endField, endField, qb.argCounter+1,
	)
	qb.conditions = append(qb.conditions, condition)
	qb.args = append(qb.args, end, start)
	qb.argCounter += 2
	return qb
}

// AddPagination добавляет LIMIT и OFFSET
func (qb *QueryBuilder) AddPagination(limit, offset int) *QueryBuilder {
	qb.baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", qb.argCounter, qb.argCounter+1)
	qb.args = append(qb.args, limit, offset)
	qb.argCounter += 2
	return qb
}

// AddOrderBy добавляет сортировку
func (qb *QueryBuilder) AddOrderBy(field, direction string) *QueryBuilder {
	// Защита от SQL-инъекций через whitelist
	allowedFields := map[string]bool{
		"created_at":   true,
		"updated_at":   true,
		"start_date":   true,
		"price":        true,
		"service_name": true,
	}

	allowedDirections := map[string]bool{
		"ASC":  true,
		"DESC": true,
	}

	if !allowedFields[field] {
		field = "created_at" 
	}

	direction = strings.ToUpper(direction)
	if !allowedDirections[direction] {
		direction = "DESC" 
	}

	qb.baseQuery += fmt.Sprintf(" ORDER BY %s %s", field, direction)
	return qb
}

// Build собирает финальный запрос
func (qb *QueryBuilder) Build() (string, []any) {
	query := qb.baseQuery

	if len(qb.conditions) > 0 {
		query += " WHERE " + strings.Join(qb.conditions, " AND ")
	}

	return query, qb.args
}
