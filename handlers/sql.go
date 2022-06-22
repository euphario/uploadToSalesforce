package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (app *App) execSqlStatements(statements []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, statement := range statements {
		app.DB.ExecContext(ctx, statement)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (app *App) execSqlInsert(describe Describe, rows []map[string]interface{}) []string {
	var columns []string
	for _, v := range describe.Fields {
		columns = append(columns, v.Name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil
	}
	defer tx.Rollback()

	for _, row := range rows {
		var fields []string
		var values []string
		var prepareValues []string

		for _, column := range columns {
			fields = append(fields, column)
			values = append(values, row[column].(string))
			prepareValues = append(prepareValues, "?")
		}
		app.DB.Prepare(fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", describe.Name, strings.Join(fields, ", "), strings.Join(prepareValues, ", ")))
		// app.DB.Exec(values...)
	}

	if err = tx.Commit(); err != nil {
		return nil
	}
	return nil
}
