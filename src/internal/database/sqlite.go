// Package database는 DB 연결과 예제용 스키마 초기화를 담당한다.
package database

import (
	"context"
	"database/sql"
	"fmt"

	"gin-start/src/ent"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3"
)

func Open(ctx context.Context, dsn string) (*ent.Client, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	// 작은 SQLite 예제의 쓰기 경합을 피하고 메모리 DB도 같은 연결에서 사용한다.
	db.SetMaxOpenConns(1)
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	if err := db.PingContext(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	// 학습용 자동 마이그레이션. 운영에서는 버전 관리된 마이그레이션을 별도로 실행한다.
	if err := client.Schema.Create(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return client, nil
}
