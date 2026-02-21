import type { PageServerLoad } from './$types';
import { brainFetch } from '$lib/server/api';

// ==============================================================================
// 1. SLA Type Definitions
// ==============================================================================

export interface SystemAlert {
    id: string;
    severity: 'info' | 'warning' | 'critical';
    category: 'ssl' | 'system' | 'security' | 'deployment';
    message: string;
    created_at: string;
}

export const load: PageServerLoad = async ({ cookies }) => {
    try {
        // 2. 🛡️ Zero-Trust & SLA: Use our hardened internal fetcher.
        // This keeps traffic off the public internet, routes via the Docker backplane,
        // and automatically securely extracts and forwards the HttpOnly JWT.
        const response = await brainFetch('/api/v1/audit/alerts?status=unresolved', {}, cookies);

        const alerts: SystemAlert[] = await response.json();

        return {
            alerts
        };
    } catch (error) {
        // 3. 🛡️ Privacy & Reliability: Log technically, degrade gracefully.
        // The user shouldn't see a raw 500 error page just because the alerts 
        // service is slow. We log it for the admin and return an empty array to the UI.
        console.error('🚨 [Dashboard Load] Failed to fetch system alerts:', error);
        
        return {
            alerts: [] // The UI will simply show "No active alerts"
        };
    }
};
