import {useEffect, useState} from "react";

import {EnvironmentCard} from "./components/EnvironmentCard";
import {Footer} from "./components/Footer";
import {Header} from "./components/Header";
import {useStatus} from "./hooks/useStatus";

const POLL_INTERVAL_MS = 20_000;

export function App() {
    const {data, error, loading, refresh} = useStatus(POLL_INTERVAL_MS);
    const [now, setNow] = useState(() => Date.now());
    useEffect(() => {
        const timer = setInterval(() => setNow(Date.now()), 1_000);
        return () => clearInterval(timer);
    }, []);
    const environments = data?.environments ?? [];
    const age = data?.generatedAt ? now - Date.parse(data.generatedAt) : Infinity;
    const staleAfterMs = Math.max(POLL_INTERVAL_MS, data?.pollIntervalMs ?? POLL_INTERVAL_MS) * 2;
    const stale = Boolean(data?.ready) && age > staleAfterMs;
    const state = error ? "disconnected" : !data?.ready ? "connecting" :
        data.collectionHealth === "failed" ? "failed" :
        data.collectionHealth === "partial" ? "partial" : stale ? "stale" : "live";

    return (
        <div className="container">
            <Header loading={loading} state={state} onRefresh={refresh} />

            {error && <div className="error-banner">Failed to load status: {error}. {data?.ready && "Showing a cached snapshot."}</div>}
            {stale && !error && <div className="error-banner">Snapshot is stale; the latest collection finished over {Math.round(staleAfterMs / 1_000)} seconds ago.</div>}

            {!data?.ready ? (
                <div className="loading-state">Waiting for the first TeamCity poll…</div>
            ) : (
                <>
                    <div className="env-grid">
                        {environments.map((environment) => (
                            <EnvironmentCard key={environment.name} environment={environment} />
                        ))}
                    </div>
                    <Footer environments={environments} generatedAt={data.generatedAt} lastSuccessfulAt={data.lastSuccessfulAt} pollDurationMs={data.pollDurationMs} failedBuilds={data.failedBuilds} now={now} />
                </>
            )}
        </div>
    );
}
