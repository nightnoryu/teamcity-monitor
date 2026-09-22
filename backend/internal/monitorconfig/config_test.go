package monitorconfig_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"teamcity-monitor/internal/monitorconfig"
)

const validConfig = `
teamcity_url = "https://teamcity.your-org.lan"
access_token = "abc..."

[[projects]]
name = "Alpha"
id = "Alpha_Testing"
environment_branch_param = "alpha_%s_branch"
monitored_builds = [
    { environment = "dev", name = "ru", id = "Alpha_Testing_Ru" },
    { environment = "dev", name = "eu", id = "Alpha_Testing_Eu" },
]

[[projects]]
name = "Beta"
id = "Beta_Testing"
environment_branch_param = "vcs_root_branch_%s"
monitored_builds = [
    { environment = "dev", name = "build", id = "Beta_Testing_Dev_Build" },
    { environment = "dev", name = "ru", id = "Beta_Testing_Dev_Ru" },
    { environment = "dev", name = "eu", id = "Beta_Testing_Dev_Eu" },
]

[[environments]]
name = "dev"
emoji = "🥭"

[[environments]]
name = "orange"
emoji = "🍊"
`

func writeConfig(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))

	return path
}

func TestValidate_RegressionCases(t *testing.T) {
	base := func() *monitorconfig.Config {
		return &monitorconfig.Config{
			TeamCityURL: "https://example.com/teamcity/", AccessToken: "token",
			Environments: []monitorconfig.Environment{{Name: "dev"}},
			Projects:     []monitorconfig.Project{{Name: "Alpha", ID: "A", EnvironmentBranchParam: "branch_%s", MonitoredBuilds: []monitorconfig.MonitoredBuild{{Environment: "dev", Name: "ru", ID: "A_Ru"}}}},
		}
	}
	tests := map[string]func(*monitorconfig.Config){
		"duplicate environment": func(c *monitorconfig.Config) {
			c.Environments = append(c.Environments, monitorconfig.Environment{Name: "dev"})
		},
		"blank project name":  func(c *monitorconfig.Config) { c.Projects[0].Name = "  " },
		"blank build name":    func(c *monitorconfig.Config) { c.Projects[0].MonitoredBuilds[0].Name = "  " },
		"escaped placeholder": func(c *monitorconfig.Config) { c.Projects[0].EnvironmentBranchParam = "branch_%%s" },
		"extra formatter":     func(c *monitorconfig.Config) { c.Projects[0].EnvironmentBranchParam = "branch_%s_%d" },
		"relative URL":        func(c *monitorconfig.Config) { c.TeamCityURL = "/teamcity" },
		"non HTTP URL":        func(c *monitorconfig.Config) { c.TeamCityURL = "ftp://example.com" },
		"query URL":           func(c *monitorconfig.Config) { c.TeamCityURL = "https://example.com/?x=1" },
		"empty query URL":     func(c *monitorconfig.Config) { c.TeamCityURL = "https://example.com/?" },
		"fragment URL":        func(c *monitorconfig.Config) { c.TeamCityURL = "https://example.com/#fragment" },
		"empty fragment URL":  func(c *monitorconfig.Config) { c.TeamCityURL = "https://example.com/#" },
		"missing hostname":    func(c *monitorconfig.Config) { c.TeamCityURL = "https://:8443/" },
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) { cfg := base(); change(cfg); require.Error(t, cfg.Validate()) })
	}
	for _, raw := range []string{"https://example.com", "https://example.com/teamcity/"} {
		t.Run("valid "+strings.TrimPrefix(raw, "https://"), func(t *testing.T) { cfg := base(); cfg.TeamCityURL = raw; require.NoError(t, cfg.Validate()) })
	}
}

func TestLoad_Valid(t *testing.T) {
	path := writeConfig(t, validConfig)

	cfg, err := monitorconfig.Load(path)
	require.NoError(t, err)

	require.Equal(t, "https://teamcity.your-org.lan", cfg.TeamCityURL)
	require.Len(t, cfg.Projects, 2)
	require.Len(t, cfg.Environments, 2)
	require.Equal(t, "ru", cfg.Projects[0].MonitoredBuilds[0].Name)
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := monitorconfig.Load(filepath.Join(t.TempDir(), "missing.toml"))
	require.Error(t, err)
}

