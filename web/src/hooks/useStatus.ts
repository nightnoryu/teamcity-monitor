import {useCallback, useEffect, useRef, useState} from "react";

import {fetchStatus} from "../api/status";
import type {StatusResponse} from "../api/types";

interface UseStatusResult {
    data: StatusResponse | null;
    error: string | null;
    loading: boolean;
    refresh: () => void;
}

export function useStatus(intervalMs: number): UseStatusResult {
    const [data, setData] = useState<StatusResponse | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(true);

    const abortRef = useRef<AbortController | null>(null);
    const timerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const mountedRef = useRef(false);
    const loadRef = useRef<() => Promise<void>>(async () => {});

    const load = useCallback(async () => {
        if (!mountedRef.current || abortRef.current) return;
        const controller = new AbortController();
        abortRef.current = controller;
        const timeout = setTimeout(() => controller.abort("Status request timed out"), Math.max(intervalMs, 30000));

        setLoading(true);
        try {
            const response = await fetchStatus(controller.signal);
            if (mountedRef.current && abortRef.current === controller && !controller.signal.aborted) {
                setData(response);
                setError(null);
            }
        } catch (err) {
            if (mountedRef.current && abortRef.current === controller) {
                setError(controller.signal.aborted ? "Status request timed out" : err instanceof Error ? err.message : "failed to load status");
            }
        } finally {
            clearTimeout(timeout);
            const current = mountedRef.current && abortRef.current === controller;
            if (abortRef.current === controller) abortRef.current = null;
            if (current) {
                setLoading(false);
                timerRef.current = setTimeout(() => void loadRef.current(), intervalMs);
            }
        }
    }, [intervalMs]);

    useEffect(() => { loadRef.current = load; }, [load]);

    useEffect(() => {
        // load() sets loading state synchronously before its first await;
        // deferring the call to a microtask keeps that setState out of the
        // effect body itself, as react-hooks/set-state-in-effect requires.
        mountedRef.current = true;
        void Promise.resolve().then(() => { if (mountedRef.current) void load(); });
        return () => {
            mountedRef.current = false;
            if (timerRef.current) clearTimeout(timerRef.current);
            abortRef.current?.abort();
            abortRef.current = null;
        };
    }, [load]);

    const refresh = useCallback(() => {
        if (timerRef.current) clearTimeout(timerRef.current);
        if (abortRef.current) return;
        void load();
    }, [load]);

    return {data, error, loading, refresh};
}
