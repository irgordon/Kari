// frontend/src/routes/(app)/dashboard/+page.ts
import type { PageLoad } from './$types';

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

export const load: PageLoad = async ({ fetch }) => {
    // Environment Agnostic: API location injected at build/runtime
    const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

    try {
        // Fetch only active, unresolved alerts
        const response = await fetch(`${apiUrl}/api/v1/audit/alerts?status=unresolved`);

        if (!response.ok) {
            throw new Error('Failed to load system alerts');
        }

        const alerts: SystemAlert[] = await response.json();

        return {
            alerts
        };
    } catch (error) {
        console.error('[Dashboard Load] Error fetching alerts:', error);
        return {
            alerts: [] // Graceful degradation to empty state
        };
    }
};
