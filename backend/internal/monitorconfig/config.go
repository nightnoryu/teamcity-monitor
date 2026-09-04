// Package monitorconfig loads and validates the TeamCity monitoring domain
// configuration (config.toml): the TeamCity connection, the monitored
// projects/build configurations, and the deployment environments to display.
package monitorconfig

import (
	"bytes"
	"net/url"
	"os"
	"strings"

	"github.com/go-faster/errors"
	"github.com/pelletier/go-toml/v2"
)

// Config is the root of config.toml.
type Config struct {
	TeamCityURL  string        `toml:"teamcity_url"`
	AccessToken  string        `toml:"access_token"`
	Projects     []Project     `toml:"projects"`
	Environments []Environment `toml:"environments"`
}

// Project is a monitored TeamCity project and the builds tracked within it.
type Project struct {
	Name string `toml:"name"`
	ID   string `toml:"id"`
	// EnvironmentBranchParam is a printf template with exactly one %s
	// placeholder for the environment name. The substituted string is the
	// name of a TeamCity project parameter; its edit history in TeamCity's
	// audit log is used to show who last changed it (see monitor package).
	EnvironmentBranchParam string           `toml:"environment_branch_param"`
	MonitoredBuilds        []MonitoredBuild `toml:"monitored_builds"`
}

// MonitoredBuild is a single TeamCity build configuration tracked for a
// project under a given environment. Name is a free-form grouping/display
// key (e.g. "ru", "eu", "build"), not an enum.
type MonitoredBuild struct {
	Environment string `toml:"environment"`
	Name        string `toml:"name"`
	ID          string `toml:"id"`
}

// Environment is a deployment tier displayed on the dashboard, in the order
// declared in config.toml.
type Environment struct {
	Name  string `toml:"name"`
	Emoji string `toml:"emoji"`
}

// Load reads, decodes, and validates the config file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read config file")
	}

	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, errors.Wrap(err, "decode config file")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "validate config")
	}

	return &cfg, nil
}

// Validate checks the config for internal consistency: required fields, a
// well-formed environment_branch_param, and that every monitored build
// references a declared environment.
func (c *Config) Validate() error {
	if c.TeamCityURL == "" {
		return errors.New("teamcity_url must not be empty")
	}
	if _, err := url.ParseRequestURI(c.TeamCityURL); err != nil {
		return errors.Wrap(err, "invalid teamcity_url")
	}
	if c.AccessToken == "" {
		return errors.New("access_token must not be empty")
	}
	if len(c.Environments) == 0 {
		return errors.New("at least one environment must be configured")
	}
	if len(c.Projects) == 0 {
		return errors.New("at least one project must be configured")
	}

	envNames := make(map[string]struct{}, len(c.Environments))
	for _, env := range c.Environments {
		if env.Name == "" {
			return errors.New("environment name must not be empty")
		}
		envNames[env.Name] = struct{}{}
	}

	return c.validateProjects(envNames)
}

func (c *Config) validateProjects(envNames map[string]struct{}) error {
	projectIDs := make(map[string]struct{}, len(c.Projects))
	buildIDs := make(map[string]struct{})

	for _, project := range c.Projects {
		if err := validateProject(project, envNames, projectIDs, buildIDs); err != nil {
			return err
		}
	}

	return nil
}

func validateProject(project Project, envNames, projectIDs, buildIDs map[string]struct{}) error {
	if project.ID == "" {
		return errors.Errorf("project %q: id must not be empty", project.Name)
	}
	if _, exists := projectIDs[project.ID]; exists {
		return errors.Errorf("duplicate project id %q", project.ID)
	}
	projectIDs[project.ID] = struct{}{}

	if strings.Count(project.EnvironmentBranchParam, "%s") != 1 {
		return errors.Errorf("project %q: environment_branch_param must contain exactly one %%s placeholder", project.Name)
	}

	if len(project.MonitoredBuilds) == 0 {
		return errors.Errorf("project %q: at least one monitored build must be configured", project.Name)
	}

	for _, build := range project.MonitoredBuilds {
		if err := validateMonitoredBuild(project.Name, build, envNames, buildIDs); err != nil {
			return err
		}
	}

	return nil
}

func validateMonitoredBuild(projectName string, build MonitoredBuild, envNames, buildIDs map[string]struct{}) error {
	if build.ID == "" {
		return errors.Errorf("project %q: monitored build id must not be empty", projectName)
	}
	if _, exists := buildIDs[build.ID]; exists {
		return errors.Errorf("duplicate monitored build id %q", build.ID)
	}
	buildIDs[build.ID] = struct{}{}

	if _, ok := envNames[build.Environment]; !ok {
		return errors.Errorf(
			"project %q: monitored build %q references unknown environment %q",
			projectName, build.Name, build.Environment,
		)
	}

	return nil
}
