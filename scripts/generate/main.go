package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"zed-themes/scripts/themeutil"
)

const schemaURL = "https://zed.dev/schema/themes/v0.2.0.json"

type Palette struct {
	Meta      Meta              `json:"meta"`
	Roles     map[string]string `json:"roles"`
	Semantic  map[string]string `json:"semantic"`
	Accents   []string          `json:"accents"`
	Colors    map[string]string `json:"colors"`
	Terminal  map[string]string `json:"terminal"`
	Style     map[string]any    `json:"style"`
	Overrides map[string]any    `json:"overrides"`
	Alpha     AlphaConfig       `json:"alpha"`
}

type Meta struct {
	Name                 string `json:"name"`
	Author               string `json:"author"`
	Appearance           string `json:"appearance"`
	ThemeName            string `json:"theme_name"`
	BackgroundAppearance string `json:"background_appearance"`
}

type AlphaConfig = themeutil.AlphaConfig

type Config struct {
	TemplatePath     string
	PalettePath      string
	AlphaPath        string
	OutPath          string
	PruneStyle       bool
	ComparePath      string
	WriteOverrides   bool
	WriteAlpha       bool
	PruneAlpha       bool
	RewriteOverrides bool
	WIP              bool
	KeepTODOs        bool
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := parseFlags()
	if cfg.PalettePath == "" {
		return errors.New("missing --palette")
	}

	palette, template, alphaCfg, err := loadInputs(cfg)
	if err != nil {
		return err
	}

	style := buildStyle(template, palette, alphaCfg, cfg.PruneStyle)
	if !cfg.KeepTODOs {
		removeTODOs(style)
	}
	theme := buildTheme(palette, style, cfg.WIP)
	outPath := resolveOutputPath(cfg)

	if cfg.ComparePath != "" {
		if err := compareAndMaybeUpdatePalette(cfg, palette, template, alphaCfg, style); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	if err := writeJSON(outPath, theme); err != nil {
		return fmt.Errorf("write theme: %w", err)
	}
	return nil
}

func parseFlags() Config {
	cfg := Config{}
	flag.StringVar(&cfg.TemplatePath, "template", "templates/base-style.json", "path to base style template")
	flag.StringVar(&cfg.PalettePath, "palette", "", "path to palette json")
	flag.StringVar(&cfg.AlphaPath, "alpha", "palettes/alpha.json", "path to alpha config")
	flag.StringVar(&cfg.OutPath, "out", "", "output theme json path")
	flag.BoolVar(&cfg.PruneStyle, "prune", true, "drop keys not present in palette style when available")
	flag.StringVar(&cfg.ComparePath, "compare", "", "reference theme json to compare generated style against")
	flag.BoolVar(&cfg.WriteOverrides, "write-overrides", false, "update palette overrides to match reference")
	flag.BoolVar(&cfg.WriteAlpha, "write-alpha", false, "update palette alpha overrides to match reference")
	flag.BoolVar(&cfg.PruneAlpha, "prune-alpha-overrides", false, "remove alpha-derived overrides after writing alpha")
	flag.BoolVar(&cfg.RewriteOverrides, "rewrite-overrides", false, "replace overrides with only reference diffs (excluding standardized keys)")
	flag.BoolVar(&cfg.WIP, "wip", true, "append WIP suffix to names and filenames")
	flag.BoolVar(&cfg.KeepTODOs, "keep-todos", false, "keep TODO values for debugging")
	flag.Parse()
	return cfg
}

func loadInputs(cfg Config) (Palette, map[string]any, AlphaConfig, error) {
	palette, err := readJSONFile[Palette](cfg.PalettePath)
	if err != nil {
		return Palette{}, nil, AlphaConfig{}, fmt.Errorf("read palette: %w", err)
	}

	template, err := readJSONFile[map[string]any](cfg.TemplatePath)
	if err != nil {
		return Palette{}, nil, AlphaConfig{}, fmt.Errorf("read template: %w", err)
	}

	alphaCfg, err := readJSONFile[AlphaConfig](cfg.AlphaPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Palette{}, nil, AlphaConfig{}, fmt.Errorf("read alpha: %w", err)
	}
	themeutil.MergeAlphaConfig(&alphaCfg, palette.Alpha)

	return palette, template, alphaCfg, nil
}

func resolveOutputPath(cfg Config) string {
	if cfg.OutPath != "" {
		return cfg.OutPath
	}
	base := strings.TrimSuffix(filepath.Base(cfg.PalettePath), filepath.Ext(cfg.PalettePath))
	if cfg.WIP {
		return filepath.Join("generated", "themes", base+".wip.json")
	}
	return filepath.Join("generated", "themes", base+".json")
}

func buildTheme(p Palette, style map[string]any, wip bool) map[string]any {
	name := p.Meta.Name
	themeName := p.Meta.ThemeName
	if wip {
		name = withWIPSuffix(name)
		themeName = withWIPSuffix(themeName)
	}
	return map[string]any{
		"$schema": schemaURL,
		"name":    name,
		"author":  p.Meta.Author,
		"themes": []any{
			map[string]any{
				"appearance": p.Meta.Appearance,
				"name":       themeName,
				"style":      style,
			},
		},
	}
}

func buildStyle(template map[string]any, p Palette, alpha AlphaConfig, prune bool) map[string]any {
	style := map[string]any{}
	maps.Copy(style, template)

	mergeAny(style, p.Style)
	applyRoleMappings(style, p, alpha)

	if p.Meta.BackgroundAppearance != "" {
		style["background.appearance"] = p.Meta.BackgroundAppearance
	}

	if len(p.Accents) > 0 {
		style["accents"] = p.Accents
	}

	mergeStringMap(style, p.Colors)

	mergeStringMap(style, p.Terminal)
	applyTerminalDims(style)
	applyTerminalAlpha(style, p, alpha)

	applyDerivedVim(style, p)
	applyDerivedPlayers(style, p, alpha)
	applyDerivedSyntax(style, p)
	mergeAny(style, p.Overrides)

	editorBg, _ := style["editor.background"].(string)
	if editorBg != "" {
		if _, ok := style["editor.gutter.background"]; !ok {
			style["editor.gutter.background"] = editorBg
		}
		if _, ok := style["tab.active_background"]; !ok || style["tab.active_background"] == "TODO" {
			style["tab.active_background"] = withAlpha(editorBg, alphaFor(p.Meta.Appearance, alpha, "tab_active"))
		}
		setSemanticBackgrounds(style, p, alpha, editorBg)
	}

	if text, ok := style["text"].(string); ok && text != "" {
		if _, ok := style["status_bar.foreground"]; !ok || style["status_bar.foreground"] == "TODO" {
			style["status_bar.foreground"] = text
		}
		if _, ok := style["title_bar.foreground"]; !ok || style["title_bar.foreground"] == "TODO" {
			style["title_bar.foreground"] = text
		}
	}

	setDefault(style, "panel.background", "#00000000")
	setDefault(style, "toolbar.background", "#00000000")
	setDefault(style, "tab_bar.background", "#00000000")
	setDefault(style, "tab.inactive_background", "#00000000")
	setDefault(style, "border", "#00000000")
	setDefault(style, "border.transparent", "#00000000")

	if prune && shouldPruneStyle(p.Style) {
		for k := range style {
			if _, ok := p.Style[k]; !ok {
				delete(style, k)
			}
		}
	}

	return style
}

func alphaFor(appearance string, cfg AlphaConfig, key string) string {
	return themeutil.AlphaFor(appearance, cfg, key)
}

func withAlpha(hex string, alpha string) string {
	return themeutil.WithAlpha(hex, alpha)
}

func setDefault(style map[string]any, key, value string) {
	if v, ok := style[key]; ok && v != "TODO" {
		return
	}
	style[key] = value
}

func writeJSON(path string, data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func mergeStringMap(dst map[string]any, src map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}

func mergeAny(dst map[string]any, src map[string]any) {
	maps.Copy(dst, src)
}

func readJSONFile[T any](path string) (T, error) {
	var out T
	b, err := os.ReadFile(path)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return out, err
	}
	return out, nil
}

func removeTODOs(style map[string]any) {
	for k, v := range style {
		if s, ok := v.(string); ok && s == "TODO" {
			delete(style, k)
		}
	}
}

func applyTerminalAlpha(style map[string]any, p Palette, alpha AlphaConfig) {
	appearance := p.Meta.Appearance
	alphaKeys := map[string]string{
		"terminal.background":      "terminal_background",
		"terminal.ansi.background": "terminal_ansi_background",
	}
	for styleKey, alphaKey := range alphaKeys {
		alphaHex := alphaFor(appearance, alpha, alphaKey)
		if alphaHex == "" {
			continue
		}
		value, ok := style[styleKey].(string)
		if !ok || value == "" || value == "TODO" {
			continue
		}
		style[styleKey] = withAlpha(value, alphaHex)
	}
}

func withWIPSuffix(name string) string {
	if name == "" {
		return name
	}
	if strings.HasSuffix(name, " (WIP)") {
		return name
	}
	return name + " (WIP)"
}

func compareAndMaybeUpdatePalette(cfg Config, palette Palette, template map[string]any, alphaCfg AlphaConfig, generated map[string]any) error {
	reference, err := readThemeStyle(cfg.ComparePath)
	if err != nil {
		return fmt.Errorf("read reference theme: %w", err)
	}

	missing, extra, diffs := diffStyle(reference, generated)
	fmt.Printf("compare %s\n", cfg.ComparePath)
	fmt.Printf("  missing in generated: %d\n", len(missing))
	fmt.Printf("  extra in generated: %d\n", len(extra))
	fmt.Printf("  value diffs: %d\n", len(diffs))

	if !cfg.WriteOverrides && !cfg.WriteAlpha && !cfg.PruneAlpha {
		return nil
	}

	updated := palette
	if cfg.WriteOverrides {
		if updated.Overrides == nil {
			updated.Overrides = map[string]any{}
		}
		if cfg.RewriteOverrides {
			updated.Overrides = map[string]any{}
		}
		if updated.Style == nil {
			updated.Style = map[string]any{}
		}
		for _, key := range missing {
			if key == "syntax" || key == "players" {
				updated.Style[key] = reference[key]
				continue
			}
			if isStandardizedKey(key) {
				continue
			}
			updated.Overrides[key] = reference[key]
		}
		for _, key := range diffs {
			if key == "syntax" || key == "players" {
				updated.Style[key] = reference[key]
				continue
			}
			if isStandardizedKey(key) {
				continue
			}
			updated.Overrides[key] = reference[key]
		}
	}

	if cfg.WriteAlpha {
		if updated.Alpha.Light == nil {
			updated.Alpha.Light = map[string]string{}
		}
		if updated.Alpha.Dark == nil {
			updated.Alpha.Dark = map[string]string{}
		}
		applyAlphaOverrides(&updated, alphaCfg, reference)
	}

	if cfg.PruneAlpha {
		pruneAlphaOverrides(&updated, template, alphaCfg, reference)
	}

	return writeJSON(cfg.PalettePath, updated)
}

func readThemeStyle(path string) (map[string]any, error) {
	theme, err := readJSONFile[map[string]any](path)
	if err != nil {
		return nil, err
	}
	themes, ok := theme["themes"].([]any)
	if !ok || len(themes) == 0 {
		return nil, errors.New("invalid theme: missing themes array")
	}
	first, ok := themes[0].(map[string]any)
	if !ok {
		return nil, errors.New("invalid theme: themes[0] not object")
	}
	style, ok := first["style"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid theme: missing style map")
	}
	return style, nil
}

func diffStyle(reference, generated map[string]any) ([]string, []string, []string) {
	var missing, extra, diffs []string
	for k := range reference {
		if _, ok := generated[k]; !ok {
			missing = append(missing, k)
		}
	}
	for k := range generated {
		if _, ok := reference[k]; !ok {
			extra = append(extra, k)
		}
	}
	for k, v := range reference {
		if gv, ok := generated[k]; ok && !valuesEqual(v, gv) {
			diffs = append(diffs, k)
		}
	}
	return missing, extra, diffs
}

func valuesEqual(a, b any) bool {
	ab, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(ab) == string(bb)
}

func applyRoleMappings(style map[string]any, p Palette, alpha AlphaConfig) {
	if len(p.Roles) == 0 {
		return
	}

	appearance := p.Meta.Appearance
	role := func(name string) string { return p.Roles[name] }

	setRole(style, "background", withAlpha(role("surface"), alphaFor(appearance, alpha, "ui")))
	setRole(style, "surface.background", withAlpha(role("surface"), alphaFor(appearance, alpha, "surface")))
	setRole(style, "elevated_surface.background", withAlpha(role("surface"), alphaFor(appearance, alpha, "elevated")))
	setRole(style, "panel.overlay_background", withAlpha(role("surface"), alphaFor(appearance, alpha, "overlay")))

	setRole(style, "editor.background", role("surface"))
	setRole(style, "editor.gutter.background", role("surface"))
	setRole(style, "editor.subheader.background", withAlpha(role("surface"), alphaFor(appearance, alpha, "subheader")))
	setRole(style, "editor.active_line.background", withAlpha(role("overlay"), alphaFor(appearance, alpha, "active_line")))
	setRole(style, "editor.highlighted_line.background", withAlpha(role("overlay"), alphaFor(appearance, alpha, "highlighted_line")))
	setRole(style, "editor.foreground", role("text"))
	setRole(style, "editor.line_number", role("muted"))
	setRole(style, "editor.active_line_number", role("foam"))
	setRole(style, "editor.invisible", role("muted"))
	setRole(style, "editor.indent_guide", withAlpha(role("muted"), alphaFor(appearance, alpha, "indent_guide")))
	setRole(style, "editor.indent_guide_active", withAlpha(role("subtle"), alphaFor(appearance, alpha, "indent_guide_active")))
	setRole(style, "editor.wrap_guide", withAlpha(role("muted"), alphaFor(appearance, alpha, "wrap_guide")))
	setRole(style, "editor.active_wrap_guide", withAlpha(role("muted"), alphaFor(appearance, alpha, "active_wrap_guide")))
	setRole(style, "editor.document_highlight.read_background", withAlpha(role("foam"), alphaFor(appearance, alpha, "doc_highlight_read")))
	setRole(style, "editor.document_highlight.write_background", withAlpha(role("muted"), alphaFor(appearance, alpha, "doc_highlight_write")))
	setRole(style, "editor.document_highlight.bracket_background", withAlpha(role("iris"), alphaFor(appearance, alpha, "doc_highlight_bracket")))
	setRole(style, "editor.debugger_active_line.background", withAlpha(role("rose"), alphaFor(appearance, alpha, "debugger_line")))

	setRole(style, "drop_target.background", withAlpha(role("text"), alphaFor(appearance, alpha, "drop_target")))

	setRole(style, "text", role("text"))
	setRole(style, "text.muted", role("muted"))
	setRole(style, "text.placeholder", role("muted"))
	setRole(style, "text.disabled", role("subtle"))
	setRole(style, "text.accent", role("foam"))
	setRole(style, "link_text.hover", role("foam"))

	setRole(style, "icon", role("text"))
	setRole(style, "icon.muted", role("muted"))
	setRole(style, "icon.placeholder", role("muted"))
	setRole(style, "icon.disabled", role("subtle"))
	setRole(style, "icon.accent", role("foam"))

	setRole(style, "border", "#00000000")
	setRole(style, "border.transparent", "#00000000")
	setRole(style, "border.variant", withAlpha(role("foam"), alphaFor(appearance, alpha, "border_variant")))
	setRole(style, "border.focused", withAlpha(role("foam"), alphaFor(appearance, alpha, "border_focused")))
	setRole(style, "border.selected", withAlpha(role("iris"), alphaFor(appearance, alpha, "border_selected")))
	setRole(style, "border.disabled", withAlpha(role("muted"), alphaFor(appearance, alpha, "border_disabled")))

	setRole(style, "tab.active_background", withAlpha(role("surface"), alphaFor(appearance, alpha, "tab_active")))
	setRole(style, "tab.inactive_background", "#00000000")
	setRole(style, "tab_bar.background", "#00000000")
	setRole(style, "tab.active_foreground", role("text"))
	setRole(style, "tab.inactive_foreground", role("muted"))

	setRole(style, "status_bar.background", withAlpha(role("surface"), alphaFor(appearance, alpha, "ui")))
	setRole(style, "title_bar.background", withAlpha(role("surface"), alphaFor(appearance, alpha, "ui")))
	setRole(style, "title_bar.inactive_background", withAlpha(role("surface"), alphaFor(appearance, alpha, "ui_inactive")))
	setRole(style, "status_bar.foreground", role("text"))
	setRole(style, "title_bar.foreground", role("text"))

	setRole(style, "element.active", withAlpha(role("highlight_med"), alphaFor(appearance, alpha, "element_active")))
	setRole(style, "element.selected", withAlpha(role("highlight_med"), alphaFor(appearance, alpha, "element_selected")))
	setRole(style, "element.hover", withAlpha(role("highlight_low"), alphaFor(appearance, alpha, "element_hover")))
	setRole(style, "element.disabled", withAlpha(role("surface"), alphaFor(appearance, alpha, "element_disabled")))
	setRole(style, "element.background", role("surface"))

	setRole(style, "ghost_element.active", withAlpha(role("highlight_high"), alphaFor(appearance, alpha, "ghost_active")))
	setRole(style, "ghost_element.selected", withAlpha(role("highlight_high"), alphaFor(appearance, alpha, "ghost_selected")))
	setRole(style, "ghost_element.hover", withAlpha(role("highlight_low"), alphaFor(appearance, alpha, "ghost_hover")))
	setRole(style, "ghost_element.disabled", withAlpha(role("surface"), alphaFor(appearance, alpha, "ghost_disabled")))
	setRole(style, "ghost_element.background", role("surface"))

	setRole(style, "minimap.thumb.background", withAlpha(role("foam"), alphaFor(appearance, alpha, "minimap_bg")))
	setRole(style, "minimap.thumb.hover_background", withAlpha(role("foam"), alphaFor(appearance, alpha, "minimap_hover")))
	setRole(style, "minimap.thumb.active_background", withAlpha(role("foam"), alphaFor(appearance, alpha, "minimap_active")))
	setAny(style, "minimap.thumb.border", nil)

	setRole(style, "pane.focused_border", withAlpha(role("muted"), alphaFor(appearance, alpha, "pane_focus_border")))
	setRole(style, "pane_group.border", withAlpha(role("muted"), alphaFor(appearance, alpha, "pane_group_border")))
	setRole(style, "panel.focused_border", withAlpha(role("muted"), alphaFor(appearance, alpha, "panel_focus_border")))
	setRole(style, "panel.indent_guide", withAlpha(role("muted"), alphaFor(appearance, alpha, "panel_indent_guide")))
	setRole(style, "panel.indent_guide_active", withAlpha(role("subtle"), alphaFor(appearance, alpha, "panel_indent_guide_active")))
	setRole(style, "panel.indent_guide_hover", role("foam"))

	setRole(style, "scrollbar.thumb.background", withAlpha(role("muted"), alphaFor(appearance, alpha, "scrollbar_thumb")))
	setRole(style, "scrollbar.thumb.hover_background", withAlpha(role("muted"), alphaFor(appearance, alpha, "scrollbar_thumb_hover")))
	setAny(style, "scrollbar.thumb.active_background", nil)
	setAny(style, "scrollbar.thumb.border", nil)
	setRole(style, "scrollbar.track.background", withAlpha(role("surface"), alphaFor(appearance, alpha, "scrollbar_track")))
	setRole(style, "scrollbar.track.border", withAlpha(role("text"), alphaFor(appearance, alpha, "scrollbar_track_border")))

	setRole(style, "search.match_background", withAlpha(role("foam"), alphaFor(appearance, alpha, "search_match")))
	setRole(style, "search.active_match_background", withAlpha(role("rose"), alphaFor(appearance, alpha, "search_active")))

	semantic := map[string]string{
		"error":       role("love"),
		"warning":     role("gold"),
		"info":        role("foam"),
		"success":     role("pine"),
		"conflict":    role("rose"),
		"created":     role("foam"),
		"deleted":     role("love"),
		"modified":    role("gold"),
		"renamed":     role("iris"),
		"hidden":      role("muted"),
		"hint":        role("subtle"),
		"ignored":     role("muted"),
		"unreachable": role("muted"),
		"predictive":  role("muted"),
	}
	for k, v := range p.Semantic {
		semantic[k] = v
	}
	for k, v := range semantic {
		if v != "" {
			setRole(style, k, v)
			setRole(style, k+".border", v)
		}
	}

	setRole(style, "version_control.added", semantic["created"])
	setRole(style, "version_control.deleted", semantic["deleted"])
	setRole(style, "version_control.modified", semantic["modified"])
	setRole(style, "version_control.renamed", semantic["renamed"])
	setRole(style, "version_control.conflict", firstNonEmpty(semantic["modified"], semantic["conflict"]))
	setRole(style, "version_control.ignored", semantic["ignored"])
	setRole(style, "version_control.conflict_marker.ours", withAlpha(semantic["warning"], alphaFor(appearance, alpha, "conflict_marker")))
	setRole(style, "version_control.conflict_marker.theirs", withAlpha(role("foam"), alphaFor(appearance, alpha, "conflict_marker")))

	setRole(style, "debugger.accent", semantic["error"])

	if len(p.Accents) == 0 {
		accents := []string{role("foam"), role("iris"), role("pine"), role("rose"), role("gold"), role("love")}
		var out []string
		for _, a := range accents {
			if a != "" {
				out = append(out, a)
			}
		}
		if len(out) > 0 {
			style["accents"] = out
		}
	}
}

func setSemanticBackgrounds(style map[string]any, p Palette, alpha AlphaConfig, editorBg string) {
	if editorBg == "" {
		return
	}

	type rule struct {
		key   string
		alpha string
	}
	rules := []rule{
		{"warning", alphaFor(p.Meta.Appearance, alpha, "semantic_bg")},
		{"info", alphaFor(p.Meta.Appearance, alpha, "semantic_bg")},
		{"success", alphaFor(p.Meta.Appearance, alpha, "semantic_bg")},
		{"unreachable", alphaFor(p.Meta.Appearance, alpha, "semantic_bg")},
		{"conflict", "26"},
		{"created", "26"},
		{"deleted", "26"},
		{"modified", "26"},
		{"renamed", "26"},
		{"ignored", "26"},
	}

	for _, r := range rules {
		bgKey := r.key + ".background"
		if v, ok := style[bgKey]; ok && v != "TODO" {
			continue
		}
		if fg, ok := style[r.key].(string); ok && fg != "" {
			style[bgKey] = withAlpha(fg, r.alpha)
			continue
		}
		style[bgKey] = editorBg
	}

	editorFallback := []string{
		"error",
		"hidden",
		"hint",
		"predictive",
	}
	for _, k := range editorFallback {
		bgKey := k + ".background"
		if v, ok := style[bgKey]; ok && v != "TODO" {
			continue
		}
		style[bgKey] = editorBg
	}
}

func setRole(style map[string]any, key, value string) {
	if value == "" {
		return
	}
	if v, ok := style[key]; ok && v != "TODO" {
		return
	}
	style[key] = value
}

func setAny(style map[string]any, key string, value any) {
	if v, ok := style[key]; ok && v != "TODO" {
		return
	}
	style[key] = value
}

func applyTerminalDims(style map[string]any) {
	dims := map[string]string{
		"terminal.ansi.dim_black":   "terminal.ansi.black",
		"terminal.ansi.dim_red":     "terminal.ansi.red",
		"terminal.ansi.dim_green":   "terminal.ansi.green",
		"terminal.ansi.dim_yellow":  "terminal.ansi.yellow",
		"terminal.ansi.dim_blue":    "terminal.ansi.blue",
		"terminal.ansi.dim_magenta": "terminal.ansi.magenta",
		"terminal.ansi.dim_cyan":    "terminal.ansi.cyan",
		"terminal.ansi.dim_white":   "terminal.ansi.white",
	}
	for dimKey, baseKey := range dims {
		if v, ok := style[dimKey]; ok {
			if s, ok := v.(string); ok && s != "TODO" {
				continue
			}
		}
		if v, ok := style[baseKey].(string); ok {
			style[dimKey] = v
		}
	}
}

func applyDerivedVim(style map[string]any, p Palette) {
	role := func(name string) string { return p.Roles[name] }
	normal := firstNonEmpty(role("foam"), role("pine"))
	insert := firstNonEmpty(role("rose"), role("gold"))
	visual := firstNonEmpty(role("iris"), role("rose"))
	replace := firstNonEmpty(role("love"), role("rose"))
	foreground := firstNonEmpty(role("base"), role("surface"), role("text"))

	setRole(style, "vim.mode.text", foreground)
	setRole(style, "vim.normal.background", normal)
	setRole(style, "vim.normal.foreground", foreground)
	setRole(style, "vim.helix_normal.background", normal)
	setRole(style, "vim.helix_normal.foreground", foreground)
	setRole(style, "vim.insert.background", insert)
	setRole(style, "vim.insert.foreground", foreground)
	setRole(style, "vim.visual.background", visual)
	setRole(style, "vim.visual.foreground", foreground)
	setRole(style, "vim.helix_select.background", visual)
	setRole(style, "vim.helix_select.foreground", foreground)
	setRole(style, "vim.visual_line.background", visual)
	setRole(style, "vim.visual_line.foreground", foreground)
	setRole(style, "vim.visual_block.background", visual)
	setRole(style, "vim.visual_block.foreground", foreground)
	setRole(style, "vim.replace.background", replace)
	setRole(style, "vim.replace.foreground", foreground)
}

func applyDerivedPlayers(style map[string]any, p Palette, alpha AlphaConfig) {
	if v, ok := style["players"]; ok && v != "TODO" {
		return
	}
	if len(p.Accents) == 0 {
		return
	}
	alphaHex := alphaFor(p.Meta.Appearance, alpha, "selection")
	if alphaHex == "" {
		alphaHex = "4D"
	}
	var players []map[string]string
	for _, c := range p.Accents {
		if c == "" {
			continue
		}
		players = append(players, map[string]string{
			"cursor":     c,
			"background": c,
			"selection":  withAlpha(c, alphaHex),
		})
	}
	if len(players) > 0 {
		style["players"] = players
	}
}

func applyDerivedSyntax(style map[string]any, p Palette) {
	if len(p.Roles) == 0 {
		return
	}
	role := func(name string) string { return p.Roles[name] }

	syntax := map[string]any{
		"text":            map[string]any{"color": role("text")},
		"comment":         map[string]any{"color": role("muted"), "font_style": "italic"},
		"punctuation":     map[string]any{"color": role("subtle")},
		"operator":        map[string]any{"color": role("subtle")},
		"keyword":         map[string]any{"color": role("pine")},
		"string":          map[string]any{"color": role("gold")},
		"number":          map[string]any{"color": role("foam")},
		"boolean":         map[string]any{"color": role("love")},
		"function":        map[string]any{"color": role("rose")},
		"type":            map[string]any{"color": role("foam")},
		"constant":        map[string]any{"color": role("foam")},
		"variable":        map[string]any{"color": role("text")},
		"property":        map[string]any{"color": role("text")},
		"tag":             map[string]any{"color": role("iris")},
		"attribute":       map[string]any{"color": role("rose")},
		"namespace":       map[string]any{"color": role("iris"), "font_style": "italic"},
		"module":          map[string]any{"color": role("iris"), "font_style": "italic"},
		"string.escape":   map[string]any{"color": role("love")},
		"string.regex":    map[string]any{"color": role("gold")},
		"string.special":  map[string]any{"color": role("pine")},
		"link_text":       map[string]any{"color": role("foam")},
		"link_uri":        map[string]any{"color": role("pine"), "font_style": "italic"},
		"emphasis":        map[string]any{"color": role("iris"), "font_style": "italic"},
		"emphasis.strong": map[string]any{"color": role("iris"), "font_weight": 700},
		"title":           map[string]any{"color": role("text"), "font_weight": 800},
	}

	if existing, ok := style["syntax"].(map[string]any); ok {
		if len(existing) >= 20 {
			style["syntax"] = existing
			return
		}
		for k, v := range existing {
			syntax[k] = v
		}
	}

	style["syntax"] = syntax
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func shouldPruneStyle(style map[string]any) bool {
	if len(style) == 0 {
		return false
	}
	if _, ok := style["syntax"]; ok {
		if len(style) == 1 {
			return false
		}
	}
	if _, ok := style["players"]; ok {
		if len(style) == 1 {
			return false
		}
		if len(style) == 2 {
			if _, ok := style["syntax"]; ok {
				return false
			}
		}
	}
	if _, ok := style["background"]; ok {
		return true
	}
	if _, ok := style["editor.background"]; ok {
		return true
	}
	if _, ok := style["text"]; ok {
		return true
	}
	if _, ok := style["terminal.foreground"]; ok {
		return true
	}
	return len(style) > 20
}

type alphaSpec struct {
	alphaKey string
	styleKey string
	base     func(p Palette) string
}

func applyAlphaOverrides(palette *Palette, base AlphaConfig, reference map[string]any) {
	appearance := strings.ToLower(palette.Meta.Appearance)
	defaults := map[string]string{}
	if appearance == "light" {
		for k, v := range base.Light {
			defaults[k] = v
		}
	} else {
		for k, v := range base.Dark {
			defaults[k] = v
		}
	}

	role := func(name string) string { return palette.Roles[name] }
	semantic := func(name string) string { return semanticColor(*palette, name) }

	specs := []alphaSpec{
		{"ui", "background", func(p Palette) string { return role("surface") }},
		{"ui_inactive", "title_bar.inactive_background", func(p Palette) string { return role("surface") }},
		{"surface", "surface.background", func(p Palette) string { return role("surface") }},
		{"elevated", "elevated_surface.background", func(p Palette) string { return role("surface") }},
		{"overlay", "panel.overlay_background", func(p Palette) string { return role("surface") }},
		{"subheader", "editor.subheader.background", func(p Palette) string { return role("surface") }},
		{"active_line", "editor.active_line.background", func(p Palette) string { return role("overlay") }},
		{"highlighted_line", "editor.highlighted_line.background", func(p Palette) string { return role("overlay") }},
		{"element_active", "element.active", func(p Palette) string { return role("highlight_med") }},
		{"element_selected", "element.selected", func(p Palette) string { return role("highlight_med") }},
		{"element_hover", "element.hover", func(p Palette) string { return role("highlight_low") }},
		{"element_disabled", "element.disabled", func(p Palette) string { return role("surface") }},
		{"ghost_active", "ghost_element.active", func(p Palette) string { return role("highlight_high") }},
		{"ghost_selected", "ghost_element.selected", func(p Palette) string { return role("highlight_high") }},
		{"ghost_hover", "ghost_element.hover", func(p Palette) string { return role("highlight_low") }},
		{"ghost_disabled", "ghost_element.disabled", func(p Palette) string { return role("surface") }},
		{"border_variant", "border.variant", func(p Palette) string { return role("foam") }},
		{"border_focused", "border.focused", func(p Palette) string { return role("foam") }},
		{"border_selected", "border.selected", func(p Palette) string { return role("iris") }},
		{"border_disabled", "border.disabled", func(p Palette) string { return role("muted") }},
		{"tab_active", "tab.active_background", func(p Palette) string { return role("surface") }},
		{"conflict_marker", "version_control.conflict_marker.ours", func(p Palette) string { return semantic("warning") }},
		{"panel_focus_border", "panel.focused_border", func(p Palette) string { return role("muted") }},
		{"panel_indent_guide", "panel.indent_guide", func(p Palette) string { return role("muted") }},
		{"panel_indent_guide_active", "panel.indent_guide_active", func(p Palette) string { return role("subtle") }},
		{"pane_focus_border", "pane.focused_border", func(p Palette) string { return role("muted") }},
		{"pane_group_border", "pane_group.border", func(p Palette) string { return role("muted") }},
		{"scrollbar_thumb", "scrollbar.thumb.background", func(p Palette) string { return role("muted") }},
		{"scrollbar_thumb_hover", "scrollbar.thumb.hover_background", func(p Palette) string { return role("muted") }},
		{"scrollbar_track", "scrollbar.track.background", func(p Palette) string { return role("surface") }},
		{"scrollbar_track_border", "scrollbar.track.border", func(p Palette) string { return role("text") }},
		{"search_match", "search.match_background", func(p Palette) string { return role("foam") }},
		{"search_active", "search.active_match_background", func(p Palette) string { return role("rose") }},
		{"debugger_line", "editor.debugger_active_line.background", func(p Palette) string { return role("rose") }},
		{"indent_guide", "editor.indent_guide", func(p Palette) string { return role("muted") }},
		{"indent_guide_active", "editor.indent_guide_active", func(p Palette) string { return role("subtle") }},
		{"wrap_guide", "editor.wrap_guide", func(p Palette) string { return role("muted") }},
		{"active_wrap_guide", "editor.active_wrap_guide", func(p Palette) string { return role("muted") }},
		{"doc_highlight_read", "editor.document_highlight.read_background", func(p Palette) string { return role("foam") }},
		{"doc_highlight_write", "editor.document_highlight.write_background", func(p Palette) string { return role("muted") }},
		{"doc_highlight_bracket", "editor.document_highlight.bracket_background", func(p Palette) string { return role("iris") }},
		{"drop_target", "drop_target.background", func(p Palette) string { return role("text") }},
		{"minimap_bg", "minimap.thumb.background", func(p Palette) string { return role("foam") }},
		{"minimap_hover", "minimap.thumb.hover_background", func(p Palette) string { return role("foam") }},
		{"minimap_active", "minimap.thumb.active_background", func(p Palette) string { return role("foam") }},
	}

	overrides := map[string]string{}
	for _, spec := range specs {
		baseColor := spec.base(*palette)
		if baseColor == "" {
			continue
		}
		refValue, ok := reference[spec.styleKey].(string)
		if !ok || refValue == "" {
			continue
		}
		alpha, ok := themeutil.InferAlpha(refValue, baseColor)
		if !ok {
			continue
		}
		if def := defaults[spec.alphaKey]; def != "" && strings.EqualFold(def, alpha) {
			continue
		}
		overrides[spec.alphaKey] = strings.ToUpper(alpha)
	}

	selectionAlpha := themeutil.InferSelectionAlpha(reference)
	if selectionAlpha != "" && !strings.EqualFold(selectionAlpha, defaults["selection"]) {
		overrides["selection"] = strings.ToUpper(selectionAlpha)
	}

	if appearance == "light" {
		for k, v := range overrides {
			palette.Alpha.Light[k] = v
		}
	} else {
		for k, v := range overrides {
			palette.Alpha.Dark[k] = v
		}
	}
}

func pruneAlphaOverrides(palette *Palette, template map[string]any, alphaCfg AlphaConfig, reference map[string]any) {
	if palette.Overrides == nil {
		return
	}
	mergedAlpha := alphaCfg
	themeutil.MergeAlphaConfig(&mergedAlpha, palette.Alpha)
	candidate := *palette
	candidate.Overrides = maps.Clone(palette.Overrides)

	alphaKeys := alphaDerivedKeys()
	for _, key := range alphaKeys {
		delete(candidate.Overrides, key)
	}
	style := buildStyle(template, candidate, mergedAlpha, false)
	for _, key := range alphaKeys {
		refValue, ok := reference[key]
		if !ok {
			continue
		}
		if genValue, ok := style[key]; ok && valuesEqual(refValue, genValue) {
			delete(palette.Overrides, key)
		}
	}
}

func alphaDerivedKeys() []string {
	return []string{
		"background",
		"surface.background",
		"elevated_surface.background",
		"panel.overlay_background",
		"editor.subheader.background",
		"editor.active_line.background",
		"editor.highlighted_line.background",
		"element.active",
		"element.selected",
		"element.hover",
		"element.disabled",
		"ghost_element.active",
		"ghost_element.selected",
		"ghost_element.hover",
		"ghost_element.disabled",
		"border.variant",
		"border.focused",
		"border.selected",
		"border.disabled",
		"tab.active_background",
		"version_control.conflict_marker.ours",
		"panel.focused_border",
		"panel.indent_guide",
		"panel.indent_guide_active",
		"pane.focused_border",
		"pane_group.border",
		"scrollbar.thumb.background",
		"scrollbar.thumb.hover_background",
		"scrollbar.track.background",
		"scrollbar.track.border",
		"search.match_background",
		"search.active_match_background",
		"editor.debugger_active_line.background",
		"editor.indent_guide",
		"editor.indent_guide_active",
		"editor.wrap_guide",
		"editor.active_wrap_guide",
		"editor.document_highlight.read_background",
		"editor.document_highlight.write_background",
		"editor.document_highlight.bracket_background",
		"drop_target.background",
		"minimap.thumb.background",
		"minimap.thumb.hover_background",
		"minimap.thumb.active_background",
	}
}

func semanticColor(p Palette, name string) string {
	return themeutil.SemanticColor(p.Roles, p.Semantic, name)
}

func isStandardizedKey(key string) bool {
	return themeutil.IsStandardizedKey(key)
}
