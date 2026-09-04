import type {EnvironmentStatus} from "../api/types";

export function Footer({environments, lastUpdated}: {environments: EnvironmentStatus[]; lastUpdated: Date | null}) {
    const projectNames = new Set(
        environments.flatMap((env) => env.groups.flatMap((group) => group.builds.map((build) => build.projectName))),
    );
    const groupCount = environments.reduce((sum, env) => sum + env.groups.length, 0);

    return (
        <div className="dashboard-footer">
            <span>
                {environments.length} environments · {projectNames.size} projects · {groupCount} groups
            </span>
            <span>Last updated: {lastUpdated ? lastUpdated.toLocaleString() : "—"}</span>
        </div>
    );
}
