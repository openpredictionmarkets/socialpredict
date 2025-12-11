import type { TutorialStep } from '$lib/components/TutorialModalOld.svelte';

export const defaultTutorialSteps: TutorialStep[] = [
  {
    title: 'A prediction market works like this…',
    description: [
      '<p class="spaced">You buy YES/NO shares on real-world events.</p>',
      '<p class="spaced">All markets start at 50¢/50¢ representing equal probability for YES and NO outcomes.</p>',
      '<p class="spaced">Share prices move with demand, reflecting the crowd’s collective forecast.</p>',
      '<p>This is the "<em>Wisdom of the Crowd</em>" in action.</p>'
    ].join('\n'),
    note: [
      '<p>Tip: Share prices represent outcome probabilities. For example, if the market price for YES is 62¢, the market expects that there is a ~62% chance it will resolve YES.</p>'
    ].join('\n')
  },
  {
    title: 'Tour a market…',
    description: [
      {
        key: 'category',
        label: 'Market type',
        detail: 'Which arena this market belongs to (Politics, Tech, Sports, etc).'
      },
      {
        key: 'resolution',
        label: 'Resolution time',
        detail: 'When the market will settle and pay out.'
      },
      {
        key: 'question',
        label: 'Market question',
        detail: 'The exact phrasing of what resolves YES vs NO.'
      },
      {
        key: 'prices',
        label: 'Prices',
        detail: 'YES/NO quotes show implied probability; they change with trades.'
      },
      {
        key: 'trend',
        label: 'Trend',
        detail: 'Recent momentum signal to see which outcome is heating up.'
      },
      {
        key: 'liquidity',
        label: 'Liquidity',
        detail: 'Indicates how well funded the market is and how many participants are active within it.'
      },
      {
        key: 'community',
        label: 'Community',
        detail: 'How many predictions/comments—shows activity and interest.'
      },
      {
        key: 'cta',
        label: 'Action',
        detail: 'How users place trades.'
      }
    ],
    note: 'Hover over each bullet to highlight its feature on the market card.'
  },
  {
    title: 'See how prices change…',
    description: [
      '<p class="spaced">Drag the sliders to see how market prices react to trades by users.</p>',
      '<p class="spaced">Notice how buying more shares of one outcome pushes its price up, while the opposing outcome\'s price drops accordingly.</p>',
      '<p>Prices are determined by supply and demand. The more users buy YES shares, the higher the price for YES becomes, indicating increased confidence in that outcome.</p>'
    ].join('\n'),
    note: 'Tip: If you were actually trading, your actual cost would be set by the price at the time of your trade.'
  },
  {
    title: 'Watch the market resolve — rewards arrive instantly…',
    description: [
      '<p class="spaced">When an event settles (when its market "resolves"), shareholders are paid out.</p>',
      '<p class="spaced">If you hold shares of the winning outcome, you receive 100% value for each share you own.</p>',
      '<p class="spaced">If the market resolves to "YES", holders of YES shares are paid $1 (100% value) for each of their shares.</p>',
      '<p class="spaced">Losing shares become worthless.</p>',
      '<p class="spaced">Funds are credited to your account instantly, ready for withdrawal or reinvestment.</p>',
      '<p>Let\'s assume you bought your shares for the current price of your chosen outcome.</p>'
    ].join('\n'),
    note: [
      'Press the "Resolve Market to YES" button to see how your payout is calculated based on your purchase price versus the resolution outcome.'
    ].join('\n')
  },
  {
    title: 'Join the SocialPredict community…',
    description:
      'Follow top forecasters, join prediction threads, and see sentiment move in real time.'
  }
];

