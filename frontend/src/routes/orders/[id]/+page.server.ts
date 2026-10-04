import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

import { getOrder, GatewayError } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ params }) => {
	try {
		return { order: await getOrder(params.id) };
	} catch (e) {
		if (e instanceof GatewayError && e.status === 404) {
			error(404, 'سفارش پیدا نشد');
		}
		error(502, 'خطای ارتباط با سرویس سفارش');
	}
};
