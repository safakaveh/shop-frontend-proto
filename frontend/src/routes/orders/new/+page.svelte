<script lang="ts">
	import { onMount } from 'svelte';
	import { enhance } from '$app/forms';

	let { form } = $props();

	let idempotencyKey = '';

	onMount(() => {
		idempotencyKey = crypto.randomUUID();
	});
</script>

<h1>ساخت سفارش</h1>

{#if form?.success}
	<p>سفارش ثبت شد: {form.orderId}</p>
{/if}

{#if form?.error}
	<p role="alert">{form.error}</p>
{/if}

<form method="POST" use:enhance>
	<input type="hidden" name="idempotencyKey" bind:value={idempotencyKey} />

	<label>
		شناسه محصول
		<input name="productId" required maxlength="128" />
	</label>

	<label>
		تعداد
		<input name="quantity" type="number" min="1" max="1000" required />
	</label>

	<button type="submit">ثبت سفارش</button>
</form>
