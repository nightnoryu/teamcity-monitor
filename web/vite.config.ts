import {defineConfig} from "vite";

export default defineConfig({
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
});
