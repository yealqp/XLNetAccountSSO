export function resolveNextTarget(nextValue: unknown) {
	const next = Array.isArray(nextValue) ? nextValue[0] : nextValue

	if (typeof next === 'string' && next.startsWith('/') && !next.startsWith('//')) {
		return next
	}

	return '/admin'
}

export function buildNextQuery(nextValue: unknown) {
	if (typeof nextValue === 'string') {
		return { next: nextValue }
	}

	if (Array.isArray(nextValue)) {
		const nextItems = nextValue.filter((item): item is string => typeof item === 'string')
		return nextItems.length > 0 ? { next: nextItems } : {}
	}

	return {}
}
