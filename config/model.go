package config

type ConfigModel struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	PgDriver string
}

type HTTPConfig struct {
	Host string
	Port string
}
