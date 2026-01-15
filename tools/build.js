const esbuild = require('esbuild');

const watch = process.argv.includes('--watch');

const ctx = esbuild.context({
    entryPoints: ['assets/js/index.tsx'],
    bundle: true,
    minify: !watch,
    sourcemap: true,
    outfile: 'static/js/bundle.js',
    logLevel: 'info',
    loader: { '.tsx': 'tsx', '.ts': 'ts' },
}).then(ctx => {
    if (watch) {
        ctx.watch();
        console.log('Watching for changes...');
    } else {
        ctx.rebuild().then(() => ctx.dispose());
    }
});
