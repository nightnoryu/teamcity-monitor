import {useEffect, useState} from "react";

export function BranchName({branch}: {branch: string}) {
    const [status, setStatus] = useState<"idle" | "copying" | "copied" | "error">("idle");

    useEffect(() => {
        if (status !== "copied") return;
        const timeout = window.setTimeout(() => setStatus("idle"), 2000);
        return () => window.clearTimeout(timeout);
    }, [status]);

    async function copyBranch() {
        setStatus("copying");
        try {
            await navigator.clipboard.writeText(branch);
            setStatus("copied");
        } catch {
            setStatus("error");
        }
    }

    return (
        <span className="branch-name">
            <span className="branch-icon" aria-hidden="true">⎇</span>
            <span className="branch-text">{branch}</span>
            <button
                type="button"
                className="branch-copy"
                onClick={copyBranch}
                disabled={status === "copying"}
                title="Copy branch name"
                aria-label={`Copy branch name ${branch}`}
            >
                {status === "copied" ? <span aria-hidden="true">✓</span> : (
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                        <rect x="8" y="8" width="12" height="12" rx="2" />
                        <path d="M16 8V4H4v12h4" />
                    </svg>
                )}
            </button>
            <span className={`branch-copy-feedback${status === "error" ? " error" : ""}`} role="status">
                {status === "copied" && "Copied"}
                {status === "error" && "Couldn’t copy. Select the branch name to copy manually."}
            </span>
        </span>
    );
}
