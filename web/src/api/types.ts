export type BuildStatus = "success" | "failure" | "error" | "running" | "queued" | "unknown" | "unavailable";

export interface ProjectBuildStatus {
    projectId: string;
    buildId: string;
    buildName: string;
    projectName: string;
    status: BuildStatus;
    branch?: string;
    buildNumber?: string;
    statusText?: string;
    startedAt?: string;
    finishedAt?: string;
    triggeredBy?: string;
    webUrl?: string;
    /** Username of whoever last changed this project's branch parameter for
     * this environment, per TeamCity's audit log. Best-effort: absent if no
     * exact "Value of the parameter X changed" audit event was found. */
    branchChangedBy?: string;
    attributionStatus: "pending" | "found" | "not_found" | "error";
    error?: string;
}

export interface RegionGroup {
    name: string;
    builds: ProjectBuildStatus[];
}

export interface EnvironmentStatus {
    name: string;
    emoji: string;
    successCount: number;
    totalCount: number;
    groups: RegionGroup[];
}

export interface StatusResponse {
    ready: boolean;
    pollIntervalMs?: number;
    generatedAt?: string;
    lastSuccessfulAt?: string;
    collectionHealth?: "healthy" | "partial" | "failed";
    pollDurationMs?: number;
    failedBuilds?: number;
    environments?: EnvironmentStatus[];
}
