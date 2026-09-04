import type {BuildStatus} from "../api/types";

const STATUS_LABEL: Record<BuildStatus, string> = {
    success: "success",
    failure: "failed",
    error: "error",
    running: "running",
    unknown: "unknown",
};

const STATUS_GLYPH: Record<BuildStatus, string> = {
    success: "✅",
    failure: "❌",
    error: "⚠",
    running: "⏳",
    unknown: "❔",
};

export function BuildStatusPill({status}: {status: BuildStatus}) {
    return (
        <span className={`build-status ${status}`}>
            <span className="status-icon">{STATUS_GLYPH[status]}</span> {STATUS_LABEL[status]}
        </span>
    );
}

function fractionColor(successCount: number, totalCount: number): "green" | "yellow" | "red" | "gray" {
    if (totalCount === 0) return "gray";
    if (successCount === 0) return "red";
    if (successCount === totalCount) return "green";
    return "yellow";
}

export function EnvFractionBadge({successCount, totalCount}: {successCount: number; totalCount: number}) {
    const color = fractionColor(successCount, totalCount);

    return (
        <span className="env-status-badge">
            <span className={`dot ${color}`} />
            {successCount} / {totalCount}
        </span>
    );
}
