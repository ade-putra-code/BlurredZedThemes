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

const (
	todoValue        = "TODO"
	transparentColor = "#00000000"
)

type roleMapping struct {
	key  string
	role string
}

type alphaSpec struct {
	alphaKey string
	styleKey string
	base     func(p Palette) string
	force    bool
}

var defaultRoleMappings = []roleMapping{
	{"background", "surface"},
	{"surface.background", "surface"},
	{"elevated_surface.background", "surface"},
	{"panel.overlay_background", "surface"},
	{"editor.background", "surface"},
	{"editor.gutter.background", "surface"},
	{"editor.subheader.background", "surface"},
	{"editor.active_line.background", "overlay"},
	{"editor.highlighted_line.background", "overlay"},
	{"editor.foreground", "text"},
	{"editor.line_number", "muted"},
	{"editor.active_line_number", "foam"},
	{"editor.invisible", "muted"},
	{"editor.indent_guide", "muted"},
	{"editor.indent_guide_active", "subtle"},
	{"editor.wrap_guide", "muted"},
	{"editor.active_wrap_guide", "muted"},
	{"editor.document_highlight.read_background", "foam"},
	{"editor.document_highlight.write_background", "muted"},
	{"editor.document_highlight.bracket_background", "iris"},
	{"editor.debugger_active_line.background", "rose"},
	{"drop_target.background", "text"},
	{"text", "text"},
	{"text.muted", "muted"},
	{"text.placeholder", "muted"},
	{"text.disabled", "subtle"},
	{"text.accent", "foam"},
	{"link_text.hover", "foam"},
	{"icon", "text"},
	{"icon.muted", "muted"},
	{"icon.placeholder", "muted"},
	{"icon.disabled", "subtle"},
	{"icon.accent", "foam"},
	{"border.variant", "foam"},
	{"border.focused", "foam"},
	{"border.selected", "iris"},
	{"border.disabled", "muted"},
	{"tab.active_background", "surface"},
	{"tab.active_foreground", "text"},
	{"tab.inactive_foreground", "muted"},
	{"status_bar.background", "surface"},
	{"title_bar.background", "surface"},
	{"title_bar.inactive_background", "surface"},
	{"status_bar.foreground", "text"},
	{"title_bar.foreground", "text"},
	{"element.active", "highlight_med"},
	{"element.selected", "highlight_med"},
	{"element.hover", "highlight_low"},
	{"element.disabled", "surface"},
	{"element.background", "surface"},
	{"ghost_element.active", "highlight_high"},
	{"ghost_element.selected", "highlight_high"},
	{"ghost_element.hover", "highlight_low"},
	{"ghost_element.disabled", "surface"},
	{"ghost_element.background", "surface"},
	{"minimap.thumb.background", "foam"},
	{"minimap.thumb.hover_background", "foam"},
	{"minimap.thumb.active_background", "foam"},
	{"pane.focused_border", "muted"},
	{"pane_group.border", "muted"},
	{"panel.focused_border", "muted"},
	{"panel.indent_guide", "muted"},
	{"panel.indent_guide_active", "subtle"},
	{"panel.indent_guide_hover", "foam"},
	{"scrollbar.thumb.background", "muted"},
	{"scrollbar.thumb.hover_background", "muted"},
	{"scrollbar.track.background", "surface"},
	{"scrollbar.track.border", "text"},
	{"search.match_background", "foam"},
	{"search.active_match_background", "rose"},
}

var defaultConstMappings = map[string]string{
	"border":                  transparentColor,
	"border.transparent":      transparentColor,
	"tab.inactive_background": transparentColor,
	"tab_bar.background":      transparentColor,
}

func roleOf(p Palette, name string) string {
	return stripAlpha(p.Roles[name])
}

func semanticOf(p Palette, name string) string {
	return semanticColor(p, name)
}

func terminalBaseOf(p Palette, key string) string {
	if p.Terminal == nil {
		return ""
	}
	if v, ok := p.Terminal[key]; ok {
		return stripAlpha(v)
	}
	return ""
}

