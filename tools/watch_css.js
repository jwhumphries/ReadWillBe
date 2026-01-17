const { spawn } = require('child_process');
const fs = require('fs');
const http = require('http');

// Run Tailwind CLI
const tailwind = spawn('bun', ['run', 'dev'], {
    stdio: 'inherit',
    shell: true
});

const triggerReload = () => {
    const req = http.request({
        hostname: 'localhost',
        port: 7332,
        path: '/_templ/reload',
        method: 'POST',
        timeout: 100
    }, () => {});

    req.on('error', () => {
        // Ignore errors (proxy might not be up yet)
    });

    req.end();
};

// Wait for file to exist before watching
const cssPath = 'static/css/main.css';
const startWatching = () => {
    let debounceTimer;
    fs.watch(cssPath, (eventType, filename) => {
        if (filename && eventType === 'change') {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => {
                console.log('CSS changed, triggering reload...');
                triggerReload();
            }, 100);
        }
    });
    console.log('Watching CSS for changes...');
};

// Poll for file existence
const waitForFile = () => {
    if (fs.existsSync(cssPath)) {
        startWatching();
    } else {
        setTimeout(waitForFile, 500);
    }
};
waitForFile();

// Handle exit signals
const cleanup = () => {
    tailwind.kill('SIGTERM');
    process.exit(0);
};
process.on('SIGINT', cleanup);
process.on('SIGTERM', cleanup);
