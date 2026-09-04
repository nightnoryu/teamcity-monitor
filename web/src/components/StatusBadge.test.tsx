import {render, screen} from "@testing-library/react";
import {describe, expect, it} from "vitest";

import {BuildStatusPill, EnvFractionBadge} from "./StatusBadge";

describe("BuildStatusPill", () => {
    it("renders the success label", () => {
        render(<BuildStatusPill status="success" />);
        expect(screen.getByText("success")).toBeInTheDocument();
    });

    it("renders the failed label for a failure status", () => {
        render(<BuildStatusPill status="failure" />);
        expect(screen.getByText("failed")).toBeInTheDocument();
    });

    it("renders the unknown label", () => {
        render(<BuildStatusPill status="unknown" />);
        expect(screen.getByText("unknown")).toBeInTheDocument();
    });

    it("renders the running label", () => {
        render(<BuildStatusPill status="running" />);
        expect(screen.getByText("running")).toBeInTheDocument();
    });
});

describe("EnvFractionBadge", () => {
    it("shows a green dot when all builds succeeded", () => {
        const {container} = render(<EnvFractionBadge successCount={3} totalCount={3} />);
        expect(container.querySelector(".dot.green")).not.toBeNull();
        expect(screen.getByText("3 / 3")).toBeInTheDocument();
    });

    it("shows a yellow dot for a partial success", () => {
        const {container} = render(<EnvFractionBadge successCount={2} totalCount={3} />);
        expect(container.querySelector(".dot.yellow")).not.toBeNull();
    });

    it("shows a red dot when nothing succeeded", () => {
        const {container} = render(<EnvFractionBadge successCount={0} totalCount={3} />);
        expect(container.querySelector(".dot.red")).not.toBeNull();
    });

    it("shows a gray dot when no builds are configured", () => {
        const {container} = render(<EnvFractionBadge successCount={0} totalCount={0} />);
        expect(container.querySelector(".dot.gray")).not.toBeNull();
    });
});
