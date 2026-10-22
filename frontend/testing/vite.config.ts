import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
export default defineConfig({plugins:[react()],base:'/apps/testing/',server:{port:5174,proxy:{'/api':'http://localhost:18080','/config.json':'http://localhost:18080'}},build:{chunkSizeWarningLimit:2000}});
