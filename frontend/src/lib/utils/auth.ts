// 🛡️ SLA: Domain Types
// We define the User interface here so our UI components have strict TypeScript safety.
export interface KariUser {
	id: string;
	email: string;
	rank: 'admin' | 'tenant' | 'viewer'; // Enforced ranks
	permissions: string[]; // e.g., ['apps:read', 'apps:write', 'domains:*']
}

/**
 * 🛡️ Core RBAC Engine: Evaluates if a user can execute a specific action.
 * Supports exact matches and wildcard grants.
 * * @param user The current user object from $page.data.user
 * @param requiredPermission The intent (e.g., 'apps:delete')
 * @returns boolean
 */
export function canPerform(user: KariUser | null | undefined, requiredPermission: string): boolean {
	// 1. Default Deny: If there is no user, they can do nothing.
	if (!user) return false;

	// 2. 🛡️ Admin Override: 'admin' rank automatically passes all UI checks.
	// (Note: The Go Brain still independently verifies this on the backend API)
	if (user.rank === 'admin') {
		return true;
	}

	// 3. Exact Match & Wildcard Resolution
	const [requiredResource, requiredAction] = requiredPermission.split(':');

	return user.permissions.some((grantedPermission) => {
		// Exact match (e.g., required 'apps:delete', granted 'apps:delete')
		if (grantedPermission === requiredPermission) {
			return true;
		}

		const [grantedResource, grantedAction] = grantedPermission.split(':');

		// Resource Wildcard (e.g., required 'apps:delete', granted 'apps:*')
		if (grantedResource === requiredResource && grantedAction === '*') {
			return true;
		}

		// Global Wildcard (e.g., granted '*:*')
		if (grantedResource === '*' && grantedAction === '*') {
			return true;
		}

		return false;
	});
}

/**
 * Convenience: Checks if a user has AT LEAST ONE of the required permissions.
 * Useful for rendering menu categories (e.g., "Settings" tab visible if user can manage billing OR users)
 */
export function canPerformAny(user: KariUser | null | undefined, permissions: string[]): boolean {
	if (!user) return false;
	return permissions.some(p => canPerform(user, p));
}

/**
 * Convenience: Checks if a user has ALL of the required permissions.
 * Useful for complex workflows (e.g., starting a deployment requires both app:write and domains:write)
 */
export function canPerformAll(user: KariUser | null | undefined, permissions: string[]): boolean {
	if (!user) return false;
	return permissions.every(p => canPerform(user, p));
}
