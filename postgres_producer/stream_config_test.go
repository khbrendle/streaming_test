package main

import (
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestParseStreamConfigYaml(t *testing.T) {
	x := `
postgres:
  channel: "test_chan"
  minReconnectInterval: 3
  maxReconnectInterval: 15
kafka:
  topic: "test"
  key: ""
`

	var c StreamConfig

	err := yaml.Unmarshal([]byte(x), &c)
	if err != nil {
		t.Error("error unmarshalling YAML")
	}

	fmt.Printf("%+v\n", c)
	fmt.Printf("%+v\n", *c.Postgres.Channel)

}