var defaultAlphaSpecs = []alphaSpec{
	{"ui", "background", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"ui", "status_bar.background", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"ui", "title_bar.background", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"ui_inactive", "title_bar.inactive_background", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"surface", "surface.background", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"elevated", "elevated_surface.background", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"overlay", "panel.overlay_background", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"subheader", "editor.subheader.background", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"active_line", "editor.active_line.background", func(p Palette) string { return roleOf(p, "overlay") }, false},
	{"highlighted_line", "editor.highlighted_line.background", func(p Palette) string { return roleOf(p, "overlay") }, false},
	{"element_active", "element.active", func(p Palette) string { return roleOf(p, "highlight_med") }, false},
	{"element_selected", "element.selected", func(p Palette) string { return roleOf(p, "highlight_med") }, false},
	{"element_hover", "element.hover", func(p Palette) string { return roleOf(p, "highlight_low") }, false},
	{"element_disabled", "element.disabled", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"ghost_active", "ghost_element.active", func(p Palette) string { return roleOf(p, "highlight_high") }, false},
	{"ghost_selected", "ghost_element.selected", func(p Palette) string { return roleOf(p, "highlight_high") }, false},
	{"ghost_hover", "ghost_element.hover", func(p Palette) string { return roleOf(p, "highlight_low") }, false},
	{"ghost_disabled", "ghost_element.disabled", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"border_variant", "border.variant", func(p Palette) string { return roleOf(p, "foam") }, false},
	{"border_focused", "border.focused", func(p Palette) string { return roleOf(p, "foam") }, false},
	{"border_selected", "border.selected", func(p Palette) string { return roleOf(p, "iris") }, false},
	{"border_disabled", "border.disabled", func(p Palette) string { return roleOf(p, "muted") }, false},
	{"tab_active", "tab.active_background", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"conflict_marker", "version_control.conflict_marker.ours", func(p Palette) string { return semanticOf(p, "warning") }, false},
	{"conflict_marker", "version_control.conflict_marker.theirs", func(p Palette) string { return roleOf(p, "foam") }, false},
	{"panel_focus_border", "panel.focused_border", func(p Palette) string { return roleOf(p, "muted") }, false},
	{"panel_indent_guide", "panel.indent_guide", func(p Palette) string { return roleOf(p, "muted") }, false},
	{"panel_indent_guide_active", "panel.indent_guide_active", func(p Palette) string { return roleOf(p, "subtle") }, false},
	{"pane_focus_border", "pane.focused_border", func(p Palette) string { return roleOf(p, "muted") }, false},
	{"pane_group_border", "pane_group.border", func(p Palette) string { return roleOf(p, "muted") }, false},
	{"scrollbar_thumb", "scrollbar.thumb.background", func(p Palette) string { return roleOf(p, "muted") }, false},
	{"scrollbar_thumb_hover", "scrollbar.thumb.hover_background", func(p Palette) string { return roleOf(p, "muted") }, false},
	{"scrollbar_track", "scrollbar.track.background", func(p Palette) string { return roleOf(p, "surface") }, false},
	{"scrollbar_track_border", "scrollbar.track.border", func(p Palette) string { return roleOf(p, "text") }, false},
	{"search_match", "search.match_background", func(p Palette) string { return roleOf(p, "foam") }, false},
	{"search_active", "search.active_match_background", func(p Palette) string { return roleOf(p, "rose") }, false},
	{"debugger_line", "editor.debugger_active_line.background", func(p Palette) string { return roleOf(p, "rose") }, false},
	{"indent_guide", "editor.indent_guide", func(p Palette) string { return roleOf(p, "muted") }, false},
	{"indent_guide_active", "editor.indent_guide_active", func(p Palette) string { return roleOf(p, "subtle") }, false},
	{"wrap_guide", "editor.wrap_guide", func(p Palette) string { return roleOf(p, "muted") }, false},
	{"active_wrap_guide", "editor.active_wrap_guide", func(p Palette) string { return roleOf(p, "muted") }, false},
	{"doc_highlight_read", "editor.document_highlight.read_background", func(p Palette) string { return roleOf(p, "foam") }, false},
	{"doc_highlight_write", "editor.document_highlight.write_background", func(p Palette) string { return roleOf(p, "muted") }, false},
	{"doc_highlight_bracket", "editor.document_highlight.bracket_background", func(p Palette) string { return roleOf(p, "iris") }, false},
	{"drop_target", "drop_target.background", func(p Palette) string { return roleOf(p, "text") }, false},
	{"minimap_bg", "minimap.thumb.background", func(p Palette) string { return roleOf(p, "foam") }, false},
	{"minimap_hover", "minimap.thumb.hover_background", func(p Palette) string { return roleOf(p, "foam") }, false},
	{"minimap_active", "minimap.thumb.active_background", func(p Palette) string { return roleOf(p, "foam") }, false},
	{"terminal_background", "terminal.background", func(p Palette) string { return terminalBaseOf(p, "terminal.background") }, true},
	{"terminal_ansi_background", "terminal.ansi.background", func(p Palette) string { return terminalBaseOf(p, "terminal.ansi.background") }, true},
}

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
	applyRoleMappings(style, p)

	if p.Meta.BackgroundAppearance != "" {
		style["background.appearance"] = p.Meta.BackgroundAppearance
	}

	if len(p.Accents) > 0 {
		style["accents"] = p.Accents
	}

	mergeStringMap(style, p.Colors)

	mergeStringMap(style, p.Terminal)
	applyTerminalDims(style)
	applyAlphaSpecs(style, p, alpha)

	applyDerivedVim(style, p)
	applyDerivedPlayers(style, p, alpha)
	applyDerivedSyntax(style, p)
	mergeAny(style, p.Overrides)

	editorBg, _ := style["editor.background"].(string)
	if editorBg != "" {
		if _, ok := style["editor.gutter.background"]; !ok {
			style["editor.gutter.background"] = editorBg
		}
		if isUnset(style, "tab.active_background") {
			if alphaHex, ok := alphaValue(p.Meta.Appearance, alpha, "tab_active"); ok {
				style["tab.active_background"] = withAlpha(editorBg, alphaHex)
			} else {
				style["tab.active_background"] = editorBg
			}
		}
		setSemanticBackgrounds(style, p, alpha, editorBg)
	}

	if text, ok := style["text"].(string); ok && text != "" {
		if isUnset(style, "status_bar.foreground") {
			style["status_bar.foreground"] = text
		}
		if isUnset(style, "title_bar.foreground") {
			style["title_bar.foreground"] = text
		}
	}

	setDefault(style, "panel.background", transparentColor)
	setDefault(style, "toolbar.background", transparentColor)
	setDefault(style, "tab_bar.background", transparentColor)
	setDefault(style, "tab.inactive_background", transparentColor)
	setDefault(style, "border", transparentColor)
	setDefault(style, "border.transparent", transparentColor)

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
	if v, ok := alphaValue(appearance, cfg, key); ok {
		return v
	}
	return ""
}

