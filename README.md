# Blurred Zed Themes

A small collection of blurred hybrid themes for the Zed editor, inspired by Neovim palettes and adapted to Zed’s modern UI.

These themes are created by overriding and extending existing Zed themes, then carefully tuning surfaces, blur, contrast, and syntax highlighting to match a modern, minimal workflow.

## Themes

<details>
<summary><strong>Evergarden Winter Green (Hybrid)</strong></summary>

Based on Evergarden / Everviolet (Neovim).

Source / inspiration:

- Evergarden (Everviolet): https://github.com/everviolet/nvim

Preview:
<img width="1941" height="1099" alt="Evergarden Winter Green (Hybrid)" src="https://github.com/user-attachments/assets/a551c81f-73b1-4aec-a0f8-476ff8aefbac" />

</details>

<details>
<summary><strong>Rosé Pine Dawn (Hybrid)</strong></summary>

Inspired by Rosé Pine Dawn (Zed) with a consistent blurred UI and stable editor background.

Source / inspiration:

- Rosé Pine (Zed): https://github.com/rose-pine/zed

Preview:
<img width="1878" height="1090" alt="Rosé Pine Dawn (Hybrid)" src="https://github.com/user-attachments/assets/1113c3bd-892e-48bf-8200-1ed5105dfbf7" />

</details>

<details>
<summary><strong>Nordic (Hybrid)</strong></summary>

A cool, desaturated theme with clean separation between UI surfaces and editor content. Based on nvim theme.

Source / inspiration:

- Nordic.nvim (Neovim): https://github.com/AlexvZyl/nordic.nvim

Preview:
<img width="1509" height="887" alt="Nordic (Hybrid)" src="https://github.com/user-attachments/assets/be112f4e-6176-411a-92bf-d7659a2838d7" />

</details>

<details>
<summary><strong>Lunar (Hybrid)</strong></summary>

A Zed reinterpretation of LunarVim’s colorscheme and blurred/borderless Zed UI.

Source / inspiration:

- LunarVim Colorschemes: https://github.com/LunarVim/Colorschemes

Preview:
<img width="1504" height="882" alt="Lunar (Hybrid)" src="https://github.com/user-attachments/assets/a0e76368-8ffb-4d9b-ad9d-99bccc3884d3" />

</details>

<details>
<summary><strong>Darkearth (Hybrid)</strong></summary>

Earth tones and low saturation with a vintage-terminal feel, tuned for modern Zed UI.

Source / inspiration:

- darkearth.nvim (Neovim): https://github.com/ptdewey/darkearth-nvim

Preview:
<img width="1512" height="889" alt="Darkearth (Hybrid)" src="https://github.com/user-attachments/assets/5ae80649-35a1-44ed-be45-e3abeb62f6ec" />

</details>

<details>
<summary><strong>Cosmos (Hybrid)</strong></summary>

High-contrast, neon-leaning theme based on the classic Cosmos palette, adapted for Zed blur and surfaces.

Source / inspiration:

- Cosmos (Zed): https://github.com/nauvalazhar/cosmos

Preview:
<img width="1512" height="887" alt="Cosmos (Hybrid)" src="https://github.com/user-attachments/assets/195383d5-5f5d-449d-af62-d9a1d0f79ef3" />

</details>

<details>
<summary><strong>Miasma Fog (Hybrid)</strong></summary>

An atmospheric, low-distraction theme with heavier blur usage and muted syntax emphasis.

Source / inspiration:

- miasma.nvim: https://github.com/xero/miasma.nvim

Preview:
<img width="1510" height="894" alt="2026-01-27_20-03-52" src="https://github.com/user-attachments/assets/c0308e82-e801-418b-9f1b-c2f2692031d0" />

</details>

## Installation

1. Copy the theme JSON file into your Zed themes directory:

```bash
~/.config/zed/themes/
```

2. Restart Zed (or reload themes).
3. Select the theme in Settings → Theme.

## Notes

- These themes rely on Zed’s `"background.appearance": "blurred"` setting.
- Designed primarily for macOS, but should work on other platforms that support blur.
- Syntax highlighting is customized beyond default Zed mappings.

## Taskfile workflow

All common workflows are wrapped in `Taskfile.yml`:

```bash
task gen-all
task sync THEME=evergarden-hybrid
task sync-all
task extract THEME=evergarden-hybrid
task publish
task verify
```

Notes:

- Palettes define roles/semantic/accents/terminal, with optional `style` for `syntax` and `players`.
- `alpha` overrides can be added per theme when needed (merged over `palettes/alpha.json`).
- `overrides` are treated as derived data and can be regenerated from a reference theme.
- The generator fills missing fields with `TODO` placeholders and applies safe defaults.
- Published/reference themes live in `themes/`.

## Credits

- Evergarden / Everviolet (Neovim): https://github.com/everviolet/nvim
- Rosé Pine (Zed): https://github.com/rose-pine/zed
- Nordic.nvim (Neovim): https://github.com/AlexvZyl/nordic.nvim
- LunarVim Colorschemes: https://github.com/LunarVim/Colorschemes
- darkearth.nvim (Neovim): https://github.com/ptdewey/darkearth-nvim
- miasma.nvim (Neovim): https://github.com/xero/miasma.nvim
- Cosmos (Zed): https://github.com/nauvalazhar/cosmos
- Catppuccin (Zed): https://github.com/catppuccin/zed
- Zed Editor: https://zed.dev

## Contributing & Feedback

Feel free to open issues or discussions if you have suggestions, improvements, or ideas for new themes.

This project is intentionally opinionated, but constructive feedback about readability, contrast, accessibility, or Zed-specific behavior is always welcome.

If you notice:

- incorrect or missing syntax mappings
- inconsistent blur or surface behavior
- UI elements that feel out of place in Zed
- differences from the original Neovim themes that should be closer

please report them.

Pull requests are also welcome, especially for:

- additional theme variants
- improved syntax coverage
- better alignment with upstream Neovim color schemes
- platform-specific adjustments
