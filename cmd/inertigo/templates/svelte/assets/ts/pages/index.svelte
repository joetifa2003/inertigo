<script lang="ts">
  let { message }: { message: string } = $props();
  let count = $state(0);
</script>

<div style="padding: 20px; font-family: sans-serif;">
  <h1>Inertigo Example</h1>
  <p>Message from Go: <strong>{message}</strong></p>

  <div style="margin-top: 20px;">
    <p>Counter: {count}</p>
    <button onclick={() => count++}>Increment</button>
  </div>
</div>
