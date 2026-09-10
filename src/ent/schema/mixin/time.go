// Package mixin은 여러 Ent 스키마에서 재사용하는 공통 필드를 정의한다.
package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	entmixin "entgo.io/ent/schema/mixin"
)

// Time은 생성 시각과 마지막 수정 시각을 추가한다.
type Time struct {
	entmixin.Schema
}

var _ ent.Mixin = Time{}

func (Time) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Default(time.Now).Immutable(),
		// SQL 기본값은 기존 행에도 NOT NULL 컬럼을 추가할 수 있게 한다.
		// 수정 시 자동 갱신은 DB 트리거가 아니라 Ent의 UpdateDefault가 담당한다.
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).
			Annotations(entsql.DefaultExpr("CURRENT_TIMESTAMP")),
	}
}
