// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package migration

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"cosmossdk.io/tools/confix"
	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/transform"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// This package includes the logic to migrate the app.toml file with the
// changes introduced in Cosmos-SDK v0.50

//go:embed v0.50-app.toml
var f embed.FS

func init() {
	confix.Migrations["v0.50"] = PlanBuilder
}

// PlanBuilder is a function that returns a transformation plan for a given diff between two files.
func PlanBuilder(from *tomledit.Document, _ string) transform.Plan {
	plan := transform.Plan{}
	deletedSections := map[string]bool{}

	target, err := LoadLocalConfig()
	if err != nil {
		panic(fmt.Errorf("failed to parse file: %w. This file should have been valid", err))
	}

	diffs := confix.DiffKeys(from, target)
	for _, diff := range diffs {
		kv := diff.KV

		var step transform.Step
		keys := strings.Split(kv.Key, ".")

		if !diff.Deleted {
			switch diff.Type {
			case confix.Section:
				step = transform.Step{
					Desc: fmt.Sprintf("add %s section", kv.Key),
					T: transform.Func(func(_ context.Context, doc *tomledit.Document) error {
						americanTitle := cases.Title(language.AmericanEnglish).String(kv.Key)
						title := fmt.Sprintf("###                    %s Configuration                    ###", americanTitle)
						doc.Sections = append(doc.Sections, &tomledit.Section{
							Heading: &parser.Heading{
								Block: parser.Comments{
									strings.Repeat("#", len(title)),
									title,
									strings.Repeat("#", len(title)),
								},
								Name: keys,
							},
						})
						return nil
					}),
				}
			case confix.Mapping:
				if len(keys) == 1 { // top-level key
					step = transform.Step{
						Desc: fmt.Sprintf("add %s key", kv.Key),
						T: transform.EnsureKey(nil, &parser.KeyValue{
							Block: kv.Block,
							Name:  parser.Key{keys[0]},
							Value: parser.MustValue(kv.Value),
						}),
					}
				} else if len(keys) > 1 {
					step = transform.Step{
						Desc: fmt.Sprintf("add %s key", kv.Key),
						T: transform.EnsureKey(keys[0:len(keys)-1], &parser.KeyValue{
							Block: kv.Block,
							Name:  parser.Key{keys[len(keys)-1]},
							Value: parser.MustValue(kv.Value),
						}),
					}
				}
			default:
				panic(fmt.Errorf("unknown diff type: %s", diff.Type))
			}
		} else {
			if diff.Type == confix.Section {
				deletedSections[kv.Key] = true
				step = transform.Step{
					Desc: fmt.Sprintf("remove %s section", kv.Key),
					T:    transform.Remove(keys),
				}
			} else {
				// when the whole section is deleted we don't need to remove the keys
				if len(keys) > 1 && deletedSections[keys[0]] {
					continue
				}

				step = transform.Step{
					Desc: fmt.Sprintf("remove %s key", kv.Key),
					T:    transform.Remove(keys),
				}
			}
		}

		plan = append(plan, step)
	}

	return plan
}

// LoadConfig loads and parses the TOML document from confix data
func LoadLocalConfig() (*tomledit.Document, error) {
	f, err := f.Open("v0.50-app.toml")
	if err != nil {
		panic(fmt.Errorf("failed to read file: %w. This file should have been included in confix", err))
	}
	defer f.Close()

	return tomledit.Parse(f)
}
