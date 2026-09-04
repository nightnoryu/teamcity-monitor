export function Header({loading, onRefresh}: {loading: boolean; onRefresh: () => void}) {
    return (
        <header className="dashboard-header">
            <div className="header-left">
                <div className="logo-icon">🦝</div>
                <h1>
                    TeamCity <span>Deployments</span>
                </h1>
            </div>
            <div className="header-right">
                <span className="status-dot" />
                <span className="live-label">Live</span>
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
