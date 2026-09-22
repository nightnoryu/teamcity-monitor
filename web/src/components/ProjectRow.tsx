import type {ProjectBuildStatus} from "../api/types";
import {formatDate} from "../utils/format";
import {BuildStatusPill} from "./StatusBadge";

function safeBuildUrl(value?: string): string | undefined {
    if (!value) return undefined;
    try {
        const url = new URL(value);
        return url.protocol === "http:" || url.protocol === "https:" ? url.href : undefined;
    } catch {
        return undefined;
    }
}

export function ProjectRow({build, showBuildName = false, showProjectId = false}: {build: ProjectBuildStatus; showBuildName?: boolean; showProjectId?: boolean}) {
    const buildUrl = safeBuildUrl(build.webUrl);
    return (
        <div className="project-row">
            <span className="project-name">{build.projectName}{showProjectId && <span className="build-label"> ({build.projectId})</span>}{showBuildName && <span className="build-label"> · {build.buildName || build.buildId}</span>}</span>
            {build.branchChangedBy && (
                <span className="branch-editor" title="Last changed the branch parameter for this environment">
                    👤 {build.branchChangedBy}
                </span>
            )}
            {build.attributionStatus === "not_found" && <span className="branch-editor" title="No exact branch parameter edit found in the 100 most recent project edit events">Editor not found</span>}
            {build.attributionStatus === "pending" && <span className="branch-editor" title="Branch parameter attribution has not been collected yet">Checking editor…</span>}
            {build.attributionStatus === "error" && <span className="branch-editor" title="Branch parameter audit lookup failed">Editor unavailable</span>}
            <div className="project-details">
                {build.branch && (
                    <span className="branch-name">
                        <span className="branch-icon">⎇</span>
                        {build.branch}
                    </span>
                )}
                {build.finishedAt && <span className="deploy-date">🕐 {formatDate(build.finishedAt)}</span>}
                {build.startedAt && <span title="Build started">Started {formatDate(build.startedAt)}</span>}
                {build.triggeredBy && <span title="Build trigger">Triggered by {build.triggeredBy}</span>}
                {build.buildNumber && (buildUrl ? <a href={buildUrl} target="_blank" rel="noopener noreferrer" aria-label={`Open TeamCity build ${build.buildNumber}`}>Build #{build.buildNumber}</a> : <span>Build #{build.buildNumber}</span>)}
                {!build.buildNumber && buildUrl && <a href={buildUrl} target="_blank" rel="noopener noreferrer">Open TeamCity build</a>}
                <BuildStatusPill status={build.status} />
                {build.statusText && <span className="build-status-text">{build.statusText}</span>}
                {build.error && <span className="build-error" role="alert">{build.error}</span>}
            </div>
        </div>
    );
}
