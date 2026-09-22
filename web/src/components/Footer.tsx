import type {EnvironmentStatus} from "../api/types";

export function Footer({environments, generatedAt, lastSuccessfulAt, pollDurationMs, failedBuilds, now}: {environments: EnvironmentStatus[]; generatedAt?: string; lastSuccessfulAt?: string; pollDurationMs?: number; failedBuilds?: number; now: number}) {
    const projectIds = new Set(
        environments.flatMap((env) => env.groups.flatMap((group) => group.builds.map((build) => build.projectId))),
    );
    const groupCount = environments.reduce((sum, env) => sum + env.groups.length, 0);

    return (
        <div className="dashboard-footer">
            <span>
                {environments.length} environments · {projectIds.size} projects · {groupCount} groups
            </span>
            <span>Snapshot checked: {generatedAt ? `${new Date(generatedAt).toLocaleString()} (${Math.max(0, Math.floor((now - Date.parse(generatedAt)) / 1000))}s ago)` : "—"} · Last successful collection: {lastSuccessfulAt ? new Date(lastSuccessfulAt).toLocaleString() : "never"}</span>
            <span>Poll: {pollDurationMs === undefined ? "—" : `${pollDurationMs} ms`} · Failed build fetches: {failedBuilds ?? 0}</span>
        </div>
    );
}
