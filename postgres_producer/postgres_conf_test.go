package main

import "testing"

// tests based on Postgres 10
// postgresql.org/docs/10/libpq-connect.html

func TestConnectionString_0(t *testing.T) {
	pg := PostgresConf{}

	conn := pg.ConnectionString()
	expect := "postgresql://"
	if conn != expect {
		t.Errorf("malformed connection string. expected %s, got %s", expect, conn)
	}
}

func TestConnectionString_1(t *testing.T) {
	pg := PostgresConf{
		Host: String("localhost"),
	}

	conn := pg.ConnectionString()
	expect := "postgresql://localhost"
	if conn != expect {
		t.Errorf("malformed connection string. expected %s, got %s", expect, conn)
	}
}

func TestConnectionString_2(t *testing.T) {
	pg := PostgresConf{
		Host: String("localhost"),
		Port: String("5433"),
	}

	conn := pg.ConnectionString()
	expect := "postgresql://localhost:5433"
	if conn != expect {
		t.Errorf("malformed connection string. expected %s, got %s", expect, conn)
	}
}

func TestConnectionString_3(t *testing.T) {
	pg := PostgresConf{
		Host:     String("localhost"),
		Database: String("mydb"),
	}

	conn := pg.ConnectionString()
	expect := "postgresql://localhost/mydb"
	if conn != expect {
		t.Errorf("malformed connection string. expected %s, got %s", expect, conn)
	}
}

func TestConnectionString_4(t *testing.T) {
	pg := PostgresConf{
		Host: String("localhost"),
		User: String("user"),
	}

	conn := pg.ConnectionString()
	expect := "postgresql://user@localhost"
	if conn != expect {
		t.Errorf("malformed connection string. expected %s, got %s", expect, conn)
	}
}

func TestConnectionString_5(t *testing.T) {
	pg := PostgresConf{
		Host:     String("localhost"),
		User:     String("user"),
		Password: String("secret"),
	}

	conn := pg.ConnectionString()
	expect := "postgresql://user:secret@localhost"
	if conn != expect {
		t.Errorf("malformed connection string. expected %s, got %s", expect, conn)
	}
}

func TestConnectionString_6(t *testing.T) {
	pg := PostgresConf{
		Host:     String("localhost"),
		User:     String("user"),
		Database: String("mydb"),
		Password: String("secret"),
		SSLMode:  String("disable"),
	}

	conn := pg.ConnectionString()
	expect := "postgresql://user:secret@localhost/mydb?sslmode=disable"
	if conn != expect {
		t.Errorf("malformed connection string. expected %s, got %s", expect, conn)
	}
}
