package docs

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
	"dappco.re/go/io"
	"dappco.re/go/scm/repos"
)

// Flag variables for sync command
var (
	docsSyncRegistryPath string
	docsSyncDryRun       bool
	docsSyncOutputDir    string
	docsSyncTarget       string
)

var docsSyncCmd = &cli.Command{
	Use: "sync",
	RunE: func(cmd *cli.Command, args []string) error {
		return runDocsSync(docsSyncRegistryPath, docsSyncOutputDir, docsSyncDryRun, docsSyncTarget)
	},
}

func init() {
	docsSyncCmd.Flags().StringVar(&docsSyncRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	docsSyncCmd.Flags().BoolVar(&docsSyncDryRun, "dry-run", false, i18n.T("cmd.docs.sync.flag.dry_run"))
	docsSyncCmd.Flags().StringVar(&docsSyncOutputDir, "output", "", i18n.T("cmd.docs.sync.flag.output"))
	docsSyncCmd.Flags().StringVar(&docsSyncTarget, "target", "php", "Target format: php (default), zensical, or gohelp")
}

// packageOutputName maps repo name to output folder name
func packageOutputName(repoName string) string {
	// core -> go (the Go framework)
	if repoName == "core" {
		return "go"
	}
	// core-admin -> admin, core-api -> api, etc.
	if strings.HasPrefix(repoName, "core-") {
		return strings.TrimPrefix(repoName, "core-")
	}
	return repoName
}

// shouldSyncRepo returns true if this repo should be synced
func shouldSyncRepo(repoName string) bool {
	// Skip core-php (it's the destination)
	if repoName == "core-php" {
		return false
	}
	// Skip template
	if repoName == "core-template" {
		return false
	}
	return true
}

func runDocsSync(registryPath string, outputDir string, dryRun bool, target string) error {
	reg, basePath, err := loadRegistry(registryPath)
	if err != nil {
		return err
	}

	switch target {
	case "zensical":
		return runZensicalSync(reg, basePath, outputDir, dryRun)
	case "gohelp":
		return runGoHelpSync(reg, basePath, outputDir, dryRun)
	default:
		return runPHPSync(reg, basePath, outputDir, dryRun)
	}
}

func runPHPSync(reg *repos.Registry, basePath string, outputDir string, dryRun bool) error {
	// Default output to core-php/docs/packages relative to registry
	if outputDir == "" {
		outputDir = filepath.Join(basePath, "core-php", "docs", "packages")
	}

	// Scan all repos for docs
	var docsInfo []RepoDocInfo
	for _, repo := range reg.List() {
		if !shouldSyncRepo(repo.Name) {
			continue
		}
		info := scanRepoDocs(repo)
		if info.HasDocs && len(info.DocsFiles) > 0 {
			docsInfo = append(docsInfo, info)
		}
	}

	if len(docsInfo) == 0 {
		cli.Text(i18n.T("cmd.docs.sync.no_docs_found"))
		return nil
	}

	cli.Print("\n%s %s\n\n", dimStyle.Render(i18n.T("cmd.docs.sync.found_label")), i18n.T("cmd.docs.sync.repos_with_docs", map[string]any{"Count": len(docsInfo)}))

	// Show what will be synced
	var totalFiles int
	for _, info := range docsInfo {
		totalFiles += len(info.DocsFiles)
		outName := packageOutputName(info.Name)
		cli.Print("  %s → %s %s\n",
			repoNameStyle.Render(info.Name),
			docsFileStyle.Render("packages/"+outName+"/"),
			dimStyle.Render(i18n.T("cmd.docs.sync.files_count", map[string]any{"Count": len(info.DocsFiles)})))

		for _, f := range info.DocsFiles {
			cli.Print("    %s\n", dimStyle.Render(f))
		}
	}

	cli.Print("\n%s %s\n",
		dimStyle.Render(i18n.Label("total")),
		i18n.T("cmd.docs.sync.total_summary", map[string]any{"Files": totalFiles, "Repos": len(docsInfo), "Output": outputDir}))

	if dryRun {
		cli.Print("\n%s\n", dimStyle.Render(i18n.T("cmd.docs.sync.dry_run_notice")))
		return nil
	}

	// Confirm
	cli.Blank()
	if !confirm(i18n.T("cmd.docs.sync.confirm")) {
		cli.Text(i18n.T("common.prompt.abort"))
		return nil
	}

	// Sync docs
	cli.Blank()
	var synced int
	for _, info := range docsInfo {
		outName := packageOutputName(info.Name)
		repoOutDir := filepath.Join(outputDir, outName)

		// Clear existing directory (recursively)
		if err := resetOutputDir(repoOutDir); err != nil {
			cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), info.Name, err)
			continue
		}

		// Copy all docs files
		docsDir := filepath.Join(info.Path, "docs")
		for _, f := range info.DocsFiles {
			src := filepath.Join(docsDir, f)
			dst := filepath.Join(repoOutDir, f)
			// Ensure parent dir
			if err := io.Local.EnsureDir(filepath.Dir(dst)); err != nil {
				cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), f, err)
				continue
			}

			if err := io.Copy(io.Local, src, io.Local, dst); err != nil {
				cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), f, err)
			}
		}

		cli.Print("  %s %s → packages/%s/\n", successStyle.Render("✓"), info.Name, outName)
		synced++
	}

	cli.Print("\n%s %s\n", successStyle.Render(i18n.T("i18n.done.sync")), i18n.T("cmd.docs.sync.synced_packages", map[string]any{"Count": synced}))

	return nil
}

