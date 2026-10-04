import { env } from '$env/dynamic/private';
import { create, fromBinary, toBinary } from '@bufbuild/protobuf';

import {
	CreateOrderRequestSchema,
	CreateOrderResponseSchema,
	GetOrderResponseSchema
} from '$lib/gen/orders/v1/orders_pb';

export class GatewayError extends Error {
	constructor(
		message: string,
		public readonly status: number
	) {
		super(message);
	}
}

const PROTO = 'application/protobuf';

function gatewayURL(path: string): string {
	const base = env.GATEWAY_INTERNAL_URL;
	if (!base) throw new Error('GATEWAY_INTERNAL_URL is not configured');
	return `${base.replace(/\/+$/, '')}${path}`;
}

async function callGateway(
	path: string,
	options: {
		method: 'GET' | 'POST';
		body?: Uint8Array;
		idempotencyKey?: string;
	}
): Promise<Uint8Array> {
	if (!env.GATEWAY_SERVICE_TOKEN) {
		throw new Error('GATEWAY_SERVICE_TOKEN is not configured');
	}

	const headers: Record<string, string> = {
		Authorization: `Bearer ${env.GATEWAY_SERVICE_TOKEN}`,
		Accept: PROTO
	};

	if (options.body) headers['Content-Type'] = PROTO;
	if (options.idempotencyKey) headers['Idempotency-Key'] = options.idempotencyKey;

	let response: Response;
	try {
		response = await fetch(gatewayURL(path), {
			method: options.method,
			headers,
			body: options.body as BodyInit,
			signal: AbortSignal.timeout(2500)
		});
	} catch {
		throw new GatewayError('Gateway is unavailable', 502);
	}

	const bytes = new Uint8Array(await response.arrayBuffer());

	if (!response.ok) {
		let message = `Gateway returned ${response.status}`;
		try {
			const problem = JSON.parse(new TextDecoder().decode(bytes));
			if (typeof problem.detail === 'string') message = problem.detail;
		} catch {
			/* پیام عمومی حفظ می‌شود */
		}
		throw new GatewayError(message, response.status);
	}

	return bytes;
}


export async function createOrder(
	input: { userId: string; productId: string; quantity: number },
	idempotencyKey: string
) {
	const body = toBinary(CreateOrderRequestSchema, create(CreateOrderRequestSchema, input));

	const res = await callGateway('/internal/v1/orders', {
		method: 'POST',
		body,
		idempotencyKey
	});

	return fromBinary(CreateOrderResponseSchema, res);
}

export async function getOrder(orderId: string) {
	const res = await callGateway(
		`/internal/v1/orders/${encodeURIComponent(orderId)}`,
		{ method: 'GET' }
	);
	return fromBinary(GetOrderResponseSchema, res);
}
