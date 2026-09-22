import type {EnvironmentStatus} from "../api/types";

export function Footer({environments, generatedAt, lastSuccessfulAt, now}: {environments: EnvironmentStatus[]; generatedAt?: string; lastSuccessfulAt?: string; now: number}) {
    const projectNames = new Set(
        environments.flatMap((env) => env.groups.flatMap((group) => group.builds.map((build) => build.projectName))),
    );
    const groupCount = environments.reduce((sum, env) => sum + env.groups.length, 0);

    return (
        <div className="dashboard-footer">
            <span>
                {environments.length} environments · {projectNames.size} projects · {groupCount} groups
            </span>
            <span>Snapshot checked: {generatedAt ? `${new Date(generatedAt).toLocaleString()} (${Math.max(0, Math.floor((now - Date.parse(generatedAt)) / 1000))}s ago)` : "—"} · Last successful collection: {lastSuccessfulAt ? new Date(lastSuccessfulAt).toLocaleString() : "never"}</span>
        </div>
    );
}
