import type { Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

export const handle: Handle = async ({ event, resolve }) => {
	const accessToken = event.cookies.get('kari_access_token');

	// 1. Initial State: Unauthenticated
	event.locals.user = null;

	if (accessToken) {
		try {
			// 2. Platform Agnostic Verification
			// We call the Go Brain's /api/v1/auth/me or verify endpoint
			// This ensures the Rank is fresh and the account isn't suspended.
			const response = await fetch(`${env.KARI_API_URL}/auth/me`, {
				headers: {
					Authorization: `Bearer ${accessToken}`
				}
			});

			if (response.ok) {
				const userData = await response.json();
				
				// 🛡️ 3. SLA: Populate locals with Rank and Permissions
				// These values are now accessible in every +page.server.ts
				event.locals.user = {
					id: userData.id,
					email: userData.email,
					rank: userData.rank, // Critical for UI hierarchy
					permissions: userData.permissions // Array of "resource:action"
				};
			} else {
				// Token expired or user banned - clear the cookie
				event.cookies.delete('kari_access_token', { path: '/' });
			}
		} catch (err) {
			console.error('Auth Bridge Failure:', err);
		}
	}

	// 🛡️ 4. Zero-Trust Protected Routes
	// Prevent unauthorized access to the /app path at the hook level
	if (event.url.pathname.startsWith('/app') && !event.locals.user) {
		return new Response('Redirecting', {
			status: 303,
			headers: { Location: '/auth/login' }
		});
	}

	return resolve(event);
};
