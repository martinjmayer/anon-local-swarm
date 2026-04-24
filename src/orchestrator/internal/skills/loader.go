// Package skills loads Agent Skills (SKILL.md files) from the skills/ directory
// and makes them available to the router as system prompts.
//
// ADR-007: Agent Skills format is used for all task-type system prompts.
// The loader reads SKILL.md files, parses YAML frontmatter for metadata, and
// caches the body (everything after the closing ---) as the system prompt string.
package skills

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill holds the parsed content of a SKILL.md file.
type Skill struct {
	// Name is from the YAML frontmatter `name` field.
	Name string

	// Description is from the YAML frontmatter `description` field.
	Description string

	// Body is the full markdown content after the closing --- delimiter.
	// Injected as the Ollama system prompt on task dispatch.
	Body string
}

// Loader reads and caches skills from a directory.
type Loader struct {
	dir    string
	skills map[string]*Skill // keyed by skill name
}

// NewLoader creates a Loader for the given directory.
func NewLoader(dir string) *Loader {
	return &Loader{dir: dir, skills: make(map[string]*Skill)}
}

// LoadAll walks dir looking for */SKILL.md files, parses each one, and caches
// the result by skill name. Returns the number of skills loaded and any errors
// encountered (non-fatal — partial loads are accepted).
func (l *Loader) LoadAll() (int, []error) {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return 0, []error{fmt.Errorf("skills: read dir %q: %w", l.dir, err)}
	}

	var errs []error
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(l.dir, e.Name(), "SKILL.md")
		skill, err := parseSkillFile(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("skills: parse %q: %w", path, err))
			continue
		}
		l.skills[skill.Name] = skill
		count++
	}
	return count, errs
}

// Get returns the skill by name, or nil if not found.
func (l *Loader) Get(name string) *Skill {
	return l.skills[name]
}

// All returns all loaded skills.
func (l *Loader) All() []*Skill {
	out := make([]*Skill, 0, len(l.skills))
	for _, s := range l.skills {
		out = append(out, s)
	}
	return out
}

// frontmatter holds the YAML fields we extract from the SKILL.md header.
type frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// parseSkillFile reads a SKILL.md file, splits on --- delimiters, parses the
// YAML header, and returns a Skill with Body set to the markdown after the header.
func parseSkillFile(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	header, body, err := splitFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("no valid frontmatter: %w", err)
	}

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(header), &fm); err != nil {
		return nil, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}
	if fm.Name == "" {
		return nil, fmt.Errorf("frontmatter missing required field: name")
	}

	return &Skill{
		Name:        fm.Name,
		Description: fm.Description,
		Body:        strings.TrimSpace(body),
	}, nil
}

// splitFrontmatter splits a markdown string into (yamlHeader, body) delimited
// by --- at the start and a second --- line. Returns an error if the pattern is
// not present.
func splitFrontmatter(content string) (header, body string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) == 0 || lines[0] != "---" {
		return "", "", fmt.Errorf("does not start with ---")
	}

	// Find the closing ---
	closing := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			closing = i
			break
		}
	}
	if closing == -1 {
		return "", "", fmt.Errorf("no closing --- found")
	}

	header = strings.Join(lines[1:closing], "\n")
	body = strings.Join(lines[closing+1:], "\n")
	return header, body, nil
}
