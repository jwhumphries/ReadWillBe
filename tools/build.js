const esbuild = require('esbuild');
const http = require('http');

const watch = process.argv.includes('--watch');

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

async function main() {
    try {
        const ctx = await esbuild.context({
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
        });

        if (watch) {
            await ctx.watch();
            console.log('Watching for changes...');
        } else {
            await ctx.rebuild();
            await ctx.dispose();
        }
    } catch (err) {
        console.error('Build failed:', err);
        process.exit(1);
    }
}

main();
