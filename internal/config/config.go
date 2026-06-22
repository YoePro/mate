package config

import (
	"os"
	"strings"
)

// Config contains all application configuration.
type Config struct {
	ServerAddress string

	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string
}

// Load returns the application configuration.
// It reads defaults first and then overrides them with values from .mate.ini or mate.ini if present.
func Load() Config {
	cfg := defaultConfig()

	if loadINI(".mate.ini", &cfg) {
		return cfg
	}

	loadINI("mate.ini", &cfg)

	return cfg
}

func defaultConfig() Config {
	return Config{
		ServerAddress: ":8325",

		Neo4jURI:      "neo4j://localhost:7687",
		Neo4jUser:     "neo4j",
		Neo4jPassword: "",
	}
}

func loadINI(path string, cfg *Config) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	sections := parseINI(string(data))

	if v, ok := sectionValue(sections, "server", "address"); ok {
		cfg.ServerAddress = v
	}

	if v, ok := sectionValue(sections, "neo4j", "uri"); ok {
		cfg.Neo4jURI = v
	}

	if v, ok := sectionValue(sections, "neo4j", "user"); ok {
		cfg.Neo4jUser = v
	}

	if v, ok := sectionValue(sections, "neo4j", "password"); ok {
		cfg.Neo4jPassword = v
	}

	return true
}

type iniSections map[string]map[string]string

func parseINI(data string) iniSections {
	sections := make(iniSections)
	current := ""

	for _, rawLine := range strings.Split(data, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			sections[current] = make(map[string]string)
			continue
		}

		if current == "" {
			continue
		}

		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(line[:idx]))
		value := strings.TrimSpace(line[idx+1:])
		sections[current][key] = value
	}

	return sections
}

func sectionValue(sections iniSections, section, key string) (string, bool) {
	if sec, ok := sections[strings.ToLower(section)]; ok {
		val, ok := sec[strings.ToLower(key)]
		return val, ok
	}
	return "", false
}
