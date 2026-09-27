package main

import (
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

func poolConfig(databaseURL, maxValue, minValue string) (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	max, min := 5, 0
	if maxValue != "" {
		max, err = strconv.Atoi(maxValue)
		if err != nil || max < 1 || max > 100 {
			return nil, fmt.Errorf("DB_MAX_CONNS must be between 1 and 100")
		}
	}
	if minValue != "" {
		min, err = strconv.Atoi(minValue)
		if err != nil || min < 0 || min > max {
			return nil, fmt.Errorf("DB_MIN_CONNS must be between 0 and DB_MAX_CONNS")
		}
	}
	config.MaxConns = int32(max)
	config.MinConns = int32(min)
	return config, nil
}
