<script lang="ts">
  import MarketCard, { type Market } from '$lib/components/MarketCard.svelte';

  export let markets: Market[] = [];
</script>

{#if markets.length === 0}
  <div class="markets-cards markets-cards--empty">No markets to display.</div>
{:else}
  <div class="markets-cards">
    <div class="cards-grid">
      {#each markets as market}
        <MarketCard {market} />
      {/each}
    </div>
  </div>
{/if}

<style>
  .markets-cards {
    width: 100%;
  }

  .markets-cards--empty {
    font-size: 0.9rem;
    color: var(--text-subtle, #6b7280);
  }

  .cards-grid {
    display: grid;
    gap: 1rem;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  }
</style>

