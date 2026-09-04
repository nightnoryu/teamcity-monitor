import type {ProjectBuildStatus} from "../api/types";
import {formatDate} from "../utils/format";
import {BuildStatusPill} from "./StatusBadge";

export function ProjectRow({build}: {build: ProjectBuildStatus}) {
    return (
        <div className="project-row">
            <span className="project-name">{build.projectName}</span>
            {build.branchChangedBy && (
                <span className="branch-editor" title="Last changed the branch parameter for this environment">
                    👤 {build.branchChangedBy}
                </span>
            )}
            <div className="project-details">
                {build.branch && (
                    <span className="branch-name">
                        <span className="branch-icon">⎇</span>
                        {build.branch}
                    </span>
                )}
                {build.finishedAt && <span className="deploy-date">🕐 {formatDate(build.finishedAt)}</span>}
                <BuildStatusPill status={build.status} />
            </div>
        </div>
    );
}
