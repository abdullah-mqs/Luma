# Luma 📝

A beautiful, fast, and minimal note-taking app built with Wails (Go) and SolidJS.

## ✨ Features

- 📂 Vault System - Import any folder as your notes vault
- 📁 Collections - Organize notes in nested folders
- 📝 Rich Text Editor - Full-featured TipTap editor with tables, code blocks, images, and more
- 🔍 Instant Search - Search across all your notes by title or content
- 🎨 Themes - 5+ built-in themes with custom theme creator
- 🌙 Dark/Light Mode - Toggle between dark and light appearance
- 📐 Resizable Panels - Customize your workspace layout
- 💾 Auto-Save - Your notes save automatically as you type
- ⚡ Blazing Fast - Built with SolidJS and Go for instant performance
- 📦 Lightweight - Only 12MB binary, unlike Electron apps

## 🚀 Tech Stack

- Backend: Go with Wails v3 framework
- Frontend: SolidJS + Vite
- Editor: TipTap
- Styling: Tailwind CSS v4

## 📦 Installation

### Arch Linux

```bash
sudo pacman -U luma.pkg.tar.zst
```

### Ubuntu / Debian

```bash
sudo apt install ./luma.deb
```

### Fedora / RHEL

```bash
sudo dnf install ./luma.rpm
```

### Windows

Download `luma.exe` and run it.

## 🛠️ From Source

```bash
git clone https://github.com/abdullah-mqs/luma.git
cd Luma
cd frontend
pnpm install
cd ..
wails3 dev
wails3 build
wails3 task linux:create:aur
wails3 task linux:create:deb
wails3 task linux:create:rpm
wails3 build GOOS=windows GOARCH=amd64
```

## 📝 Requirements

- Go 1.24+
- Node.js + pnpm
- GTK4 + WebKitGTK 6.0 (Linux)
