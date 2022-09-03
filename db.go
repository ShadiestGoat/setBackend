package main

import (
	"context"
	"time"

	"github.com/Dextication/snowflake"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

var (
	SnowNode *snowflake.Node
	DB *pgx.Conn
	BASE_ID_TIME     = time.Date(2021, time.June, 1, 0, 0, 0, 0, time.UTC)
	BASE_ID_STAMP    = BASE_ID_TIME.UnixMilli()
)

func init() {
	node, err := snowflake.NewNode(0, BASE_ID_TIME, 41, 11, 11)
	PanicIfErr(err)
	SnowNode = node
}

const sqlUserTable = `CREATE TABLE IF NOT EXISTS users (
id TEXT PRIMARY KEY,
token TEXT,
name TEXT,
wins INTEGER,
losses INTEGER
);`

func InitDB() {
	node, err := snowflake.NewNode(0, BASE_ID_TIME, 41, 11, 11)
	PanicIfErr(err)
	SnowNode = node

	config, err := pgx.ParseConfig(DB_URL)
	if err != nil {
		panic("The provided 'DB_URL' is not valid.")
	}

	conn, err := pgx.ConnectConfig(context.Background(), config)
	PanicIfErr(err)

	err = conn.Ping(context.Background())
	PanicIfErr(err)
	
	DB = conn

	_, err = DBExec(sqlUserTable)
	PanicIfErr(err)

}

func DBExec(sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return DB.Exec(context.Background(), sql, args...)
}


func DBQuery(sql string, args ...interface{}) (pgx.Rows, error) {
	return DB.Query(context.Background(), sql, args...)
}


func DBQueryRow(sql string, args ...interface{}) (pgx.Row) {
	return DB.QueryRow(context.Background(), sql, args...)
}
