package discovery

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/vbrdnk/tmx/pkg/config"
	"github.com/vbrdnk/tmx/pkg/search"
	"github.com/vbrdnk/tmx/pkg/ui"
)

// DirectorySelector orchestrates directory discovery and selection
type DirectorySelector struct {
	searcher *search.DirectorySearcher
	config   *config.Config
}

// NewDirectorySelector creates a new DirectorySelector instance
func NewDirectorySelector(cfg *config.Config) *DirectorySelector {
	return &DirectorySelector{
		searcher: search.NewDirectorySearcher(),
		config:   cfg,
	}
}

// SelectDirectory orchestrates the full directory selection workflow
func (ds *DirectorySelector) SelectDirectory(basePath string, cliDepth int) (string, error) {
	dirList, err := ds.BuildList(basePath, cliDepth)
	if err != nil {
		return "", fmt.Errorf("error building directory list: %v", err)
	}

	selectedDir, err := ui.FuzzyFind(dirList)
	if err != nil {
		if errors.Is(err, ui.ErrNoSelection) {
			color.Yellow("No folder selected, exiting.")
			os.Exit(0)
		}
		return "", err
	}

	// Strip the frecency indicator if present
	selectedDir = strings.TrimPrefix(selectedDir, "★ ")
	return selectedDir, nil
}

// BuildList constructs a list of directories combining zoxide frecency and find/fd results
func (ds *DirectorySelector) BuildList(path string, cliDepth int) ([]byte, error) {
	searchDepth := ds.config.GetSearchDepth(cliDepth)
	useZoxide := ds.config.GetUseZoxide()

	var directories []string
	seenPaths := make(map[string]bool)

	// 1. Get zoxide results if enabled
	if useZoxide {
		zoxideResults, err := ds.searcher.QueryZoxideCache(path)
		if err == nil && len(zoxideResults) > 0 {
			for _, dir := range zoxideResults {
				if !seenPaths[dir] {
					directories = append(directories, "★ "+dir)
					seenPaths[dir] = true
				}
			}
		}
		// Silently ignore zoxide errors (not installed, no results, etc.)
	}

	// 2. Get find/fd results
	findResults, err := ds.searcher.Search(path, searchDepth)
	if err != nil {
		return nil, err
	}

	for _, dir := range findResults {
		if !seenPaths[dir] {
			directories = append(directories, dir)
			seenPaths[dir] = true
		}
	}

	return []byte(strings.Join(directories, "\n")), nil
}