func TestLoad_Invalid(t *testing.T) {
	tests := map[string]string{ //nolint:gosec // TOML fixtures, not real credentials
		"unknown field": `
teamcity_url = "https://teamcity.your-org.lan"
access_token = "abc"
unexpected_field = "oops"
[[projects]]
name = "Alpha"
id = "Alpha_Testing"
environment_branch_param = "alpha_%s_branch"
monitored_builds = [{ environment = "dev", name = "ru", id = "Alpha_Ru" }]
[[environments]]
name = "dev"
emoji = "🥭"
`,
		"missing teamcity_url": `
access_token = "abc"
[[projects]]
name = "Alpha"
id = "Alpha_Testing"
environment_branch_param = "alpha_%s_branch"
monitored_builds = [{ environment = "dev", name = "ru", id = "Alpha_Ru" }]
[[environments]]
name = "dev"
emoji = "🥭"
`,
		"missing access_token": `
teamcity_url = "https://teamcity.your-org.lan"
[[projects]]
name = "Alpha"
id = "Alpha_Testing"
environment_branch_param = "alpha_%s_branch"
monitored_builds = [{ environment = "dev", name = "ru", id = "Alpha_Ru" }]
[[environments]]
name = "dev"
emoji = "🥭"
`,
		"no environments": `
teamcity_url = "https://teamcity.your-org.lan"
access_token = "abc"
[[projects]]
name = "Alpha"
id = "Alpha_Testing"
environment_branch_param = "alpha_%s_branch"
monitored_builds = [{ environment = "dev", name = "ru", id = "Alpha_Ru" }]
`,
		"no projects": `
teamcity_url = "https://teamcity.your-org.lan"
access_token = "abc"
[[environments]]
name = "dev"
emoji = "🥭"
`,
		"branch param without placeholder": `
teamcity_url = "https://teamcity.your-org.lan"
access_token = "abc"
[[projects]]
name = "Alpha"
id = "Alpha_Testing"
environment_branch_param = "alpha_branch"
monitored_builds = [{ environment = "dev", name = "ru", id = "Alpha_Ru" }]
[[environments]]
name = "dev"
emoji = "🥭"
`,
		"branch param with two placeholders": `
teamcity_url = "https://teamcity.your-org.lan"
access_token = "abc"
[[projects]]
name = "Alpha"
id = "Alpha_Testing"
environment_branch_param = "alpha_%s_%s_branch"
monitored_builds = [{ environment = "dev", name = "ru", id = "Alpha_Ru" }]
[[environments]]
name = "dev"
emoji = "🥭"
`,
		"dangling environment reference": `
teamcity_url = "https://teamcity.your-org.lan"
access_token = "abc"
[[projects]]
name = "Alpha"
id = "Alpha_Testing"
environment_branch_param = "alpha_%s_branch"
monitored_builds = [{ environment = "staging", name = "ru", id = "Alpha_Ru" }]
[[environments]]
name = "dev"
emoji = "🥭"
`,
		"duplicate project id": `
teamcity_url = "https://teamcity.your-org.lan"
access_token = "abc"
[[projects]]
name = "Alpha"
id = "Alpha_Testing"
environment_branch_param = "alpha_%s_branch"
monitored_builds = [{ environment = "dev", name = "ru", id = "Alpha_Ru" }]
[[projects]]
name = "Alpha-dup"
id = "Alpha_Testing"
environment_branch_param = "alpha_%s_branch"
monitored_builds = [{ environment = "dev", name = "eu", id = "Alpha_Eu" }]
[[environments]]
name = "dev"
emoji = "🥭"
`,
		"duplicate build id": `
teamcity_url = "https://teamcity.your-org.lan"
access_token = "abc"
[[projects]]
name = "Alpha"
id = "Alpha_Testing"
environment_branch_param = "alpha_%s_branch"
monitored_builds = [
    { environment = "dev", name = "ru", id = "Alpha_Ru" },
    { environment = "dev", name = "eu", id = "Alpha_Ru" },
]
[[environments]]
name = "dev"
emoji = "🥭"
`,
	}

	for name, contents := range tests {
		t.Run(name, func(t *testing.T) {
			path := writeConfig(t, contents)

			_, err := monitorconfig.Load(path)
			require.Error(t, err)
		})
	}
}
