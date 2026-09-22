import {render, screen} from "@testing-library/react";
import {expect, it} from "vitest";

import {RegionGroup} from "./RegionGroup";

it("shows build labels when one project has multiple builds in a group", () => {
    render(<RegionGroup group={{name: "ru", builds: [
        {projectId: "A", buildId: "A_One", buildName: "API", projectName: "Alpha", status: "queued", attributionStatus: "not_found"},
        {projectId: "A", buildId: "A_Two", buildName: "Worker", projectName: "Alpha", status: "running", attributionStatus: "not_found"},
        {projectId: "B", buildId: "B_One", buildName: "API", projectName: "Alpha", status: "success", attributionStatus: "not_found"},
    ]}} />);
    expect(screen.getByText(/· API/)).toBeInTheDocument();
    expect(screen.getByText(/· Worker/)).toBeInTheDocument();
    expect(screen.getByText(/\(B\)/)).toBeInTheDocument();
    expect(screen.getByText("3 builds")).toBeInTheDocument();
});
