// just-install - The simple package installer for Windows
// Copyright (C) 2020 just-install authors.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3 of the License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package registry4

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/just-install/just-install/pkg/architecture"
)

// PackageMap maps package names to their associated metadata.
type PackageMap map[string]*Package

// Registry represents a package registry.
type Registry struct {
	Packages PackageMap `json:"packages"`
	Schema   string     `json:"$schema"`
	Version  int        `json:"version"`
}

// SortedPackageNames returns the list of packages present in the registry, sorted alphabetically.
func (r *Registry) SortedPackageNames() []string {
	var keys []string

	for k := range r.Packages {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// Package represents a single package.
type Package struct {
	Installer *Installer `json:"installer"`
	SkipAudit bool       `json:"skipAudit,omitempty"`
	Version   string     `json:"version"`
}

// Installer contains information to fetch and execute the installer for a package.
type Installer struct {
	Interactive bool                   `json:"interactive,omitempty"`
	Kind        string                 `json:"kind"`
	Options     map[string]interface{} `json:"options,omitempty"`
	X86         string                 `json:"x86,omitempty"`
	X86_64      string                 `json:"x86_64,omitempty"`
}

// OptionsForArch returns the options object for the given architecture
func (i *Installer) OptionsForArch(arch string) (*Options, error) {
	if !architecture.IsValid(arch) {
		return nil, fmt.Errorf("invalid architecture: %v", arch)
	}

	// XXX: this whole endeavour would have been easier with Rust and Serde... is there a better way
	// to do this in Go?

	// Check whether we have some architecture-specific options...
	optionsKeysLUT := map[string]bool{}

	for _, arch := range architecture.Architectures() {
		_, ok := i.Options[arch]
		optionsKeysLUT[arch] = ok
	}

	hasArchitectureSpecificKeys := false
	for _, v := range optionsKeysLUT {
		if v {
			hasArchitectureSpecificKeys = true
			break
		}
	}

	// ... then make a choice
	var optionsToUnmarshal interface{}
	if hasArchitectureSpecificKeys {
		options, ok := i.Options[arch]
		if !ok {
			return nil, fmt.Errorf("could not find options for architecture %v", arch)
		}

		optionsToUnmarshal = options
	} else {
		optionsToUnmarshal = i.Options
	}

	// It appears the only way to convert a map[string]interface{} back into an object is to first
	// marshal it to bytes and then marshal those back to the target object...
	b, err := json.Marshal(optionsToUnmarshal)
	if err != nil {
		return nil, fmt.Errorf("could not perform intermediate marshal on options object: %w", err)
	}

	ret := &Options{}
	if err := json.Unmarshal(b, ret); err != nil {
		return nil, fmt.Errorf("could not unmarshal options object: %w", err)
	}

	return ret, nil
}

// Options are the options that can be used to customise the install process of a package.
type Options struct {
	Arguments   []string    `json:"arguments,omitempty"`
	Container   *Container  `json:"container,omitempty"`
	Destination string      `json:"destination,omitempty"`
	Shims       []string    `json:"shims,omitempty"`
	Shortcuts   []*Shortcut `json:"shortcuts,omitempty"`
}

// Container represents options to run an installer wrapped inside a container format.
type Container struct {
	Installer string `json:"installer"`
	Kind      string `json:"kind"`
}

// Shortcut represents a shortcut to a program that is created after an installer has run.
type Shortcut struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}
