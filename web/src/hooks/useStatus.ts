import {useCallback, useEffect, useRef, useState} from "react";

import {fetchStatus} from "../api/status";
import type {StatusResponse} from "../api/types";

interface UseStatusResult {
    data: StatusResponse | null;
    error: string | null;
    loading: boolean;
    lastUpdated: Date | null;
    refresh: () => void;
}

export function useStatus(intervalMs: number): UseStatusResult {
    const [data, setData] = useState<StatusResponse | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(true);
    const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

    const abortRef = useRef<AbortController | null>(null);
    const timerRef = useRef<ReturnType<typeof setInterval> | undefined>(undefined);

    const load = useCallback(async () => {
        abortRef.current?.abort();
        const controller = new AbortController();
        abortRef.current = controller;

        setLoading(true);
        try {
            const response = await fetchStatus(controller.signal);
            setData(response);
            setError(null);
            setLastUpdated(new Date());
        } catch (err) {
            if (controller.signal.aborted) return;
            setError(err instanceof Error ? err.message : "failed to load status");
        } finally {
            if (!controller.signal.aborted) setLoading(false);
        }
    }, []);

    const scheduleInterval = useCallback(() => {
        if (timerRef.current) clearInterval(timerRef.current);
        timerRef.current = setInterval(() => void load(), intervalMs);
    }, [load, intervalMs]);

    useEffect(() => {
        // load() sets loading state synchronously before its first await;
        // deferring the call to a microtask keeps that setState out of the
        // effect body itself, as react-hooks/set-state-in-effect requires.
        void Promise.resolve().then(load);
        scheduleInterval();
        return () => {
            if (timerRef.current) clearInterval(timerRef.current);
            abortRef.current?.abort();
        };
    }, [load, scheduleInterval]);

    const refresh = useCallback(() => {
        void load();
        scheduleInterval();
    }, [load, scheduleInterval]);

    return {data, error, loading, lastUpdated, refresh};
}
