export type BuildStatus = "success" | "failure" | "error" | "running" | "unknown";

export interface ProjectBuildStatus {
    projectName: string;
    status: BuildStatus;
    branch?: string;
    buildNumber?: string;
    startedAt?: string;
    finishedAt?: string;
    triggeredBy?: string;
    webUrl?: string;
    /** Username of whoever last changed this project's branch parameter for
     * this environment, per TeamCity's audit log. Best-effort: absent if no
     * exact "Value of the parameter X changed" audit event was found. */
    branchChangedBy?: string;
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
    generatedAt?: string;
    environments?: EnvironmentStatus[];
}
