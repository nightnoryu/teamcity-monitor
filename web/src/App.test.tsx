import {render, screen, waitFor} from "@testing-library/react";
import {afterEach, describe, expect, it, vi} from "vitest";

import {App} from "./App";
import type {StatusResponse} from "./types";

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
            generatedAt: "2026-09-02T14:23:12Z",
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
    });

    it("shows an error banner when the request fails", async () => {
        mockedFetchStatus.mockRejectedValue(new Error("network down"));

        render(<App />);

        await waitFor(() => expect(screen.getByText(/failed to load status/i)).toBeInTheDocument());
        expect(screen.getByText(/network down/)).toBeInTheDocument();
    });
});
