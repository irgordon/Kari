import { redirect, type Handle } from '@sveltejs/kit';
import * as jose from 'jose';
import { env } from '$env/dynamic/private';

// 🛡️ Zero-Trust: Strictly defined asset prefixes to prevent bypass via dots in filenames
const ASSET_PREFIXES = ['/_app/', '/favicon.ico', '/static/'];
const PUBLIC_ROUTES = ['/login', '/health'];

export const handle: Handle = async ({ event, resolve }) => {
    const { pathname } = event.url;

    // 1. 🛡️ Performance: High-speed bypass for verified static assets
    if (ASSET_PREFIXES.some(prefix => pathname.startsWith(prefix))) {
        return await resolve(event);
    }

    let accessToken = event.cookies.get('kari_access_token');
    const refreshToken = event.cookies.get('kari_refresh_token');

    // 2. 🔄 Hardened Silent Refresh Pipeline
    if (!accessToken && refreshToken) {
        try {
            const response = await event.fetch(`${env.INTERNAL_API_URL}/api/v1/auth/refresh`, {
                method: 'POST',
                headers: { 'Cookie': `kari_refresh_token=${refreshToken}` }
            });

            if (response.ok) {
                // 🛡️ SLA: Proxy all Set-Cookie headers from Brain to Browser correctly
                const setCookieHeaders = response.headers.getSetCookie();
                setCookieHeaders.forEach((cookie) => {
                    // This ensures all attributes (Secure, HttpOnly, SameSite) are preserved
                    event.setHeaders({ 'Set-Cookie': cookie });
                });

                // Re-extract for immediate locals population
                const cookies = response.headers.get('Set-Cookie');
                accessToken = cookies?.split(';')
                    .find(c => c.trim().startsWith('kari_access_token='))
                    ?.split('=')[1];
            } else {
                event.cookies.delete('kari_access_token', { path: '/' });
                event.cookies.delete('kari_refresh_token', { path: '/' });
            }
        } catch (err) {
            console.error("🚨 [SLA FATAL] Kari Brain Offline during refresh:", err);
        }
    }

    // 3. 🛡️ Cryptographic Verification (Strict Algorithm Enforcement)
    event.locals.user = null;

    if (accessToken) {
        try {
            const secret = new TextEncoder().encode(env.JWT_SECRET);
            const { payload } = await jose.jwtVerify(accessToken, secret, {
                algorithms: ['HS256'],
                issuer: 'kari:brain',
                audience: 'kari:panel'
            });

            event.locals.user = {
                id: payload.sub as string,
                role: payload.role as 'admin' | 'tenant'
            };
        } catch (error) {
            event.cookies.delete('kari_access_token', { path: '/' });
        }
    }

    // 4. 🛡️ Hardened Route Guarding
    const isAuthRoute = PUBLIC_ROUTES.includes(pathname);
    const isDashboard = pathname.startsWith('/dashboard') || pathname === '/';

    if (!event.locals.user && !isAuthRoute) {
        throw redirect(303, '/login');
    }

    if (event.locals.user && isAuthRoute) {
        throw redirect(303, '/dashboard');
    }

    // 5. 🛡️ SLA: Defense-in-Depth Headers
    const response = await resolve(event);
    
    // Protect against clickjacking and MIME-sniffing
    response.headers.set('X-Frame-Options', 'DENY');
    response.headers.set('X-Content-Type-Options', 'nosniff');
    response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
    
    // Strict Transport Security (1 year)
    response.headers.set('Strict-Transport-Security', 'max-age=31536000; includeSubDomains; preload');

    // 🛡️ Content Security Policy: Prevent XSS and unauthorized data exfiltration
    response.headers.set(
        'Content-Security-Policy',
        "default-src 'self'; " +
        "script-src 'self' 'unsafe-inline'; " + // xterm.js might need inline for some themes
        "style-src 'self' 'unsafe-inline'; " +
        "connect-src 'self' ws: wss:; " + // Allow WebSockets/SSE for telemetry
        "img-src 'self' data:; " +
        "frame-ancestors 'none';"
    );

    return response;
};
