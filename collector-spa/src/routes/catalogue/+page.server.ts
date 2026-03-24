import { loadProducts } from '$lib/server/products';

export const load = async ({ fetch }) => {
	const products = await loadProducts(fetch);

	return { products };
};
