package database

type BackendDB string

type Config struct {
	BackendDB           BackendDB
	ConnectionPath      string
	NumberOfConnections int
}
