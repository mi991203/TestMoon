import{mkdirSync,copyFileSync}from'node:fs';mkdirSync('dist',{recursive:true});copyFileSync('src/worker.mjs','dist/worker.mjs');
