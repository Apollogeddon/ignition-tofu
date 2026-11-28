// Command docdefaults adds each resource attribute's default to the reference pages tfplugindocs writes in docs/:
// the schema OpenTofu exports has no defaults, so tfplugindocs can't show them. It reads them from the provider's
// own schemas and appends "Defaults to `x`." to each attribute's line, unless the description already gives one.
//
// Usage, from the repository root after tfplugindocs: go run ./internal/tools/docdefaults docs
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/apollogeddon/ignition-tofu/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: docdefaults <docs directory>")
		os.Exit(2)
	}
	if err := run(context.Background(), os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "docdefaults:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, dir string) error {
	added := 0
	for _, factory := range provider.New("docs")().Resources(ctx) {
		r := factory()
		var meta resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "ignition"}, &meta)
		var sch resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &sch)
		found := map[string]string{}
		collect(ctx, "", sch.Schema.Attributes, found)
		if len(found) == 0 {
			continue
		}
		page := filepath.Join(dir, "resources", strings.TrimPrefix(meta.TypeName, "ignition_")+".md")
		n, missing, err := annotate(page, found)
		if err != nil {
			return err
		}
		if len(missing) > 0 {
			return fmt.Errorf("%s: no line for the defaults of %s", page, strings.Join(missing, ", "))
		}
		added += n
	}
	fmt.Printf("docdefaults: added %d defaults\n", added)
	return nil
}

// collect records the default of every attribute that has one, keyed by its path ("policy.retention").
func collect(ctx context.Context, prefix string, attrs map[string]schema.Attribute, out map[string]string) {
	for name, attr := range attrs {
		key := prefix + name
		switch a := attr.(type) {
		case schema.StringAttribute:
			if a.Default != nil {
				var resp defaults.StringResponse
				a.Default.DefaultString(ctx, defaults.StringRequest{}, &resp)
				out[key] = strconv.Quote(resp.PlanValue.ValueString())
			}
		case schema.BoolAttribute:
			if a.Default != nil {
				var resp defaults.BoolResponse
				a.Default.DefaultBool(ctx, defaults.BoolRequest{}, &resp)
				out[key] = strconv.FormatBool(resp.PlanValue.ValueBool())
			}
		case schema.Int64Attribute:
			if a.Default != nil {
				var resp defaults.Int64Response
				a.Default.DefaultInt64(ctx, defaults.Int64Request{}, &resp)
				out[key] = strconv.FormatInt(resp.PlanValue.ValueInt64(), 10)
			}
		case schema.Float64Attribute:
			if a.Default != nil {
				var resp defaults.Float64Response
				a.Default.DefaultFloat64(ctx, defaults.Float64Request{}, &resp)
				out[key] = strconv.FormatFloat(resp.PlanValue.ValueFloat64(), 'f', -1, 64)
			}
		case schema.SingleNestedAttribute:
			collect(ctx, key+".", a.Attributes, out)
		case schema.ListNestedAttribute:
			collect(ctx, key+".", a.NestedObject.Attributes, out)
		case schema.SetNestedAttribute:
			collect(ctx, key+".", a.NestedObject.Attributes, out)
		}
	}
}

// annotate appends the defaults to the page's attribute lines ("- `name` (Type) Description."), which
// tfplugindocs groups under "### Nested Schema for `path`" for nested attributes. It returns how many it
// added and the paths it found no line for.
func annotate(page string, found map[string]string) (int, []string, error) {
	data, err := os.ReadFile(page) //nolint:gosec // a page under the docs directory the developer names
	if err != nil {
		return 0, nil, err
	}
	lines := strings.Split(string(data), "\n")
	section, added := "", 0
	seen := map[string]bool{}
	for i, line := range lines {
		if rest, ok := strings.CutPrefix(line, "### Nested Schema for `"); ok {
			section, _, _ = strings.Cut(rest, "`")
			section += "."
			continue
		}
		rest, ok := strings.CutPrefix(line, "- `")
		if !ok {
			continue
		}
		name, _, _ := strings.Cut(rest, "`")
		value, ok := found[section+name]
		if !ok {
			continue
		}
		seen[section+name] = true
		if strings.Contains(strings.ToLower(line), "default") {
			continue
		}
		lines[i] = strings.TrimRight(line, " ") + " Defaults to `" + value + "`."
		added++
	}
	var missing []string
	for key := range found {
		if !seen[key] {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if err := os.WriteFile(page, []byte(strings.Join(lines, "\n")), 0o600); err != nil { //nolint:gosec // as above
		return 0, nil, err
	}
	return added, missing, nil
}
