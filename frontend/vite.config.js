import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { VitePWA } from 'vite-plugin-pwa';
export default defineConfig({
    plugins: [
        react(),
        VitePWA({
            registerType: 'autoUpdate',
            includeAssets: ['favicon.ico', 'robots.txt'],
            manifest: {
                name: 'CommunityShield',
                short_name: 'CommunityShield',
                description: 'Community Security Platform',
                theme_color: '#2563eb',
                background_color: '#ffffff',
                display: 'standalone',
                icons: [
                    {
                        src: 'icon-192x192.png',
                        sizes: '192x192',
                        type: 'image/png'
                    },
                    {
                        src: 'icon-512x512.png',
                        sizes: '512x512',
                        type: 'image/png'
                    }
                ]
            },
            workbox: {
                runtimeCaching: [
                    {
                        urlPattern: /^https:\/\/communityshield-backend\.onrender\.com\/api\/.*/i,
                        handler: 'NetworkFirst',
                        options: {
                            cacheName: 'api-cache',
                            expiration: {
                                maxEntries: 100,
                                maxAgeSeconds: 60 * 60 * 24 * 7
                            },
                            networkTimeoutSeconds: 10
                        }
                    },
                    {
                        urlPattern: /^https:\/\/tile\.openstreetmap\.org\/.*/i,
                        handler: 'CacheFirst',
                        options: {
                            cacheName: 'map-tiles',
                            expiration: {
                                maxEntries: 500,
                                maxAgeSeconds: 60 * 60 * 24 * 30
                            }
                        }
                    }
                ]
            }
        })
    ],
    server: {
        port: 5173,
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
        },
    },
    build: {
        rollupOptions: {
            output: {
                manualChunks: (id) => {
                    if (id.includes('react') || id.includes('react-dom') || id.includes('react-router-dom')) {
                        return 'react-vendor';
                    }
                    if (id.includes('leaflet') || id.includes('react-leaflet')) {
                        return 'map-vendor';
                    }
                    if (id.includes('chart.js') || id.includes('react-chartjs-2')) {
                        return 'chart-vendor';
                    }
                    if (id.includes('axios')) {
                        return 'ai-vendor';
                    }
                }
            }
        },
        chunkSizeWarningLimit: 1000,
        minify: 'terser',
    },
});
