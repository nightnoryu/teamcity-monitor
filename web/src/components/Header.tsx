export function Header({loading, state, onRefresh}: {loading: boolean; state: "connecting" | "live" | "partial" | "failed" | "disconnected" | "stale"; onRefresh: () => void}) {
    const labels = {connecting: "Connecting", live: "Live", partial: "Partial TeamCity failure", failed: "TeamCity unavailable", disconnected: "Browser disconnected", stale: "Snapshot stale"};
    return (
        <header className="dashboard-header">
            <div className="header-left">
                <div className="logo-icon">🦝</div>
                <h1>
                    TeamCity <span>Deployments</span>
                </h1>
            </div>
            <div className="header-right">
                <span className={`status-dot ${state}`} />
                <span className="live-label">{labels[state]}</span>
                <button
                    className={`refresh-btn${loading ? " spinning" : ""}`}
                    onClick={onRefresh}
                    title="Refresh data"
                    aria-label="Refresh data"
                >
                    ⟳
                </button>
            </div>
        </header>
    );
}