func withAlpha(hex string, alpha string) string {
	return themeutil.WithAlpha(hex, alpha)
}

func stripAlpha(hex string) string {
	h := strings.TrimPrefix(hex, "#")
	if len(h) == 8 {
		return "#" + h[:6]
	}
	if strings.HasPrefix(hex, "#") {
		return hex
	}
	return hex
}

func hasAlpha(hex string) bool {
	h := strings.TrimPrefix(hex, "#")
	return len(h) == 8
}

func isTodoValue(value any) bool {
	s, ok := value.(string)
	return ok && s == todoValue
}

func isUnset(style map[string]any, key string) bool {
	value, ok := style[key]
	if !ok {
		return true
	}
	return isTodoValue(value)
}

func hasValue(style map[string]any, key string) bool {
	value, ok := style[key]
	if !ok {
		return false
	}
	return !isTodoValue(value)
}

func setDefault(style map[string]any, key, value string) {
	if !isUnset(style, key) {
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
		if isTodoValue(v) {
			delete(style, k)
		}
	}
}

func alphaValue(appearance string, cfg AlphaConfig, key string) (string, bool) {
	if strings.EqualFold(appearance, "light") {
		if v := cfg.Light[key]; v != "" {
			return v, true
		}
		return "", false
	}
	if v := cfg.Dark[key]; v != "" {
		return v, true
	}
	return "", false
}

func applyAlphaSpecs(style map[string]any, p Palette, alpha AlphaConfig) {
	appearance := p.Meta.Appearance
	for _, spec := range alphaSpecs() {
		alphaHex, ok := alphaValue(appearance, alpha, spec.alphaKey)
		if !ok {
			continue
		}
		base := spec.base(p)
		if base == "" {
			continue
		}
		if current, ok := style[spec.styleKey].(string); ok && current != "" && !isTodoValue(current) {
			if !spec.force {
				if hasAlpha(current) {
					continue
				}
				if !strings.EqualFold(stripAlpha(current), stripAlpha(base)) {
					continue
				}
			}
		}
		style[spec.styleKey] = withAlpha(base, alphaHex)
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

func applyRoleMappings(style map[string]any, p Palette) {
	if len(p.Roles) == 0 {
		return
	}

	role := func(name string) string { return roleOf(p, name) }

	for _, mapping := range defaultRoleMappings {
		setRole(style, mapping.key, role(mapping.role))
	}

	for key, value := range defaultConstMappings {
		setRole(style, key, value)
	}
	setAny(style, "minimap.thumb.border", nil)

	setAny(style, "scrollbar.thumb.active_background", nil)
	setAny(style, "scrollbar.thumb.border", nil)

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
	setRole(style, "version_control.conflict_marker.ours", semantic["warning"])
	setRole(style, "version_control.conflict_marker.theirs", role("foam"))

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
		if hasValue(style, bgKey) {
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
		if hasValue(style, bgKey) {
			continue
		}
		style[bgKey] = editorBg
	}
}

func setRole(style map[string]any, key, value string) {
	if value == "" {
		return
	}
	if hasValue(style, key) {
		return
	}
	style[key] = value
}

func setAny(style map[string]any, key string, value any) {
	if hasValue(style, key) {
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
			if s, ok := v.(string); ok && !isTodoValue(s) {
				continue
			}
		}
		if v, ok := style[baseKey].(string); ok {
			style[dimKey] = v
		}
	}
}

func applyDerivedVim(style map[string]any, p Palette) {
	role := func(name string) string { return roleOf(p, name) }
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
	if hasValue(style, "players") {
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
	role := func(name string) string { return roleOf(p, name) }

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

func alphaSpecs() []alphaSpec {
	return defaultAlphaSpecs
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

	specs := alphaSpecs()

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
	specs := alphaSpecs()
	seen := map[string]struct{}{}
	keys := make([]string, 0, len(specs))
	for _, spec := range specs {
		if spec.styleKey == "" {
			continue
		}
		if _, ok := seen[spec.styleKey]; ok {
			continue
		}
		seen[spec.styleKey] = struct{}{}
		keys = append(keys, spec.styleKey)
	}
	return keys
}

func semanticColor(p Palette, name string) string {
	return themeutil.SemanticColor(p.Roles, p.Semantic, name)
}

func isStandardizedKey(key string) bool {
	return themeutil.IsStandardizedKey(key)
}
