package monitor

import (
	"fmt"

	"teamcity-monitor/internal/monitorconfig"
)

// fetchTask is one monitored build to fetch, with the pre-computed slot in
// the snapshot skeleton its result belongs in.
type fetchTask struct {
	envIndex, groupIndex, rowIndex int

	projectName string
	buildTypeID string
}

// rowRef is a slot in the snapshot skeleton.
type rowRef struct {
	envIndex, groupIndex, rowIndex int
}

// auditTask looks up who last changed a project's branch parameter for one
// environment, and writes the result into every row for that
// (project, environment) pair — the parameter is project-scoped, so it
// applies identically to all of a project's monitored builds (ru/eu/build,
// ...) within that environment.
type auditTask struct {
	projectID string
	paramName string
	targets   []rowRef
}

type groupKey struct {
	environment string
	name        string
}

type auditKey struct {
	projectID   string
	environment string
}

// planTasks builds the ordered snapshot skeleton (environments in
// config.toml order, groups within an environment in first-seen order
// across projects), the flat list of TeamCity build fetches, and the list
// of branch-parameter audit lookups needed to fill it. It performs no I/O
// and is deterministic for a given config.
func planTasks(cfg *monitorconfig.Config) ([]EnvironmentStatus, []fetchTask, []auditTask) {
	envIndex := make(map[string]int, len(cfg.Environments))
	skeleton := make([]EnvironmentStatus, len(cfg.Environments))
	for i, env := range cfg.Environments {
		skeleton[i] = EnvironmentStatus{Name: env.Name, Emoji: env.Emoji, Groups: []RegionGroup{}}
		envIndex[env.Name] = i
	}

	groupIndex := make(map[groupKey]int)
	auditIndex := make(map[auditKey]int)
	var tasks []fetchTask
	var auditTasks []auditTask

	for _, project := range cfg.Projects {
		for _, build := range project.MonitoredBuilds {
			ei, ok := envIndex[build.Environment]
			if !ok {
				// Validate() rejects this at config-load time; defensive only.
				continue
			}

			gi := groupIndexFor(&skeleton[ei], groupIndex, build.Environment, build.Name)

			skeleton[ei].Groups[gi].Builds = append(
				skeleton[ei].Groups[gi].Builds,
				ProjectBuildStatus{ProjectName: project.Name},
			)
			ri := len(skeleton[ei].Groups[gi].Builds) - 1

			tasks = append(tasks, fetchTask{
				envIndex:    ei,
				groupIndex:  gi,
				rowIndex:    ri,
				projectName: project.Name,
				buildTypeID: build.ID,
			})

			addAuditTarget(&auditTasks, auditIndex, project, build.Environment, rowRef{ei, gi, ri})
		}
	}

	return skeleton, tasks, auditTasks
}

func addAuditTarget(
	auditTasks *[]auditTask, auditIndex map[auditKey]int, project monitorconfig.Project, environment string, ref rowRef,
) {
	key := auditKey{projectID: project.ID, environment: environment}
	if i, ok := auditIndex[key]; ok {
		(*auditTasks)[i].targets = append((*auditTasks)[i].targets, ref)
		return
	}

	auditIndex[key] = len(*auditTasks)
	*auditTasks = append(*auditTasks, auditTask{
		projectID: project.ID,
		paramName: fmt.Sprintf(project.EnvironmentBranchParam, environment),
		targets:   []rowRef{ref},
	})
}

func groupIndexFor(env *EnvironmentStatus, groupIndex map[groupKey]int, environment, name string) int {
	key := groupKey{environment: environment, name: name}
	if gi, ok := groupIndex[key]; ok {
		return gi
	}

	env.Groups = append(env.Groups, RegionGroup{Name: name})
	gi := len(env.Groups) - 1
	groupIndex[key] = gi

	return gi
}
