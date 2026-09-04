package monitor

import (
	"testing"

	"github.com/stretchr/testify/require"

	"teamcity-monitor/internal/monitorconfig"
)

func sampleConfig() *monitorconfig.Config {
	return &monitorconfig.Config{
		TeamCityURL: "https://teamcity.your-org.lan",
		AccessToken: "abc",
		Environments: []monitorconfig.Environment{
			{Name: "dev", Emoji: "🥭"},
			{Name: "orange", Emoji: "🍊"},
		},
		Projects: []monitorconfig.Project{
			{
				Name:                   "Alpha",
				ID:                     "Alpha_Testing",
				EnvironmentBranchParam: "alpha_%s_branch",
				MonitoredBuilds: []monitorconfig.MonitoredBuild{
					{Environment: "dev", Name: "ru", ID: "Alpha_Testing_Dev_Ru"},
					{Environment: "dev", Name: "eu", ID: "Alpha_Testing_Dev_Eu"},
				},
			},
			{
				Name:                   "Beta",
				ID:                     "Beta_Testing",
				EnvironmentBranchParam: "vcs_root_branch_%s",
				MonitoredBuilds: []monitorconfig.MonitoredBuild{
					{Environment: "dev", Name: "build", ID: "Beta_Testing_Dev_Build"},
					{Environment: "dev", Name: "ru", ID: "Beta_Testing_Dev_Ru"},
					{Environment: "dev", Name: "eu", ID: "Beta_Testing_Dev_Eu"},
				},
			},
		},
	}
}

func TestPlanTasks_EnvironmentOrder(t *testing.T) {
	skeleton, _, _ := planTasks(sampleConfig())

	require.Len(t, skeleton, 2)
	require.Equal(t, "dev", skeleton[0].Name)
	require.Equal(t, "orange", skeleton[1].Name)
	require.Empty(t, skeleton[1].Groups, "orange has no monitored builds")
}

// An environment with no monitored builds must serialize Groups as JSON []
// rather than null, or a naive frontend .map() over it panics.
func TestPlanTasks_EmptyEnvironmentGroupsAreNotNil(t *testing.T) {
	skeleton, _, _ := planTasks(sampleConfig())

	require.NotNil(t, skeleton[1].Groups)
}

func TestPlanTasks_GroupOrderFirstSeen(t *testing.T) {
	skeleton, _, _ := planTasks(sampleConfig())

	dev := skeleton[0]
	require.Len(t, dev.Groups, 3)
	require.Equal(t, "ru", dev.Groups[0].Name)
	require.Equal(t, "eu", dev.Groups[1].Name)
	require.Equal(t, "build", dev.Groups[2].Name)

	require.Len(t, dev.Groups[0].Builds, 2, "ru group has Alpha and Beta")
	require.Len(t, dev.Groups[2].Builds, 1, "build group only has Beta")
	require.Equal(t, "Beta", dev.Groups[2].Builds[0].ProjectName)
}

func TestPlanTasks_TaskCountMatchesMonitoredBuilds(t *testing.T) {
	_, tasks, _ := planTasks(sampleConfig())
	require.Len(t, tasks, 5)
}

func TestPlanTasks_AuditTasksGroupedByProjectAndEnvironment(t *testing.T) {
	_, _, auditTasks := planTasks(sampleConfig())

	// Alpha/dev and Beta/dev: one audit task each, regardless of how many
	// monitored builds (groups) that project has under that environment.
	require.Len(t, auditTasks, 2)

	byProject := make(map[string]auditTask, len(auditTasks))
	for _, task := range auditTasks {
		byProject[task.projectID] = task
	}

	alpha := byProject["Alpha_Testing"]
	require.Equal(t, "alpha_dev_branch", alpha.paramName)
	require.Len(t, alpha.targets, 2, "Alpha has 2 monitored builds under dev")

	beta := byProject["Beta_Testing"]
	require.Equal(t, "vcs_root_branch_dev", beta.paramName)
	require.Len(t, beta.targets, 3, "Beta has 3 monitored builds under dev")
}
