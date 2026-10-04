// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		interface User {
			id: string;
			// فیلدهای اختیاری دیگر مثل:
			// email?: string;
			// role?: string;
		}

		interface Locals {
			user?: User; // یا user: User | null; بسته به پیاده‌سازی hook.server.ts
		}
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
