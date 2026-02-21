import type { Handle } from '@sveltejs/kit';
import * as jose from 'jose'; // 🛡️ Use jose for server-side verification
import { env } from '$env/dynamic/private';

export const handle: Handle = async ({ event, resolve }) => {
	// 🛡️ Performance: Skip auth logic for static assets or internal SvelteKit calls
	if (event.url.pathname.startsWith('/_app') || event.url.pathname.includes('.')) {
		return await resolve(event);
	}

	let accessToken = event.cookies.get('kari_access_token');
	const refreshToken = event.cookies.get('kari_refresh_token');

	// 1. Silent Refresh Logic
	if (!accessToken && refreshToken) {
		try {
			// Ask the Go API for a new token pair
			const response = await event.fetch(`${env.INTERNAL_API_URL}/api/v1/auth/refresh`, {
				method: 'POST',
				// Forwarding the refresh cookie
				headers: { 'Cookie': `kari_refresh_token=${refreshToken}` }
			});

			if (response.ok) {
				// 🛡️ SvelteKit's fetch automatically handles 'Set-Cookie' 
				// if configured, but manual proxying is safer for cross-domain.
				accessToken = event.cookies.get('kari_access_token');
			} else {
				event.cookies.delete('kari_refresh_token', { path: '/' });
				event.cookies.delete('kari_access_token', { path: '/' });
			}
		} catch (err) {
			console.error("Brain Connectivity Error:", err);
		}
	}

	// 2. 🛡️ Secure Verification
	if (accessToken) {
		try {
			const secret = new TextEncoder().encode(env.JWT_SECRET);
			const { payload } = await jose.jwtVerify(accessToken, secret);

			// Inject verified data into locals
			event.locals.user = {
				id: payload.sub as string,
				role: payload.role as 'admin' | 'tenant'
			};
		} catch (error) {
			// Token invalid, expired, or tampered with
			event.locals.user = null;
		}
	} else {
		event.locals.user = null;
	}

	return await resolve(event);
};
