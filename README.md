# Blurred Zed Themes

A small collection of **blurred hybrid themes for the Zed editor**, inspired by  
**Evergarden (Neovim)** and **Rosé Pine**, with a focus on:

- consistent blur usage across the UI
- readable and expressive syntax highlighting
- subtle, modern contrast
- cohesive light & dark setups

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

### Rosé Pine Dawn (Hybrid)

<img
  width="2383"
  height="1272"
  alt="Rosé Pine Dawn Hybrid preview"
  src="https://github.com/user-attachments/assets/fa755cf3-ca31-48d8-b272-fc8207068569"
/>

### Evergarden Winter Green (Hybrid)

<img
  width="2390"
  height="1281"
  alt="Evergarden Winter Green Hybrid preview"
  src="https://github.com/user-attachments/assets/a882623e-d3a1-4f6a-acc8-52439f75d565"
/>

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
