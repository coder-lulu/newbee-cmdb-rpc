package mixins

import (
	"context"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// SoftDeleteMixin implements the soft delete pattern for schemas.
type SoftDeleteMixin struct {
	mixin.Schema
}

// Fields of the SoftDeleteMixin.
func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").
			Optional().
			Comment("Delete Time | 删除日期"),
	}
}

type softDeleteKey struct{}

// SkipSoftDelete returns a new context that skips the soft-delete interceptor/mutators.
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

// Hooks of the SoftDeleteMixin.
func (SoftDeleteMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				// Skip soft-delete, means delete the entity permanently.
				if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
					return next.Mutate(ctx, m)
				}

				// Only apply to delete operations
				if m.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
					// Set deleted_at instead of actually deleting
					if setter, ok := m.(interface{ SetDeletedAt(time.Time) }); ok {
						setter.SetDeletedAt(time.Now())
						// Change operation to update
						if opSetter, ok := m.(interface{ SetOp(ent.Op) }); ok {
							opSetter.SetOp(ent.OpUpdate)
						}
					}
				}

				return next.Mutate(ctx, m)
			})
		},
	}
}

// softDeleteInterceptor implements ent.Interceptor for filtering soft-deleted records
type softDeleteInterceptor struct{}

func (softDeleteInterceptor) Intercept(next ent.Querier) ent.Querier {
	return ent.QuerierFunc(func(ctx context.Context, q ent.Query) (ent.Value, error) {
		// Skip soft-delete filter if explicitly requested
		if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
			return next.Query(ctx, q)
		}

		// Add WHERE deleted_at IS NULL to all queries
		if sqlQuery, ok := q.(*sql.Selector); ok {
			sqlQuery.Where(sql.IsNull("deleted_at"))
		}

		return next.Query(ctx, q)
	})
}

// Interceptors of the SoftDeleteMixin to filter out soft-deleted records in queries.
func (SoftDeleteMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		softDeleteInterceptor{},
	}
}
