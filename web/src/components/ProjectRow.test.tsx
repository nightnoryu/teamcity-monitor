import {render, screen} from "@testing-library/react";
import {describe, expect, it} from "vitest";

import type {ProjectBuildStatus} from "../types";
import {ProjectRow} from "./ProjectRow";

const baseBuild: ProjectBuildStatus = {
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
        render(<ProjectRow build={{projectName: "Beta", status: "unknown"}} />);

        expect(screen.getByText("Beta")).toBeInTheDocument();
        expect(screen.queryByText(/feature\//)).toBeNull();
        expect(screen.queryByText(/a\.kovalev/)).toBeNull();
    });

    it("shows who last changed the branch parameter, when known", () => {
        render(<ProjectRow build={{...baseBuild, branchChangedBy: "a.kovalev"}} />);

        expect(screen.getByText(/a\.kovalev/)).toBeInTheDocument();
    });
});
