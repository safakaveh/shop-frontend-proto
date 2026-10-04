<script lang="ts">
	let { data } = $props();
</script>

<h1>سفارش {data.order.orderId}</h1>
<ul>
	<li>محصول: {data.order.productId}</li>
	<li>تعداد: {data.order.quantity}</li>
	<li>وضعیت: {data.order.status}</li>
</ul>
