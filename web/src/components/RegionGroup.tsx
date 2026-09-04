import type {RegionGroup as RegionGroupData} from "../api/types";
import {ProjectRow} from "./ProjectRow";

export function RegionGroup({group}: {group: RegionGroupData}) {
    return (
        <div className="region-group">
            <div className="region-group-header">
                <span className="region-label">{group.name.toUpperCase()}</span>
                <span className="region-count">
                    {group.builds.length} project{group.builds.length === 1 ? "" : "s"}
                </span>
            </div>
            {group.builds.map((build) => (
                <ProjectRow key={build.projectName} build={build} />
            ))}
        </div>
    );
}
