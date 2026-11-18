package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"go.yaml.in/yaml/v3"

	"github.com/gasparsuarez/repora/internal/domain"
)

var (
	topLevelKeys = map[string]struct{}{
		"version":      {},
		"repositories": {},
		"workspaces":   {},
		"preferences":  {},
	}

	repositoryKeys = map[string]struct{}{
		"id": {}, "name": {}, "path": {}, "created_at": {},
	}

	workspaceKeys = map[string]struct{}{
		"id": {}, "name": {}, "description": {}, "repositories": {},
	}

	preferenceKeys = map[string]struct{}{
		"theme": {}, "refresh_interval": {}, "status_workers": {}, "default_workspace": {},
	}
)

// Save validates cfg and writes it atomically, preserving unknown fields from
// the existing file where possible (D-014).
//
// Known sections are owned by Repora: list membership follows cfg, so a
// workspace removed in memory is removed from the file. Unknown fields inside
// repository, workspace, and preferences objects are retained by matching IDs.
func Save(path string, cfg *domain.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	current, err := readNode(path)
	if err != nil {
		return err
	}

	fresh, err := cfgNode(cfg)
	if err != nil {
		return err
	}

	merged, err := mergeMapping(current, fresh)
	if err != nil {
		return err
	}

	doc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{merged}}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("serializing configuration: %w", err)
	}

	return WriteFileAtomic(path, out, 0o600)
}

// readNode loads the current document mapping, or an empty mapping when the
// file does not exist.
func readNode(path string) (*yaml.Node, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading configuration %s: %w", path, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parsing configuration %s: %w", path, err)
	}
	if len(doc.Content) == 0 {
		return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}, nil
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("configuration %s: root is not a mapping", path)
	}
	return root, nil
}

func cfgNode(cfg *domain.Config) (*yaml.Node, error) {
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("serializing configuration: %w", err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("re-parsing configuration: %w", err)
	}
	if len(doc.Content) == 0 {
		return nil, errors.New("empty configuration document")
	}
	return doc.Content[0], nil
}

// mergeMapping returns a mapping with Repora's known keys taken from fresh and
// every other key preserved from current.
func mergeMapping(current, fresh *yaml.Node) (*yaml.Node, error) {
	if current == nil || current.Kind != yaml.MappingNode {
		return fresh, nil
	}
	if fresh == nil || fresh.Kind != yaml.MappingNode {
		return nil, errors.New("generated configuration is not a mapping")
	}

	out := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	seen := make(map[string]struct{}, len(topLevelKeys))

	// Walk the existing document first to preserve key order and unknown keys.
	for i := 0; i+1 < len(current.Content); i += 2 {
		key := current.Content[i]
		oldValue := current.Content[i+1]
		name := key.Value

		if _, known := topLevelKeys[name]; !known {
			out.Content = append(out.Content, key, oldValue)
			continue
		}

		newValue, ok := mappingValue(fresh, name)
		if !ok {
			out.Content = append(out.Content, key, oldValue)
			seen[name] = struct{}{}
			continue
		}

		merged, err := mergeKnown(name, newValue, oldValue)
		if err != nil {
			return nil, err
		}
		out.Content = append(out.Content, key, merged)
		seen[name] = struct{}{}
	}

	// Append known keys the old file did not have.
	for _, name := range []string{"version", "repositories", "workspaces", "preferences"} {
		if _, ok := seen[name]; ok {
			continue
		}
		value, ok := mappingValue(fresh, name)
		if !ok {
			continue
		}
		out.Content = append(out.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: name}, value)
	}

	return out, nil
}

func mergeKnown(name string, fresh, old *yaml.Node) (*yaml.Node, error) {
	switch name {
	case "repositories":
		return mergeSequence(fresh, old, repositoryKeys), nil
	case "workspaces":
		return mergeSequence(fresh, old, workspaceKeys), nil
	case "preferences":
		return mergeObject(fresh, old, preferenceKeys), nil
	default:
		return fresh, nil
	}
}

// mergeSequence rebuilds the list from fresh, restoring unknown fields on items
// that still exist in old (matched by id).
func mergeSequence(fresh, old *yaml.Node, known map[string]struct{}) *yaml.Node {
	if fresh == nil || fresh.Kind != yaml.SequenceNode {
		return fresh
	}

	oldByID := make(map[string]*yaml.Node)
	if old != nil && old.Kind == yaml.SequenceNode {
		for _, item := range old.Content {
			if id, ok := scalarValue(item, "id"); ok {
				oldByID[id] = item
			}
		}
	}

	out := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, item := range fresh.Content {
		if id, ok := scalarValue(item, "id"); ok {
			if previous, found := oldByID[id]; found {
				out.Content = append(out.Content, mergeObject(item, previous, known))
				continue
			}
		}
		out.Content = append(out.Content, item)
	}
	if len(out.Content) == 0 {
		out.Style = yaml.FlowStyle
	}
	return out
}

// mergeObject copies unknown keys from old into a copy of fresh.
func mergeObject(fresh, old *yaml.Node, known map[string]struct{}) *yaml.Node {
	if fresh == nil || fresh.Kind != yaml.MappingNode || old == nil || old.Kind != yaml.MappingNode {
		return fresh
	}

	out := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	out.Content = append(out.Content, fresh.Content...)

	for i := 0; i+1 < len(old.Content); i += 2 {
		name := old.Content[i].Value
		if _, isKnown := known[name]; isKnown {
			continue
		}
		if _, exists := mappingValue(out, name); exists {
			continue
		}
		out.Content = append(out.Content, old.Content[i], old.Content[i+1])
	}
	return out
}

func mappingValue(m *yaml.Node, name string) (*yaml.Node, bool) {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil, false
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == name {
			return m.Content[i+1], true
		}
	}
	return nil, false
}

func scalarValue(m *yaml.Node, name string) (string, bool) {
	value, ok := mappingValue(m, name)
	if !ok || value.Kind != yaml.ScalarNode {
		return "", false
	}
	return value.Value, true
}