// zensicalOutputName maps repo name to Zensical content section and folder.
func zensicalOutputName(repoName string) (string, string) {
	if repoName == "cli" {
		return "getting-started", ""
	}
	if repoName == "core" {
		return "cli", ""
	}
	if strings.HasPrefix(repoName, "go-") {
		return "go", repoName
	}
	if strings.HasPrefix(repoName, "core-") {
		return "php", strings.TrimPrefix(repoName, "core-")
	}
	return "go", repoName
}

// injectFrontMatter prepends Hugo front matter to markdown content if missing.
func injectFrontMatter(content []byte, title string, weight int) []byte {
	if bytes.HasPrefix(bytes.TrimSpace(content), []byte("---")) {
		return content
	}
	fm := fmt.Sprintf("---\ntitle: %q\nweight: %d\n---\n\n", title, weight)
	return append([]byte(fm), content...)
}

// titleFromFilename derives a human-readable title from a filename.
func titleFromFilename(filename string) string {
	name := strings.TrimSuffix(filepath.Base(filename), ".md")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	words := strings.Fields(name)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// copyWithFrontMatter copies a markdown file, injecting front matter if missing.
func copyWithFrontMatter(src, dst string, weight int) error {
	if err := io.Local.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	content, err := io.Local.Read(src)
	if err != nil {
		return err
	}
	title := titleFromFilename(src)
	result := injectFrontMatter([]byte(content), title, weight)
	return io.Local.Write(dst, string(result))
}

func runZensicalSync(reg *repos.Registry, basePath string, outputDir string, dryRun bool) error {
	if outputDir == "" {
		outputDir = filepath.Join(basePath, "docs-site", "docs")
	}

	var docsInfo []RepoDocInfo
	for _, repo := range reg.List() {
		if repo.Name == "core-template" || repo.Name == "core-claude" {
			continue
		}
		info := scanRepoDocs(repo)
		if info.HasDocs {
			docsInfo = append(docsInfo, info)
		}
	}

	if len(docsInfo) == 0 {
		cli.Text("No documentation found")
		return nil
	}

	cli.Print("\n  Zensical sync: %d repos with docs → %s\n\n", len(docsInfo), outputDir)

	for _, info := range docsInfo {
		section, folder := zensicalOutputName(info.Name)
		target := section
		if folder != "" {
			target = section + "/" + folder
		}
		fileCount := len(info.DocsFiles) + len(info.KBFiles)
		if info.Readme != "" {
			fileCount++
		}
		cli.Print("  %s → %s/ (%d files)\n", repoNameStyle.Render(info.Name), target, fileCount)
	}

	if dryRun {
		cli.Print("\n  Dry run — no files written\n")
		return nil
	}

	cli.Blank()
	if !confirm("Sync to Zensical docs directory?") {
		cli.Text("Aborted")
		return nil
	}

	cli.Blank()
	var synced int
repoLoop:
	for _, info := range docsInfo {
		section, folder := zensicalOutputName(info.Name)

		destDir := filepath.Join(outputDir, section)
		if folder != "" {
			destDir = filepath.Join(destDir, folder)
		}

		if err := resetOutputDir(destDir); err != nil {
			cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), info.Name, err)
			continue
		}

		weight := 10
		docsDir := filepath.Join(info.Path, "docs")
		for _, f := range info.DocsFiles {
			src := filepath.Join(docsDir, f)
			dst := filepath.Join(destDir, f)
			if err := copyWithFrontMatter(src, dst, weight); err != nil {
				cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), f, err)
				continue
			}
			weight += 10
		}

		if info.Readme != "" {
			if err := copyZensicalReadme(info.Readme, destDir); err != nil {
				cli.Print("  %s README: %s\n", errorStyle.Render("✗"), err)
			}
		}

		if len(info.KBFiles) > 0 {
			suffix := strings.TrimPrefix(info.Name, "go-")
			kbDestDir := filepath.Join(outputDir, "kb", suffix)
			if err := resetOutputDir(kbDestDir); err != nil {
				cli.Print("  %s KB: %s\n", errorStyle.Render("✗"), err)
				continue repoLoop
			}
			kbDir := filepath.Join(info.Path, "KB")
			kbWeight := 10
			for _, f := range info.KBFiles {
				src := filepath.Join(kbDir, f)
				dst := filepath.Join(kbDestDir, f)
				if err := copyWithFrontMatter(src, dst, kbWeight); err != nil {
					cli.Print("  %s KB/%s: %s\n", errorStyle.Render("✗"), f, err)
					continue
				}
				kbWeight += 10
			}
		}

		cli.Print("  %s %s\n", successStyle.Render("✓"), info.Name)
		synced++
	}

	cli.Print("\n  Synced %d repos to Zensical docs\n", synced)
	return nil
}

