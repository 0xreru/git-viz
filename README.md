# git-viz

> **The simple Git exfiltration analysis tool.** A lightning-fast, zero-dependency Git reconnaissance engine and visualizer designed for post-exploitation and CTF forensics.

`git-viz` is a specialized security tool for red teamers and penetration testers. It automates the analysis of exposed or dumped `.git` directories, hunting for secrets, credentials, and **orphaned (dangling) commits** that traditional Git clients and scanners often overlook.

Instead of scrolling through endless terminal logs, `git-recon-viz` generates a **single, standalone HTML dashboard** with an interactive, dark-themed DAG graph—allowing you to visualize the entire history of a repository in seconds.

---

## Core Capabilities

* **The Orphan Hunter:** Direct object-store scanning. It iterates through `.git/objects` to find dangling commits—remnants of `git reset --hard` or deleted branches where flags and keys are often hidden.
* **Single-File Portability:** The Go engine embeds the React frontend via `//go:embed`. Execution drops one `recon.html` file; no web server, Node.js, or local Git installation is required on the target machine.
* **High-Performance Secret Scanning:** Pre-configured with 50+ patterns for AWS, GitHub, Slack, JWT, and private keys. All scans are performed locally and offline.
* **Dynamic Regex Flags:** Tailored for CTFs. Use the `-f` flag to hunt for custom flag formats (e.g., `flag{...}`, `thm{...}`, `pwn{...}`) across all commits and stashes.
* **Interactive DAG Graph:** Built with `React Flow` and `dagre`. Nodes are color-coded by severity, with dedicated icons for orphans, secrets, and stashes.

---

## Installation & Build

### **From Binaries (Recommended)**

Download the pre-compiled, zero-dependency binary for your OS from the [Releases](https://www.google.com/search?q=https://github.com/0xreru/git-viz/releases) page.

### **From Source**

Requires **Go 1.21+** and **Node.js**.

```bash
git clone https://github.com/0xreru/git-viz.git
cd git-viz

# Build for your current OS (detects Windows/Linux/Mac)
make all

```

**Cross-Compilation:**

```bash
make build-linux    # Targets bin/git-recon-linux-amd64
make build-windows  # Targets bin/git-recon-windows-amd64.exe
make build-mac      # Targets bin/git-recon-darwin-arm64

```

---

## 🚀 Quick Start

Point the binary at a dumped `.git` folder.

```bash
# Basic recon (outputs to recon.html by default)
./git-recon ./dumped-repo/.git

# Full scan with custom CTF flag regex
./git-recon -o report.html -f "(?i)flag\{.*?\}" ./dumped-repo/.git

# Stealth mode (disable secret scanner, visual graph only)
./git-recon -no-secrets ./target-repo/

```

![Preview](./docs/screenshot.png)

---

### **The Output Interface**

Open the generated `report.html` in any browser:

* **Ghost Icon (Red Node):** Orphaned/Dangling commit.
* **Shield Icon:** Commit contains detected secrets/credentials.
* **Side Panel:** Provides full patch diffs, author metadata, and highlighted secret matches.

---

## 🛠️ Use Cases

* **CTF Web Exploitation:** Instantly analyze source code exfiltrated via `git-dumper` to find flags in the `reflog` or orphaned commits.
* **External Pentesting:** Review misconfigured `.git` directories on production servers for high-entropy strings and hardcoded API keys.
* **Code Auditing:** Quickly visualize branch divergence and forgotten stashes in a clean GUI.

---

## License

Distributed under the **MIT License**. See `LICENSE` for more information.

---

## GitHub Release Notes (v1.0.0)

**Title:** `v1.0.0 - The Initial Patch`

**Changelog:**

* **Initial Release:** Complete Go/React hybrid architecture.
* **Orphan Search:** Implemented raw `.git/objects` walking for dangling commit discovery.
* **Pattern Engine:** Added 50+ high-fidelity secret patterns.
* **Smooth UI:** Dark-mode dashboard with `Lucide` iconography and `React Flow` visualization.
* **Single-File Build:** Standalone HTML generation with fully inlined CSS/JS.

**Assets:**

* `git-recon-windows-amd64.exe`
* `git-recon-linux-amd64`
