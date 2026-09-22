import type {RegionGroup as RegionGroupData} from "../api/types";
import {ProjectRow} from "./ProjectRow";

export function RegionGroup({group}: {group: RegionGroupData}) {
    return (
        <div className="region-group">
            <div className="region-group-header">
                <span className="region-label">{group.name.toUpperCase()}</span>
                <span className="region-count">
                    {group.builds.length} build{group.builds.length === 1 ? "" : "s"}
                </span>
            </div>
            {group.builds.map((build) => (
                <ProjectRow key={build.buildId} build={build}
                    showBuildName={group.builds.some((other) => other !== build && other.projectId === build.projectId)}
                    showProjectId={group.builds.some((other) => other.projectName === build.projectName && other.projectId !== build.projectId)} />
            ))}
        </div>
    );
}
