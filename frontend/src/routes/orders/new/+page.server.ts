import { fail } from '@sveltejs/kit';
import type { Actions } from './$types';

import { createOrder, GatewayError } from '$lib/server/gateway';

export const actions: Actions = {
	default: async ({ request, locals }) => {
		const form = await request.formData();

		// فرض: locals.user در hook.ts از session پر می‌شود.
		const userId = locals.user?.id;
		if (!userId) return fail(401, { error: 'ابتدا وارد حساب کاربری شوید' });

		const productId = String(form.get('productId') ?? '').trim();
		const quantity = Number(form.get('quantity'));
		const idempotencyKey = String(form.get('idempotencyKey') ?? '').trim();

		if (!productId || productId.length > 128)
			return fail(400, { error: 'شناسهٔ محصول نامعتبر است' });

		if (!Number.isInteger(quantity) || quantity < 1 || quantity > 1000)
			return fail(400, { error: 'تعداد نامعتبر است' });

		if (!idempotencyKey || idempotencyKey.length > 128)
			return fail(400, { error: 'کلید درخواست نامعتبر است' });

		try {
			const result = await createOrder({ userId, productId, quantity }, idempotencyKey);
			return { success: true, orderId: result.orderId };
		} catch (error) {
			if (error instanceof GatewayError && error.status >= 400 && error.status < 600) {
				return fail(error.status, { error: error.message });
			}
			return fail(502, { error: 'خطای ارتباط با سرویس سفارش' });
		}
	}
};
