import {render, screen} from "@testing-library/react";
import {expect, it} from "vitest";

import {Footer} from "./Footer";

it("counts projects by ID when display names collide", () => {
    render(<Footer now={Date.now()} environments={[{name: "dev", emoji: "", successCount: 0, totalCount: 2, groups: [{name: "ru", builds: [
        {projectId: "A", buildId: "A_Build", buildName: "ru", projectName: "Alpha", status: "success", attributionStatus: "not_found"},
        {projectId: "B", buildId: "B_Build", buildName: "ru", projectName: "Alpha", status: "success", attributionStatus: "not_found"},
    ]}]}]} />);
    expect(screen.getByText(/2 projects/)).toBeInTheDocument();
});
