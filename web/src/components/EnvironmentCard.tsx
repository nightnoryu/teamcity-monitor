import type {EnvironmentStatus} from "../types";
import {RegionGroup} from "./RegionGroup";
import {EnvFractionBadge} from "./StatusBadge";

export function EnvironmentCard({environment}: {environment: EnvironmentStatus}) {
    return (
        <div className="env-card">
            <div className="env-card-header">
                <div className="env-name-group">
                    <span className="env-emoji">{environment.emoji}</span>
                    <span className="env-name">{environment.name}</span>
                </div>
                <EnvFractionBadge successCount={environment.successCount} totalCount={environment.totalCount} />
            </div>
            <div className="region-list">
                {environment.groups.map((group) => (
                    <RegionGroup key={group.name} group={group} />
                ))}
            </div>
        </div>
    );
}
