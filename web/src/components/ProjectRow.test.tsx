import {act, fireEvent, render, screen} from "@testing-library/react";
import {afterEach, describe, expect, it, vi} from "vitest";

import type {ProjectBuildStatus} from "../api/types";
import {ProjectRow} from "./ProjectRow";

const baseBuild: ProjectBuildStatus = {
    projectId: "A", buildId: "A_Ru", buildName: "ru", attributionStatus: "found",
    projectName: "Alpha",
    status: "success",
    branch: "feature/order-export",
    finishedAt: "2026-09-02T11:20:00+03:00",
};

describe("ProjectRow", () => {
    afterEach(() => {
        vi.unstubAllGlobals();
        vi.useRealTimers();
    });

    it("copies the exact branch name and briefly confirms success", async () => {
        vi.useFakeTimers();
        const writeText = vi.fn().mockResolvedValue(undefined);
        vi.stubGlobal("navigator", {clipboard: {writeText}});
        render(<ProjectRow build={baseBuild} />);

        await act(async () => {
            fireEvent.click(screen.getByRole("button", {name: "Copy branch name feature/order-export"}));
        });

        expect(writeText).toHaveBeenCalledWith("feature/order-export");
        expect(screen.getByRole("status")).toHaveTextContent("Copied");
        act(() => vi.advanceTimersByTime(2000));
        expect(screen.getByRole("status")).toBeEmptyDOMElement();
    });

    it("shows a copy failure and allows retrying", async () => {
        const writeText = vi.fn().mockRejectedValueOnce(new Error("Permission denied")).mockResolvedValueOnce(undefined);
        vi.stubGlobal("navigator", {clipboard: {writeText}});
        render(<ProjectRow build={baseBuild} />);
        const button = screen.getByRole("button", {name: "Copy branch name feature/order-export"});

        fireEvent.click(button);
        expect(await screen.findByText(/Couldn’t copy/)).toBeInTheDocument();
        expect(screen.getByText("feature/order-export")).toBeInTheDocument();
        fireEvent.click(button);
        expect(await screen.findByText("Copied")).toBeInTheDocument();
    });

    it("explains manual copying when clipboard access is unavailable", async () => {
        vi.stubGlobal("navigator", {});
        render(<ProjectRow build={baseBuild} />);

        fireEvent.click(screen.getByRole("button", {name: "Copy branch name feature/order-export"}));

        expect(await screen.findByText(/Select the branch name to copy manually/)).toBeInTheDocument();
    });

    it("resets feedback when the branch changes, including a pending copy", async () => {
        let finishCopy!: () => void;
        const writeText = vi.fn().mockReturnValue(new Promise<void>((resolve) => { finishCopy = resolve; }));
        vi.stubGlobal("navigator", {clipboard: {writeText}});
        const {rerender} = render(<ProjectRow build={baseBuild} />);
        fireEvent.click(screen.getByRole("button", {name: "Copy branch name feature/order-export"}));

        rerender(<ProjectRow build={{...baseBuild, branch: "main"}} />);
        await act(async () => finishCopy());

        expect(screen.getByRole("button", {name: "Copy branch name main"})).toBeEnabled();
        expect(screen.getByRole("status")).toBeEmptyDOMElement();
    });

    it("renders project name, branch, and status", () => {
        render(<ProjectRow build={baseBuild} />);

        expect(screen.getByText("Alpha")).toBeInTheDocument();
        expect(screen.getByText("feature/order-export")).toBeInTheDocument();
        expect(screen.getByText("success")).toBeInTheDocument();
    });

    it("omits optional fields that are absent", () => {
        render(<ProjectRow build={{projectId: "B", buildId: "B_Ru", buildName: "ru", attributionStatus: "not_found", projectName: "Beta", status: "unknown"}} />);

        expect(screen.getByText("Beta")).toBeInTheDocument();
        expect(screen.queryByText(/feature\//)).toBeNull();
        expect(screen.queryByText(/a\.kovalev/)).toBeNull();
        expect(screen.queryByRole("button", {name: /Copy branch name/})).toBeNull();
    });

    it("shows who last changed the branch parameter, when known", () => {
        render(<ProjectRow build={{...baseBuild, branchChangedBy: "a.kovalev"}} />);

        expect(screen.getByText(/a\.kovalev/)).toBeInTheDocument();
    });

    it("links to an HTTP TeamCity build and shows investigation details", () => {
        render(<ProjectRow build={{...baseBuild, buildNumber: "42", webUrl: "https://teamcity.example/build/42", statusText: "Tests failed", triggeredBy: "alice", startedAt: "2026-09-02T11:00:00+03:00"}} />);
        expect(screen.getByRole("link", {name: "Open TeamCity build 42"})).toHaveAttribute("href", "https://teamcity.example/build/42");
        expect(screen.getByText("Tests failed")).toBeInTheDocument();
        expect(screen.getByText("Triggered by alice")).toBeInTheDocument();
    });

    it("does not link unsafe build URLs", () => {
        render(<ProjectRow build={{...baseBuild, buildNumber: "42", webUrl: "javascript:alert(1)"}} />);
        expect(screen.queryByRole("link")).toBeNull();
        expect(screen.getByText("Build #42")).toBeInTheDocument();
    });
});
