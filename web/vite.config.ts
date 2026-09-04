import react from "@vitejs/plugin-react";
import {defineConfig} from "vitest/config";

export default defineConfig({
    plugins: [react()],
    server: {
        host: true,
        port: 3000,
        strictPort: true,
        allowedHosts: ["teamcity-monitor.lan"],
        watch: {
            usePolling: true,
            interval: 100,
        },
    },
    test: {
        environment: "jsdom",
        setupFiles: ["./src/test-setup.ts"],
    },
});
