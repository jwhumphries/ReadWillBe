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
        port: 7331,
        path: '/_templ/reload',
        method: 'POST',
        timeout: 100
    }, (res) => {
        // console.log(`Reload triggered: ${res.statusCode}`);
    });
    
    req.on('error', (e) => {
        // Ignore errors
    });
    
    req.end();
};

// Watch the output file
let debounceTimer;
fs.watch('static/css/main.css', (eventType, filename) => {
    if (filename && eventType === 'change') {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => {
            console.log('CSS changed, triggering reload...');
            triggerReload();
        }, 100);
    }
});

// Handle exit
process.on('SIGINT', () => {
    tailwind.kill();
    process.exit();
});
