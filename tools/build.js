const esbuild = require('esbuild');
const http = require('http');

const watch = process.argv.includes('--watch');

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
        // Ignore errors (proxy might not be up yet)
    });
    
    req.end();
};

const ctx = esbuild.context({
    entryPoints: ['assets/js/index.tsx'],
    bundle: true,
    minify: !watch,
    sourcemap: true,
    outfile: 'static/js/bundle.js',
    logLevel: 'info',
    loader: { '.tsx': 'tsx', '.ts': 'ts' },
    plugins: [{
        name: 'reload-plugin',
        setup(build) {
            build.onEnd(result => {
                if (result.errors.length === 0 && watch) {
                    triggerReload();
                }
            });
        },
    }],
}).then(ctx => {
    if (watch) {
        ctx.watch();
        console.log('Watching for changes...');
    } else {
        ctx.rebuild().then(() => ctx.dispose());
    }
});