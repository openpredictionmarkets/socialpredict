<script lang="ts">
  import type { Market } from '$lib/components/MarketCard.svelte';

  /**
   * Fields that can be displayed for each market.
   * You can restrict the columns via the `filter` prop:
   *   filter="title,yes,no"
   */
  const ALL_FIELDS = [
    'title',
    'yes',
    'no',
    'trend',
    'liquidity',
    'category',
    'resolves',
    'community'
  ] as const;

  type MarketField = (typeof ALL_FIELDS)[number];

  export let markets: Market[] = [];
  /**
   * Optional comma-separated list of fields to display.
   * Example: "title,yes,no,trend"
   * Defaults to all fields when omitted or empty.
   */
  export let filter: string | undefined;

  const fieldLabels: Record<MarketField, string> = {
    title: 'Market',
    yes: 'YES',
    no: 'NO',
    trend: 'Trend',
    liquidity: 'Liquidity',
    category: 'Category',
    resolves: 'Resolves',
    community: 'Community'
  };

  function parseFilter(value: string | undefined): MarketField[] {
    if (!value || !value.trim()) return [...ALL_FIELDS];
    const requested = value
      .split(',')
      .map((f) => f.trim())
      .filter(Boolean);

    const active: MarketField[] = [];
    for (const name of requested) {
      if ((ALL_FIELDS as readonly string[]).includes(name) && !active.includes(name as MarketField)) {
        active.push(name as MarketField);
      }
    }
    return active.length > 0 ? active : [...ALL_FIELDS];
  }

  const columns: MarketField[] = parseFilter(filter);

  function cellClass(field: MarketField, market: Market): string {
    if (field === 'trend') {
      return market.trend >= 0 ? 'trend trend-up' : 'trend trend-down';
    }
    return field;
  }

  function formatCell(field: MarketField, market: Market): string {
    switch (field) {
      case 'title':
        return market.title;
      case 'yes':
        return `${market.yes}¢`;
      case 'no':
        return `${market.no}¢`;
      case 'trend': {
        const sign = market.trend >= 0 ? '▲' : '▼';
        return `${sign} ${Math.abs(market.trend)}%`;
      }
      case 'liquidity':
        return market.liquidity;
      case 'category':
        return market.category;
      case 'resolves':
        return market.resolves;
      case 'community':
        return market.community;
      default:
        return '';
    }
  }
</script>

{#if markets.length === 0}
  <div class="markets-list markets-list--empty">No markets to display.</div>
{:else}
  <div class="markets-list">
    <table>
      <thead>
        <tr>
          {#each columns as field}
            <th scope="col">{fieldLabels[field]}</th>
          {/each}
        </tr>
      </thead>
      <tbody>
        {#each markets as market}
          <tr>
            {#each columns as field}
              <td class={cellClass(field, market)}>{formatCell(field, market)}</td>
            {/each}
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<style>
  .markets-list {
    width: 100%;
    overflow-x: auto;
  }

  .markets-list--empty {
    font-size: 0.9rem;
    color: var(--text-subtle, #6b7280);
  }

  .markets-list td.yes {
    color: var(--color-accent, #16a34a);
    background: var(--color-accent-soft, rgba(22, 163, 74, 0.14));
    font-weight: 700;
  }

  .markets-list td.no {
    color: var(--color-danger, #dc2626);
    background: var(--color-danger-soft, rgba(220, 38, 38, 0.15));
    font-weight: 700;
  }

  .markets-list td.trend {
    font-weight: 700;
  }

  .markets-list td.trend-up {
    color: var(--color-accent, #16a34a);
  }

  .markets-list td.trend-down {
    color: var(--color-danger, #dc2626);
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }

  thead {
    background: var(--panel, rgba(0, 0, 0, 0.02));
  }

  th,
  td {
    padding: 0.5rem 0.6rem;
    text-align: left;
    border-bottom: 1px solid var(--border, rgba(148, 163, 184, 0.35));
    white-space: nowrap;
  }

  th {
    font-weight: 600;
    color: var(--text-subtle, #4b5563);
  }

  tbody tr:hover {
    background: var(--panel, rgba(0, 0, 0, 0.03));
  }
</style>
