// Package snapshot generates frozen core.json release manifests.
package snapshot

import (
	"encoding/json"
	"time"

	log "dappco.re/go/log"

	"dappco.re/go/scm/manifest"
)

// Snapshot is the frozen release manifest written as core.json.
type Snapshot struct {
	Schema      int                            `json:"schema"`
	Code        string                         `json:"code"`
	Name        string                         `json:"name"`
	Version     string                         `json:"version"`
	Description string                         `json:"description,omitempty"`
	Commit      string                         `json:"commit"`
	Tag         string                         `json:"tag"`
	Built       string                         `json:"built"`
	Daemons     map[string]manifest.DaemonSpec `json:"daemons,omitempty"`
	Layout      string                         `json:"layout,omitempty"`
	Slots       map[string]string              `json:"slots,omitempty"`
	Permissions *manifest.Permissions          `json:"permissions,omitempty"`
	Modules     []string                       `json:"modules,omitempty"`
}

// Generate creates a core.json snapshot from a manifest.
// The built timestamp is set to the current time.
func Generate(m *manifest.Manifest, commit, tag string) ([]byte, error) {
	return GenerateAt(m, commit, tag, time.Now().UTC())
}

// GenerateAt creates a core.json snapshot with an explicit build timestamp.
func GenerateAt(m *manifest.Manifest, commit, tag string, built time.Time) ([]byte, error) {
	if m == nil {
		return nil, log.E("snapshot", "manifest is nil", nil)
	}

	snap := Snapshot{
		Schema:      1,
		Code:        m.Code,
		Name:        m.Name,
		Version:     m.Version,
		Description: m.Description,
		Commit:      commit,
		Tag:         tag,
		Built:       built.Format(time.RFC3339),
		Daemons:     m.Daemons,
		Layout:      m.Layout,
		Slots:       m.Slots,
		Modules:     m.Modules,
	}

	if m.Permissions.Read != nil || m.Permissions.Write != nil ||
		m.Permissions.Net != nil || m.Permissions.Run != nil {
		snap.Permissions = &m.Permissions
	}

	return json.MarshalIndent(snap, "", "  ")
}
