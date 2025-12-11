<script lang="ts">
  'use runes';

  import type { Branding } from '$lib/config/branding';

  const { branding } = $props<{ branding: Branding }>();

  let open = $state(false);
  const prettyJson = $derived(JSON.stringify(branding, null, 2));
</script>

<div class="brand-debug">
  <button class="toggle" type="button" onclick={() => (open = !open)}>
    {open ? 'Hide branding data' : 'Show branding data'}
  </button>

  {#if open}
    <div class="panel">
      <pre><code>{prettyJson}</code></pre>
    </div>
  {/if}
</div>

<style>
  .brand-debug {
    margin-top: 1rem;
  }

  .toggle {
    padding: 0.4rem 0.8rem;
    border-radius: 0.5rem;
    border: 1px solid var(--border, rgba(148, 163, 184, 0.4));
    background: var(--panel, rgba(0, 0, 0, 0.02));
    color: var(--text, #0e0f14);
    font-size: 0.85rem;
    cursor: pointer;
  }

  .toggle:hover {
    border-color: var(--color-primary, #2563eb);
  }

  .panel {
    margin-top: 0.6rem;
    padding: 0.75rem;
    border-radius: 0.6rem;
    background: var(--panel, rgba(0, 0, 0, 0.02));
    border: 1px solid var(--border, rgba(148, 163, 184, 0.4));
    max-height: 640px;
    overflow: auto;
  }

  pre {
    margin: 0;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono',
      'Courier New', monospace;
    font-size: 0.8rem;
    white-space: pre;
  }
</style>