// copyZensicalReadme copies a repository README to index.md in the target directory.
func copyZensicalReadme(src, destDir string) error {
	dst := filepath.Join(destDir, "index.md")
	return copyWithFrontMatter(src, dst, 1)
}

// resetOutputDir clears and recreates a target directory before copying files into it.
func resetOutputDir(dir string) error {
	if err := io.Local.DeleteAll(dir); err != nil {
		return err
	}
	return io.Local.EnsureDir(dir)
}

// goHelpOutputName maps repo name to output folder name for go-help.
func goHelpOutputName(repoName string) string {
	if repoName == "core" {
		return "go"
	}
	if strings.HasPrefix(repoName, "core-") {
		return strings.TrimPrefix(repoName, "core-")
	}
	return repoName
}

func runGoHelpSync(reg *repos.Registry, basePath string, outputDir string, dryRun bool) error {
	if outputDir == "" {
		outputDir = filepath.Join(basePath, "docs", "content")
	}

	var docsInfo []RepoDocInfo
	for _, repo := range reg.List() {
		if repo.Name == "core-template" || repo.Name == "core-claude" {
			continue
		}
		info := scanRepoDocs(repo)
		if info.HasDocs && len(info.DocsFiles) > 0 {
			docsInfo = append(docsInfo, info)
		}
	}

	if len(docsInfo) == 0 {
		cli.Text("No documentation found")
		return nil
	}

	cli.Print("\n  Go-help sync: %d repos with docs → %s\n\n", len(docsInfo), outputDir)

	var totalFiles int
	for _, info := range docsInfo {
		outName := goHelpOutputName(info.Name)
		totalFiles += len(info.DocsFiles)
		cli.Print("  %s → content/%s/ (%d files)\n", repoNameStyle.Render(info.Name), outName, len(info.DocsFiles))
	}

	cli.Print("\n  %s %d files from %d repos → %s\n",
		dimStyle.Render("Total:"), totalFiles, len(docsInfo), outputDir)

	if dryRun {
		cli.Print("\n  Dry run — no files written\n")
		return nil
	}

	cli.Blank()
	if !confirm("Sync to go-help content directory?") {
		cli.Text("Aborted")
		return nil
	}

	cli.Blank()
	var synced int
	for _, info := range docsInfo {
		outName := goHelpOutputName(info.Name)
		repoOutDir := filepath.Join(outputDir, outName)

		// Clear existing directory
		if err := resetOutputDir(repoOutDir); err != nil {
			cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), info.Name, err)
			continue
		}

		// Plain copy of docs files (no frontmatter injection)
		docsDir := filepath.Join(info.Path, "docs")
		for _, f := range info.DocsFiles {
			src := filepath.Join(docsDir, f)
			dst := filepath.Join(repoOutDir, f)
			if err := io.Local.EnsureDir(filepath.Dir(dst)); err != nil {
				cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), f, err)
				continue
			}
			if err := io.Copy(io.Local, src, io.Local, dst); err != nil {
				cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), f, err)
			}
		}

		cli.Print("  %s %s → content/%s/\n", successStyle.Render("✓"), info.Name, outName)
		synced++
	}

	cli.Print("\n  Synced %d repos to go-help content\n", synced)
	return nil
}
