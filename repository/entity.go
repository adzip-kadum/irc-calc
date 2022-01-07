package repository

import (
	"context"

	"github.com/adzip-kadum/irc-calc/errs"
	"github.com/adzip-kadum/irc-calc/postgres"
	"github.com/pkg/errors"
)

type CalcsRepository struct {
	pool *postgres.PgxPool
}

func NewCalcsRepository(pool *postgres.PgxPool) *CalcsRepository {
	return &CalcsRepository{
		pool: pool,
	}
}

func (r *CalcsRepository) AddCalc(ctx context.Context, params AddCalcParams) (_ int64, reterr error) {
	defer errs.Recover(&reterr)

	q, closer, err := getDB(ctx, r.pool)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer closer()

	id, err := q.AddCalc(ctx, params)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return id, nil
}

func (r *CalcsRepository) GetCalcs(ctx context.Context, params GetCalcsParams) (_ []IrcCalc, reterr error) {
	defer errs.Recover(&reterr)

	q, closer, err := getDB(ctx, r.pool)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer closer()

	list, err := q.GetCalcs(ctx, params)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return list, nil
}

func getDB(ctx context.Context, pool *postgres.PgxPool) (*Queries, func(), error) {
	tx := postgres.GetTx(ctx)
	if tx != nil {
		return New(tx), func() {}, nil
	}
	conn, err := pool.Pool().Acquire(ctx)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return New(conn), func() { conn.Release() }, nil
}
