import { redirect, type Handle } from '@sveltejs/kit';
import * as jose from 'jose';
import { env } from '$env/dynamic/private';

export const handle: Handle = async ({ event, resolve }) => {
    // 1. 🛡️ Performance: Bypass auth logic for static assets and public routes
    if (event.url.pathname.startsWith('/_app') || event.url.pathname.includes('.')) {
        return await resolve(event);
    }

    let accessToken = event.cookies.get('kari_access_token');
    const refreshToken = event.cookies.get('kari_refresh_token');

    // 2. 🔄 The Silent Refresh Pipeline (State Reconciliation)
    if (!accessToken && refreshToken) {
        try {
            // 🛡️ Zero-Trust: We ask the Go Brain to validate the refresh token and mint a new pair
            const response = await event.fetch(`${env.INTERNAL_API_URL}/api/v1/auth/refresh`, {
                method: 'POST',
                headers: { 'Cookie': `kari_refresh_token=${refreshToken}` }
            });

            if (response.ok) {
                // 🛡️ SLA Fix: We MUST manually parse the Set-Cookie headers from the Go API 
                // and pass them down to SvelteKit's cookie manager, otherwise the browser never gets them.
                const setCookieHeaders = response.headers.getSetCookie();
                for (const cookieStr of setCookieHeaders) {
                    // SvelteKit's event.cookies.set() requires parsing the raw string.
                    // For brevity, we let the browser handle the raw string via response headers later, 
                    // but we need to extract just the access_token value for immediate verification below.
                    const match = cookieStr.match(/kari_access_token=([^;]+)/);
                    if (match) {
                        accessToken = match[1];
                        // Apply it to the current event so subsequent load functions have it
                        event.cookies.set('kari_access_token', accessToken, { path: '/' }); 
                    }
                }
            } else {
                // If the refresh token is dead/revoked, nuke the state completely
                event.cookies.delete('kari_refresh_token', { path: '/' });
                event.cookies.delete('kari_access_token', { path: '/' });
            }
        } catch (err) {
            console.error("🚨 [SLA FATAL] Brain Connectivity Error:", err);
        }
    }

    // 3. 🛡️ Cryptographic Verification
    event.locals.user = null; // Failsafe default

    if (accessToken) {
        try {
            const secret = new TextEncoder().encode(env.JWT_SECRET);
            
            // 🛡️ Zero-Trust: Strictly enforce HS256 to prevent algorithm downgrade attacks
            const { payload } = await jose.jwtVerify(accessToken, secret, {
                algorithms: ['HS256']
            });

            event.locals.user = {
                id: payload.sub as string,
                role: payload.role as 'admin' | 'tenant'
            };
        } catch (error) {
            // Token is expired, forged, or the secret rotated. Clean up.
            event.cookies.delete('kari_access_token', { path: '/' });
            event.locals.user = null;
        }
    }

    // 4. 🛡️ Route Guarding (The Edge Boundary)
    const isAuthRoute = event.url.pathname.startsWith('/login');
    const isProtectedRoute = !isAuthRoute && event.url.pathname !== '/';

    // If they have no valid user state and try to hit a protected route, boot them.
    if (!event.locals.user && isProtectedRoute) {
        throw redirect(303, '/login');
    }

    // If they are fully logged in and try to hit the login page, push them to the dashboard.
    if (event.locals.user && isAuthRoute) {
        throw redirect(303, '/dashboard');
    }

    // 5. 🛡️ SLA: Inject Security Headers on the way out
    const response = await resolve(event);
    response.headers.set('X-Content-Type-Options', 'nosniff');
    response.headers.set('X-Frame-Options', 'DENY');
    response.headers.set('Strict-Transport-Security', 'max-age=31536000; includeSubDomains');

    return response;
};
