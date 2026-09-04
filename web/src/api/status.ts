import type {StatusResponse} from "../types";

export async function fetchStatus(signal?: AbortSignal): Promise<StatusResponse> {
    const response = await fetch("/api/status", {signal});
    if (!response.ok) {
        throw new Error(`status request failed: ${response.status}`);
    }
    return (await response.json()) as StatusResponse;
}
