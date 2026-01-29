package ranges

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RangeMeta contains only title and description for menu display
type RangeMeta struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	FilePath    string `yaml:"-"`
}

// Action represents a single action type with its color and hands
type Action struct {
	Name  string   `yaml:"name"`
	Title string   `yaml:"title"`
	Color string   `yaml:"color"`
	Hands []string `yaml:"hands"`
}

// RangeFile represents the full YAML structure
type RangeFile struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Details     string   `yaml:"details"`
	Actions     []Action `yaml:"actions"`
}

// HandAction stores color for a hand
type HandAction struct {
	Color string
}

// LoadRangeMetas loads only title/description from all YAML files in a directory
func LoadRangeMetas(dir string) ([]RangeMeta, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	metas := make([]RangeMeta, 0, len(files))
	for _, file := range files {
		meta, err := loadMeta(file)
		if err != nil {
			continue // skip invalid files
		}
		meta.FilePath = file
		metas = append(metas, meta)
	}

	return metas, nil
}

// loadMeta loads only title and description from a YAML file
func loadMeta(path string) (RangeMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RangeMeta{}, err
	}

	var meta RangeMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return RangeMeta{}, err
	}

	return meta, nil
}

// LoadRangeMeta loads title/description from a single file (exported)
func LoadRangeMeta(path string) (RangeMeta, error) {
	return loadMeta(path)
}

// LoadRangeFile loads the complete range file (on-demand when selected)
func LoadRangeFile(path string) (*RangeFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rf RangeFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return nil, err
	}

	return &rf, nil
}

// ToHandColors converts a RangeFile to a map of hand -> color for rendering
func (rf *RangeFile) ToHandColors() map[string]string {
	colors := make(map[string]string)

	for _, action := range rf.Actions {
		for _, hand := range action.Hands {
			colors[hand] = action.Color
		}
	}

	return colors
}

// GetLegend returns the actions for building a legend
func (rf *RangeFile) GetLegend() []Action {
	return rf.Actions
}
