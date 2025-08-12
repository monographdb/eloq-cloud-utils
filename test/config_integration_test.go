package config_test

import (
	"os"
	"strings"
	"testing"

	parser "github.com/monographdb/eloq-cloud-utils/parser"
)

func TestParseINIExample_SQL(t *testing.T) {
	raw, err := os.ReadFile("example/eloqsql.cnf")
	if err != nil {
		t.Fatalf("failed to read INI example: %v", err)
	}

	p, err := parser.NewConfigParser(parser.FormatINI)
	if err != nil {
		t.Fatalf("failed to create config parser: %v", err)
	}
	flat, err := p.Parse(string(raw))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// spot-check a few keys
	if flat["mariadb.port"] != "3315" {
		t.Errorf("mariadb.port expected 3315, got %q", flat["mariadb.port"])
	}
	if flat["mariadb.plugin_load_add"] != "ha_eloq" {
		t.Errorf("mariadb.plugin_load_add expected ha_eloq, got %q", flat["mariadb.plugin_load_add"])
	}

	// round-trip write should include flag-style keys without value
	out := p.UpsertConfig(map[string]string{}, flat)
	if !strings.Contains(out, "\nskip-log-bin\n") {
		t.Errorf("round-trip INI should contain 'skip-log-bin' as flag, got:\n%s", out)
	}
	if !strings.Contains(out, "\neloq\n") {
		t.Errorf("round-trip INI should contain 'eloq' as flag, got:\n%s", out)
	}
}

func TestParseINIExample_KV(t *testing.T) {
	raw, err := os.ReadFile("example/eloqkv.cnf")
	if err != nil {
		t.Fatalf("failed to read KV example: %v", err)
	}

	p, err := parser.NewConfigParser(parser.FormatINI)
	if err != nil {
		t.Fatalf("failed to create config parser: %v", err)
	}
	flat, err := p.Parse(string(raw))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if flat["local.ip"] != "127.0.0.1" {
		t.Errorf("local.ip expected 127.0.0.1, got %q", flat["local.ip"])
	}
	if flat["cluster.ip_port_list"] == "" {
		t.Errorf("cluster.ip_port_list should not be empty")
	}
}

func TestParseYAMLExample_Doc(t *testing.T) {
	raw, err := os.ReadFile("example/eloqdoc.cnf")
	if err != nil {
		t.Fatalf("failed to read YAML example: %v", err)
	}

	p, err := parser.NewConfigParser(parser.FormatYAML)
	if err != nil {
		t.Fatalf("failed to create config parser: %v", err)
	}
	flat, err := p.Parse(string(raw))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if flat["net.port"] != "27017" {
		t.Errorf("net.port expected 27017, got %q", flat["net.port"])
	}
	if flat["storage.eloq.txService.nodeMemoryLimitMB"] != "8192" {
		t.Errorf("nodeMemoryLimitMB expected 8192, got %q", flat["storage.eloq.txService.nodeMemoryLimitMB"])
	}
}

func TestUpsertAndUpdate_INI(t *testing.T) {
	p, err := parser.NewConfigParser(parser.FormatINI)
	if err != nil {
		t.Fatalf("failed to create config parser: %v", err)
	}

	old := map[string]string{
		"mariadb.port":    "3306",
		"mariadb.datadir": "/var/lib/mysql",
	}
	// upsert adds and updates
	up := map[string]string{
		"mariadb.port":         "3307",
		"mariadb.bind-address": "0.0.0.0",
	}
	outUp := p.UpsertConfig(mapCopy(old), up)
	flatUp, err := p.Parse(outUp)
	if err != nil {
		t.Fatalf("re-parse upsert failed: %v", err)
	}
	if flatUp["mariadb.port"] != "3307" || flatUp["mariadb.bind-address"] != "0.0.0.0" {
		t.Errorf("upsert did not apply changes correctly: %+v", flatUp)
	}

	// update only changes existing keys
	upd := map[string]string{
		"mariadb.port":        "3310",
		"mariadb.new-setting": "X",
	}
	outUpd := p.UpdateConfig(mapCopy(old), upd)
	flatUpd, err := p.Parse(outUpd)
	if err != nil {
		t.Fatalf("re-parse update failed: %v", err)
	}
	if flatUpd["mariadb.port"] != "3310" {
		t.Errorf("expected mariadb.port=3310, got %q", flatUpd["mariadb.port"])
	}
	if _, ok := flatUpd["mariadb.new-setting"]; ok {
		t.Errorf("update should not add new keys")
	}
}

func TestUpsertAndInsert_YAML(t *testing.T) {
	p, err := parser.NewConfigParser(parser.FormatYAML)
	if err != nil {
		t.Fatalf("failed to create config parser: %v", err)
	}

	old := map[string]string{
		"net.port":               "27017",
		"storage.dbPath":         "/data/db",
		"storage.eloq.bootstrap": "false",
	}
	// upsert adds and updates
	up := map[string]string{
		"net.port":                       "28017",
		"systemLog.verbosity":            "2",
		"storage.eloq.reservedThreadNum": "8",
	}
	outUp := p.UpsertConfig(mapCopy(old), up)
	flatUp, err := p.Parse(outUp)
	if err != nil {
		t.Fatalf("re-parse upsert failed: %v", err)
	}
	if flatUp["net.port"] != "28017" || flatUp["systemLog.verbosity"] != "2" || flatUp["storage.eloq.reservedThreadNum"] != "8" {
		t.Errorf("upsert did not apply YAML changes correctly: %+v", flatUp)
	}

	// insert adds only new keys
	ins := map[string]string{
		"net.port":               "29017", // should be ignored
		"setParameter.ttl":       "true",
		"storage.eloq.bootstrap": "true", // should be ignored (exists)
	}
	outIns := p.InsertConfig(mapCopy(old), ins)
	flatIns, err := p.Parse(outIns)
	if err != nil {
		t.Fatalf("re-parse insert failed: %v", err)
	}
	if flatIns["net.port"] != "27017" {
		t.Errorf("insert should not overwrite existing net.port")
	}
	if flatIns["setParameter.ttl"] != "true" {
		t.Errorf("insert should add new key setParameter.ttl=true")
	}
}

func mapCopy(m map[string]string) map[string]string {
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}
