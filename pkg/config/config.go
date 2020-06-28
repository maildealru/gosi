package config

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Ignore Ignore
	Groups []Group
}

type config struct {
	Ignore Ignore        `yaml:"ignore"`
	Groups map[int]Group `yaml:"groups"`
}

type Group struct {
	Name   string   `yaml:"name"`
	Prefix string   `yaml:"prefix"`
	Paths  []string `yaml:"paths"`
}

type Ignore struct {
	Dirs  []string `yaml:"dirs"`
	Files []string `yaml:"files"`
}

const (
	CfgFileName   = ".gosi.yaml"
	GoModFileName = "go.mod"

	StdGroupIdx  = 0
	ProjGroupIdx = 1

	StdGroupName  = "std"
	ProjGroupName = "proj"
)

func Parse(projDir string) (*Config, error) {
	content, err := ioutil.ReadFile(filepath.Join(projDir, CfgFileName))
	if err != nil {
		return nil, err
	}

	c := &config{}
	if err = yaml.Unmarshal(content, c); err != nil {
		return nil, err
	}

	if err = c.check(); err != nil {
		return nil, err
	}
	if err = c.addReservedGroups(projDir); err != nil {
		return nil, err
	}

	return c.makeConfig(), nil
}

func (c *config) check() error {
	for idx, group := range c.Groups {
		if idx <= 1 {
			return errors.New("idx should be >= 2. idx values 0, 1 are reserved")
		}
		if group.Name == StdGroupName || group.Name == ProjGroupName {
			return errors.Errorf("group '%s' is not allowed", group.Name)
		}
	}
	return nil
}

func (c *config) addReservedGroups(projDir string) error {
	if c.Groups == nil {
		c.Groups = make(map[int]Group, 2)
	}

	if err := c.addStdLibGroup(); err != nil {
		return err
	}
	if err := c.addProjGroup(projDir); err != nil {
		return err
	}

	return nil
}

func (c *config) addStdLibGroup() error {
	list, err := packages.Load(nil, "std")
	if err != nil {
		return err
	}

	g := Group{
		Name:  StdGroupName,
		Paths: make([]string, 0),
	}
	for _, item := range list {
		path := item.PkgPath
		if strings.HasPrefix(path, "vendor/") {
			continue
		}
		g.Paths = append(g.Paths, path)
	}

	c.Groups[StdGroupIdx] = g
	return nil
}

func (c *config) addProjGroup(projDir string) error {
	fp := filepath.Join(projDir, GoModFileName)

	name, err := parseModNameFromModFile(fp)
	if err != nil {
		return err
	}

	c.Groups[ProjGroupIdx] = Group{
		Name: ProjGroupName, Prefix: name,
	}

	return nil
}

func parseModNameFromModFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()

	r := regexp.MustCompile(`^\s*module\s+(\w[^\s]*)\s*$`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if groups := r.FindSubmatch(scanner.Bytes()); len(groups) == 2 {
			return string(groups[1]), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", errors.New("no module name")
}

func (c *config) makeConfig() *Config {
	cfg := &Config{
		Ignore: c.Ignore,
		Groups: make([]Group, len(c.Groups)),
	}
	for idx, group := range c.Groups {
		cfg.Groups[idx] = group
	}
	return cfg
}
