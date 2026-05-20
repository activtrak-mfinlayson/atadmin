package mcp

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// positionalRe matches <arg-name> patterns in a Cobra Use string.
var positionalRe = regexp.MustCompile(`<([^>]+)>`)

// mutationVerbs is the set of command names classified as mutations.
var mutationVerbs = map[string]bool{
	"update": true,
	"delete": true,
	"remove": true,
	"bulk":   true,
	"add":    true,
	"create": true,
}

// excludedTools are tool names (resource_action format) never exposed via MCP.
var excludedTools = map[string]bool{
	"users_list": true, // identity API endpoint requires elevated token scope
}

// excludedFlags are never exposed as MCP tool parameters.
var excludedFlags = map[string]bool{
	"help":     true,
	"token":    true,
	"base-url": true,
}

// ToolParam describes one MCP tool input parameter derived from a Cobra flag.
type ToolParam struct {
	Name        string
	JSONType    string // "string", "integer", or "boolean"
	Description string
}

// ToolDef is the complete MCP tool definition derived from one leaf Cobra command.
type ToolDef struct {
	Name        string
	Description string
	CommandPath []string   // e.g. ["users", "list"]
	Params      []ToolParam
	Positionals []string   // ordered positional arg names from Use, e.g. ["id"]
	HasJSONFlag bool
	IsMutation  bool
}

// Walk traverses the Cobra command tree and returns a ToolDef for every
// eligible leaf command.  When allowMutations is false, commands whose name
// matches a mutation verb are excluded.  The root command name itself is not
// included in tool names — only its children and their descendants are.
func Walk(root *cobra.Command, allowMutations bool) []ToolDef {
	var defs []ToolDef
	for _, sub := range root.Commands() {
		walkCmd(sub, nil, allowMutations, &defs)
	}
	return defs
}

func walkCmd(cmd *cobra.Command, path []string, allowMutations bool, out *[]ToolDef) {
	if cmd.Hidden || cmd.Deprecated != "" {
		return
	}

	name := cmdSegment(cmd)
	if name == "mcp" || name == "help" {
		return
	}

	// Build a new slice so we don't mutate the parent's backing array.
	currentPath := make([]string, len(path)+1)
	copy(currentPath, path)
	currentPath[len(path)] = name

	subs := cmd.Commands()
	if len(subs) == 0 {
		// Leaf — build a ToolDef.
		mutation := isMutation(cmd)
		if mutation && !allowMutations {
			return
		}
		toolname := toolName(currentPath)
		if excludedTools[toolname] {
			return
		}
		positionals := extractPositionals(cmd)
		*out = append(*out, ToolDef{
			Name:        toolname,
			Description: cmdDescription(cmd),
			CommandPath: currentPath,
			Params:      extractParams(cmd, positionals),
			Positionals: positionals,
			HasJSONFlag: cmd.Flags().Lookup("json") != nil,
			IsMutation:  mutation,
		})
		return
	}

	for _, sub := range subs {
		walkCmd(sub, currentPath, allowMutations, out)
	}
}

// cmdSegment returns the primary name of a command (first word of Use).
func cmdSegment(cmd *cobra.Command) string {
	use := cmd.Use
	if idx := strings.IndexByte(use, ' '); idx > 0 {
		return use[:idx]
	}
	return use
}

// isMutation reports whether a command is a mutation verb.
func isMutation(cmd *cobra.Command) bool {
	return mutationVerbs[cmdSegment(cmd)]
}

// toolName converts a command path slice to an MCP tool name.
// Dashes within segments are replaced with underscores.
// Example: ["audit-log", "list"] → "audit_log_list"
func toolName(path []string) string {
	parts := make([]string, len(path))
	for i, s := range path {
		parts[i] = strings.ReplaceAll(strings.ToLower(s), "-", "_")
	}
	return strings.Join(parts, "_")
}

// cmdDescription returns the Short description, falling back to the first line
// of Long, then the Use string.
func cmdDescription(cmd *cobra.Command) string {
	if cmd.Short != "" {
		return cmd.Short
	}
	if cmd.Long != "" {
		if idx := strings.IndexByte(cmd.Long, '\n'); idx > 0 {
			return cmd.Long[:idx]
		}
		return cmd.Long
	}
	return cmd.Use
}

// extractPositionals parses <arg-name> tokens from cmd.Use and returns them in order.
// Example: "get <id>" → ["id"], "get <group-id> <type> <member-id>" → ["group-id","type","member-id"].
func extractPositionals(cmd *cobra.Command) []string {
	matches := positionalRe.FindAllStringSubmatch(cmd.Use, -1)
	if len(matches) == 0 {
		return nil
	}
	names := make([]string, len(matches))
	for i, m := range matches {
		names[i] = m[1]
	}
	return names
}

// extractParams collects all non-excluded flags from a command's own FlagSet
// and its inherited persistent flags, then appends any positional args.
func extractParams(cmd *cobra.Command, positionals []string) []ToolParam {
	seen := map[string]bool{}
	var params []ToolParam

	visit := func(f *pflag.Flag) {
		if seen[f.Name] || excludedFlags[f.Name] {
			return
		}
		seen[f.Name] = true
		params = append(params, ToolParam{
			Name:        f.Name,
			JSONType:    pflagTypeToJSONType(f.Value.Type()),
			Description: f.Usage,
		})
	}

	cmd.Flags().VisitAll(visit)
	cmd.InheritedFlags().VisitAll(visit)

	for _, pos := range positionals {
		if !seen[pos] {
			seen[pos] = true
			params = append(params, ToolParam{
				Name:        pos,
				JSONType:    "string",
				Description: pos,
			})
		}
	}
	return params
}

// pflagTypeToJSONType maps pflag type strings to JSON Schema primitive types.
func pflagTypeToJSONType(t string) string {
	switch t {
	case "bool":
		return "boolean"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	default:
		return "string"
	}
}

// BuildInputSchema constructs a raw JSON Schema object from a slice of ToolParams.
func BuildInputSchema(params []ToolParam) json.RawMessage {
	type prop struct {
		Type        string `json:"type"`
		Description string `json:"description,omitempty"`
	}
	properties := make(map[string]prop, len(params))
	for _, p := range params {
		properties[p.Name] = prop{
			Type:        p.JSONType,
			Description: p.Description,
		}
	}
	schema := map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   []string{},
	}
	b, _ := json.Marshal(schema)
	return b
}
