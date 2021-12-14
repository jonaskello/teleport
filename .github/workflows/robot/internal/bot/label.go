/*
Copyright 2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bot

import (
	"context"
	"log"
	"strings"

	"github.com/gravitational/trace"
)

// Label parses the content of the PR (branch name, files, etc) and sets
// appropriate labels.
func (b *Bot) Label(ctx context.Context) error {
	labels, err := b.label(ctx)
	if err != nil {
		return trace.Wrap(err)
	}

	err = b.c.GitHub.AddLabels(ctx,
		b.c.Environment.Organization,
		b.c.Environment.Repository,
		b.c.Environment.Number,
		labels)
	if err != nil {
		return trace.Wrap(err)
	}

	return nil
}

func (b *Bot) label(ctx context.Context) ([]string, error) {
	// Use a map for deduplication.
	m := map[string]bool{}

	// The branch name is unsafe, but here we are simply adding a label.
	if strings.HasPrefix(b.c.Environment.UnsafeBranch, "branch/") {
		log.Printf("Label: Found backport branch.")
		m["backport"] = true
	}

	files, err := b.c.GitHub.ListFiles(ctx,
		b.c.Environment.Organization,
		b.c.Environment.Repository,
		b.c.Environment.Number)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	for _, file := range files {
		if hasDocs(file) {
			log.Printf("Label: Found documentation.")
			m["documentation"] = true
		}
		if strings.HasPrefix(file, "examples/chart/") {
			log.Printf("Label: Found Helm charts.")
			m["helm"] = true
		}
	}

	labels := make([]string, 0, len(m))
	for k, _ := range m {
		labels = append(labels, k)
	}
	return labels, nil
}
