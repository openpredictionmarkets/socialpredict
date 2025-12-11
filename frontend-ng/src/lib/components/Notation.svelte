<script lang="ts">
  export let text: string;
</script>

<div class="notation" role="note">
  <span class="notation__text">{text}</span>
</div>

<style>
  .notation {
    display: inline-flex;
    align-items: center;
    max-width: 100%;
    padding: 0.4rem 0.75rem;
    border-radius: 999px;
    background: transparent;
    color: var(--text, #0f172a);
    font-size: 0.85rem;
    font-weight: 500;
    line-height: 1.2;
  }

  .notation__text {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>
