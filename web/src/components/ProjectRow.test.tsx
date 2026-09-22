import {render, screen} from "@testing-library/react";
import {describe, expect, it} from "vitest";

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
