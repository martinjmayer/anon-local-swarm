package skills_test

import (
	"os"
	"path/filepath"
	"testing"

	"als/orchestrator/internal/skills"
)

func writeSkill(t *testing.T, dir, name, content string) {
	t.Helper()
	skillDir := filepath.Join(dir, name)
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
}

// ADR-007: Loader parses name, description, and body from SKILL.md.
func TestLoaderLoadAll_ADR007_ParsesFrontmatter(t *testing.T) {
	dir := t.TempDir()
	writeSkill(t, dir, "als-decompose", `---
name: als-decompose
description: Decomposes a seed task into atomic subtasks
---
# ALS Decompose

You are a task decomposer. Break the goal into atomic tasks.
Use NEW_TASK: prefix for each subtask.
`)

	l := skills.NewLoader(dir)
	count, errs := l.LoadAll()
	if len(errs) > 0 {
		t.Fatalf("LoadAll errors: %v", errs)
	}
	if count != 1 {
		t.Fatalf("want 1 skill loaded, got %d", count)
	}

	s := l.Get("als-decompose")
	if s == nil {
		t.Fatal("skill als-decompose not found")
	}
	if s.Description != "Decomposes a seed task into atomic subtasks" {
		t.Errorf("wrong description: %q", s.Description)
	}
	if s.Body == "" {
		t.Error("body must not be empty")
	}
}

// ADR-007: Skill body is injected as the system prompt — must contain the markdown content.
func TestLoaderLoadAll_ADR007_BodyIsSystemPrompt(t *testing.T) {
	dir := t.TempDir()
	writeSkill(t, dir, "als-research", `---
name: als-research
description: Research skill
---
You are a research agent. Use MCP tools to gather information.
`)

	l := skills.NewLoader(dir)
	l.LoadAll()

	s := l.Get("als-research")
	if s == nil {
		t.Fatal("skill not found")
	}
	if !contains(s.Body, "research agent") {
		t.Errorf("body does not contain expected content: %q", s.Body)
	}
}

// Multiple skills are all loaded.
func TestLoaderLoadAll_MultipleSkills(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"skill-a", "skill-b", "skill-c"} {
		writeSkill(t, dir, name, "---\nname: "+name+"\ndescription: test\n---\nbody\n")
	}

	l := skills.NewLoader(dir)
	count, errs := l.LoadAll()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if count != 3 {
		t.Errorf("want 3 skills, got %d", count)
	}
}

// Missing SKILL.md in a subdirectory produces an error, not a panic.
func TestLoaderLoadAll_MissingSkillFile(t *testing.T) {
	dir := t.TempDir()
	// Create a subdirectory with no SKILL.md
	os.MkdirAll(filepath.Join(dir, "broken-skill"), 0755)

	l := skills.NewLoader(dir)
	count, errs := l.LoadAll()
	if count != 0 {
		t.Errorf("want 0 skills loaded (all broken), got %d", count)
	}
	if len(errs) == 0 {
		t.Error("want at least 1 error for missing SKILL.md")
	}
}

// Skill missing `name` in frontmatter is rejected.
func TestLoaderLoadAll_MissingNameField(t *testing.T) {
	dir := t.TempDir()
	writeSkill(t, dir, "nameless", "---\ndescription: no name here\n---\nbody\n")

	l := skills.NewLoader(dir)
	_, errs := l.LoadAll()
	if len(errs) == 0 {
		t.Error("want error for missing name field")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
