package main

import (
	"testing"
)

func TestReadValueFile(t *testing.T) {
	options := NewOptions()
	options.AddValue("APP", "commgame")
	options.AddValue("ENV", "alpha")

	options.AddFiles("service.yaml", "deployment.yaml")
	options.AddLayers(1, "deployment-tencent.yaml")
	options.selector = "tencent"
	options.config = "values.yaml"
	options.input = "./data/game"
	options.output = "./data/game_out"
	options.expand = true
	options.Process()
}

func TestLogstash(t *testing.T) {
	options := NewOptions()
	options.AddValue("APP", "word")
	options.AddValue("IMAGE", "docker.elastic.co/logstash/logstash:7.0.1")
	options.AddFiles("logstash.yaml")
	options.input = "./data/logstash"
	options.output = "./data/logstash_out"
	options.Process()
}
