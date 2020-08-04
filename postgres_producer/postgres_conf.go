package main

type PostgresConf struct {
	Host     *string `yaml:"host"`
	Port     *string `yaml:"port"`
	Database *string `yaml:"database"`
	User     *string `yaml:"user"`
	Password *string `yaml:"password"`
	SSLMode  *string `yaml:"sslmode"`
}

func (p *PostgresConf) ConnectionString() (conn string) {
	// conn := "postgres://postgres:webapp@localhost:5432/postgres?sslmode=disable"
	conn = "postgresql://"
	if p.User != nil {
		conn += *p.User
	}
	if p.Password != nil {
		if p.User == nil {
			logger.Print("password supplied but no user")
			// TODO: this should be an error
		}
		conn += ":" + *p.Password
	}
	if p.Host != nil {
		if p.User != nil {
			conn += "@"
		}
		conn += *p.Host
	}
	if p.Port != nil {
		conn += ":" + *p.Port
	}
	if p.Database != nil {
		conn += "/" + *p.Database
	}
	if p.SSLMode != nil {
		conn += "?sslmode=" + *p.SSLMode
	}
	return
}

func (p *PostgresConf) CheckVars() {
	if p.Host == nil {
		logger.Print("no Postgres hostname/ip address specified")
	}
	if p.Port == nil {
		p.Port = String("5432")
	}
	// Database will run driver default which is inferred as user
	// user will run driver default
}
