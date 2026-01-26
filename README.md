# Blurred Zed Themes

A small collection of **blurred hybrid themes for the Zed editor**, inspired by  
**Evergarden (Neovim)** and **Rosé Pine**, with a focus on my view of Modern UI.

These themes are created by **overriding and extending official Zed themes**, then
carefully tuning colors, surfaces, and syntax to match my personal workflow and taste.

## Included Themes

- **Evergarden Winter Green (Hybrid)**  
  Dark, soft, nature-inspired theme based on the Evergarden Neovim palette.

- **Rosé Pine Dawn (Hybrid)**  
  Light, warm, minimal theme inspired by Rosé Pine Dawn, with consistent blurred UI.

Both themes:
- use blurred panels, toolbars, tab bars, and bars consistently
- keep the editor background readable and stable
- avoid heavy borders for a flat, modern look
- include extended and carefully tuned syntax highlighting

## Installation

1. Copy the theme JSON file into your Zed themes directory:

```bash
~/.config/zed/themes/
```

2. Example path:

```text
~/.config/zed/themes/rose-pine-dawn-hybrid.json
```

3. Restart Zed (or reload themes).
4. Select the theme in **Settings → Theme**.

## Preview

### Evergarden Winter Green (Hybrid)

<img width="1941" height="1099" alt="2026-01-26_19-59-10" src="https://github.com/user-attachments/assets/a551c81f-73b1-4aec-a0f8-476ff8aefbac" />

### Rosé Pine Dawn (Hybrid)

<img width="1878" height="1090" alt="2026-01-26_20-01-30" src="https://github.com/user-attachments/assets/1113c3bd-892e-48bf-8200-1ed5105dfbf7" />

## Notes

- These themes rely on Zed’s **`background.appearance = "blurred"`** setting.
- Designed primarily for macOS, but should work on other platforms that support blur.
- Syntax highlighting is heavily customized and inspired by the original Neovim themes.

## Credits & Inspiration

- [**Evergarden (Everviolet)**](https://github.com/everviolet/nvim) — Neovim theme by *comfysage*
- [**Rosé Pine**](https://github.com/rose-pine/zed) — official Zed theme
- [**Zed Editor**](https://zed.dev) — theming system and blur support
- [**Catppuccin**](https://github.com/catppuccin/zed) — used as a reference for Zed’s theme structure and completeness, particularly for understanding how to properly cover all UI and syntax keys.

---

Feel free to open issues or discussions if you have suggestions or improvements.
