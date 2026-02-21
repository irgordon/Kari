// frontend/src/lib/utils/auth.ts

/**
 * hasAuthority checks if a user's rank is sufficient.
 * Remember: In Karı, LOWER rank number = HIGHER authority.
 */
export function hasAuthority(userRank: number, requiredRank: number): boolean {
	return userRank <= requiredRank;
}

export function canPerform(userPermissions: string[], action: string): boolean {
	return userPermissions.includes(action);
}
