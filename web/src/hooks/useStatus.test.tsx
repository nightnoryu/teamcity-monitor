import {act, renderHook, waitFor} from "@testing-library/react";
import {afterEach, expect, it, vi} from "vitest";

import {useStatus} from "./useStatus";

vi.mock("../api/status", () => ({fetchStatus: vi.fn()}));
const {fetchStatus} = await import("../api/status");
const mockedFetchStatus = vi.mocked(fetchStatus);

afterEach(() => { vi.useRealTimers(); vi.clearAllMocks(); });

it("does not start a deferred fetch after immediate unmount", async () => {
    const {unmount} = renderHook(() => useStatus(1000));
    unmount();
    await Promise.resolve();
    expect(mockedFetchStatus).not.toHaveBeenCalled();
});

it("keeps a slow request in flight and retries after it completes", async () => {
    let resolveFirst!: (value: {ready: boolean}) => void;
    mockedFetchStatus.mockReturnValueOnce(new Promise((resolve) => { resolveFirst = resolve; }));
    mockedFetchStatus.mockResolvedValue({ready: true});
    const {result} = renderHook(() => useStatus(1000));
    await waitFor(() => expect(mockedFetchStatus).toHaveBeenCalledTimes(1));
    vi.useFakeTimers();
    await act(async () => { await vi.advanceTimersByTimeAsync(1500); });
    expect(mockedFetchStatus).toHaveBeenCalledTimes(1);
    await act(async () => { resolveFirst({ready: true}); });
    expect(result.current.data?.ready).toBe(true);
    await act(async () => { await vi.advanceTimersByTimeAsync(1000); });
    expect(mockedFetchStatus).toHaveBeenCalledTimes(2);
});

it("ignores an old request that settles after the effect restarts", async () => {
    let rejectFirst!: (reason: Error) => void;
    mockedFetchStatus.mockReturnValueOnce(new Promise((_resolve, reject) => { rejectFirst = reject; }));
    mockedFetchStatus.mockResolvedValue({ready: true});
    const {result, rerender} = renderHook(({interval}) => useStatus(interval), {initialProps: {interval: 1000}});
    await waitFor(() => expect(mockedFetchStatus).toHaveBeenCalledTimes(1));
    rerender({interval: 2000});
    await waitFor(() => expect(mockedFetchStatus).toHaveBeenCalledTimes(2));
    await waitFor(() => expect(result.current.data?.ready).toBe(true));
    await act(async () => { rejectFirst(new Error("old request failed")); });
    expect(result.current.error).toBeNull();
    expect(result.current.data?.ready).toBe(true);
});
