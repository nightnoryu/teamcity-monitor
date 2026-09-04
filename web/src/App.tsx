import {EnvironmentCard} from "./components/EnvironmentCard";
import {Footer} from "./components/Footer";
import {Header} from "./components/Header";
import {useStatus} from "./hooks/useStatus";

const POLL_INTERVAL_MS = 20_000;

export function App() {
    const {data, error, loading, lastUpdated, refresh} = useStatus(POLL_INTERVAL_MS);
    const environments = data?.environments ?? [];

    return (
        <div className="container">
            <Header loading={loading} onRefresh={refresh} />

            {error && <div className="error-banner">Failed to load status: {error}</div>}

            {!data?.ready ? (
                <div className="loading-state">Waiting for the first TeamCity poll…</div>
            ) : (
                <>
                    <div className="env-grid">
                        {environments.map((environment) => (
                            <EnvironmentCard key={environment.name} environment={environment} />
                        ))}
                    </div>
                    <Footer environments={environments} lastUpdated={lastUpdated} />
                </>
            )}
        </div>
    );
}
