package main

import "time"

type Config struct {
	HostPort          string        `long:"http-host-port" description:"Host port to start server" env:"HOST_PORT"`
	DbDsn             string        `long:"db-dsn" description:"Dsn to connect to the postgreSQL" env:"DB_DSN"`
	DbConnMaxLifetime time.Duration `long:"db-conn-max-lifetime" description:"ConnMaxLifetime " env:"DB_CONN_MAX_LIFETIME"`
	DbMaxOpenConns    int           `long:"db-max-open-conns" description:"Db proper configuration of max open connections" env:"DB_MAX_OPEN_CONNS"`
	DbMaxIdleConns    int           `long:"db-max-idle-conns" description:"Db proper configuration of max idle connections " env:"DB_MAX_IDLE_CONNS"`
	AmqpConnectionURL string        `long:"amqp-conn-url" description:"URL to connect to rabbitMQ" env:"AMQP_CONNECTION_URL"`
}
