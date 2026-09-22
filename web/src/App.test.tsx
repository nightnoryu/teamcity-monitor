import {fireEvent, render, screen, waitFor} from "@testing-library/react";
import {afterEach, describe, expect, it, vi} from "vitest";

import {App} from "./App";
import type {StatusResponse} from "./api/types";

vi.mock("./api/status", () => ({
    fetchStatus: vi.fn(),
}));

const {fetchStatus} = await import("./api/status");
const mockedFetchStatus = vi.mocked(fetchStatus);

afterEach(() => {
    vi.clearAllMocks();
});

describe("App", () => {
    it("shows a waiting state before the first poll completes", () => {
        mockedFetchStatus.mockReturnValue(new Promise(() => {}));

        render(<App />);

        expect(screen.getByText(/waiting for the first teamcity poll/i)).toBeInTheDocument();
    });

    it("renders environment cards once data is ready", async () => {
        const response: StatusResponse = {
            ready: true,
            generatedAt: new Date().toISOString(),
            lastSuccessfulAt: new Date().toISOString(),
            collectionHealth: "healthy",
            environments: [
                {
                    name: "dev",
                    emoji: "🥭",
                    successCount: 2,
                    totalCount: 2,
                    groups: [
                        {
                            name: "ru",
                            builds: [
                                {projectName: "Alpha", status: "success"},
                                {projectName: "Beta", status: "success"},
                            ],
                        },
                    ],
                },
            ],
        };
        mockedFetchStatus.mockResolvedValue(response);

        render(<App />);

        await waitFor(() => expect(screen.getByText("dev")).toBeInTheDocument());
        expect(screen.getByText("Alpha")).toBeInTheDocument();
        expect(screen.getByText("2 / 2")).toBeInTheDocument();
        expect(screen.getByText("Live")).toBeInTheDocument();
        const checked = screen.getByText(/Snapshot checked:/).textContent;
        fireEvent.click(screen.getByRole("button", {name: "Refresh data"}));
        await waitFor(() => expect(mockedFetchStatus).toHaveBeenCalledTimes(2));
        expect(screen.getByText(/Snapshot checked:/).textContent).toBe(checked);
    });

    it("shows upstream failure and row errors separately from browser connectivity", async () => {
        mockedFetchStatus.mockResolvedValue({
            ready: true, generatedAt: new Date().toISOString(), collectionHealth: "partial",
            environments: [{name: "dev", emoji: "", successCount: 0, totalCount: 2, groups: [{name: "ru", builds: [
                {projectName: "Alpha", status: "unavailable", error: "teamcity: unauthorized"},
                {projectName: "Beta", status: "unknown"},
            ]}]}],
        });
        render(<App />);
        expect(await screen.findByText("Partial TeamCity failure")).toBeInTheDocument();
        expect(screen.getByText("teamcity: unauthorized")).toBeInTheDocument();
        expect(screen.getByText("never run")).toBeInTheDocument();
        expect(screen.getByText(/Last successful collection: never/)).toBeInTheDocument();
    });

    it("shows total upstream failure", async () => {
        mockedFetchStatus.mockResolvedValue({ready: true, generatedAt: new Date().toISOString(), collectionHealth: "failed", environments: []});
        render(<App />);
        expect(await screen.findByText("TeamCity unavailable")).toBeInTheDocument();
    });

    it("marks a retained snapshot as cached after browser failure", async () => {
        mockedFetchStatus.mockResolvedValueOnce({ready: true, generatedAt: new Date().toISOString(), collectionHealth: "healthy", environments: []});
        mockedFetchStatus.mockRejectedValueOnce(new Error("network down"));
        render(<App />);
        expect(await screen.findByText("Live")).toBeInTheDocument();
        fireEvent.click(screen.getByRole("button", {name: "Refresh data"}));
        expect(await screen.findByText("Browser disconnected")).toBeInTheDocument();
        expect(screen.getByText(/Showing a cached snapshot/)).toBeInTheDocument();
    });

    it("marks an old but reachable snapshot as stale", async () => {
        mockedFetchStatus.mockResolvedValue({ready: true, generatedAt: new Date(Date.now() - 120_000).toISOString(), collectionHealth: "healthy", environments: []});
        render(<App />);
        expect(await screen.findByText("Snapshot stale")).toBeInTheDocument();
    });

    it("uses the backend poll interval when judging freshness", async () => {
        mockedFetchStatus.mockResolvedValue({ready: true, generatedAt: new Date(Date.now() - 50_000).toISOString(), pollIntervalMs: 60_000, collectionHealth: "healthy", environments: []});
        render(<App />);
        expect(await screen.findByText("Live")).toBeInTheDocument();
        expect(screen.queryByText(/Snapshot is stale/)).toBeNull();
    });

    it("shows an error banner when the request fails", async () => {
        mockedFetchStatus.mockRejectedValue(new Error("network down"));

        render(<App />);

        await waitFor(() => expect(screen.getByText(/failed to load status/i)).toBeInTheDocument());
        expect(screen.getByText(/network down/)).toBeInTheDocument();
    });
});
