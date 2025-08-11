import React, { useState } from 'react';
import {
  BetYesButton,
  BetNoButton,
} from '../../components/buttons/trade/BetButtons';
import {
  ResolveButton,
  SelectNoButton,
  SelectYesButton,
  ConfirmResolveButton,
} from '../../components/buttons/marketDetails/ResolveButtons';
import SiteButton from '../../components/buttons/SiteButtons';
import SiteTabs from '../../components/tabs/SiteTabs';
import Sidebar from '../../components/sidebar/Sidebar';
import Header from '../../components/header/Header';
import {
  RegularInput,
  SuccessInput,
  ErrorInput,
  PersonInput,
  LockInput,
} from '../../components/inputs/InputBar';
import RegularInputBox from '../../components/inputs/InputBox';
import DatetimeSelector from '../../components/datetimeSelector/DatetimeSelector';
import LoginModalButton from '../../components/modals/login/LoginModalClick';
import MarketsTable from '../../components/tables/MarketTables';
import MarketChart from '../../components/charts/MarketChart';
import TestMarketData from '../../tests/TestData';
import LoadingSpinner from '../../components/loaders/LoadingSpinner';
import { SharesBadge } from '../../components/buttons/trade/SellButtons';
import ExpandableText from '../../components/utils/ExpandableText';
import ExpandableLink from '../../components/utils/ExpandableLink';

const Style = () => {
  const [isSelected, setIsSelected] = useState(false);

  const tabsData = [
    { label: 'Comments', content: <div>Comments Go here...</div> },
    { label: 'Positions', content: <div>Positions Go here...</div> },
    { label: 'Bets', content: <div>Bets go here...</div> },
  ];

  return (
    <div className='overflow-auto'>
      <Header />
      
      {/* Responsive Grids (Spec) Section */}
      <div className='bg-primary-background p-6'>
        <h2 className='text-2xl font-bold text-white mb-6'>Responsive Grids (Spec)</h2>
        <p className='text-gray-300 mb-4'>
          Mobile-responsive grid patterns used throughout SocialPredict. These components adapt from mobile-first 
          (iPhone 12 mini @ 375px) to desktop layouts using Tailwind's responsive utilities.
        </p>
        
        {/* Import Documentation */}
        <div className='bg-gray-900 p-4 rounded-lg mb-8'>
          <h3 className='text-lg font-semibold text-white mb-3'>Component Imports</h3>
          <div className='space-y-3'>
            <div>
              <div className='text-emerald-400 text-sm font-medium mb-1'>Bets Grid Component</div>
              <code className='text-xs text-gray-300 bg-gray-800 px-2 py-1 rounded block'>
                import BetsActivity from '../components/layouts/activity/bets/BetsActivity';
              </code>
              <div className='text-gray-400 text-xs mt-1'>Usage: {`<BetsActivity marketId={marketId} />`}</div>
            </div>
            <div>
              <div className='text-emerald-400 text-sm font-medium mb-1'>Leaderboard Grid Component</div>
              <code className='text-xs text-gray-300 bg-gray-800 px-2 py-1 rounded block'>
                import LeaderboardActivity from '../components/layouts/activity/leaderboard/LeaderboardActivity';
              </code>
              <div className='text-gray-400 text-xs mt-1'>Usage: {`<LeaderboardActivity marketId={marketId} />`}</div>
            </div>
          </div>
        </div>
        
        {/* Bets Grid Demo */}
        <div className='mb-8'>
          <h3 className='text-xl font-semibold text-white mb-4'>Bets Grid</h3>
          <div className='bg-gray-900 p-4 rounded-lg'>
            <div className="sp-grid-bets-header">
              <div>Username</div>
              <div className="text-center">Outcome</div>
              <div className="text-right">Amount</div>
              <div className="text-right">After</div>
              <div className="text-right">Placed</div>
            </div>
            
            <div className="sp-grid-bets-row mt-2">
              <div className="sp-cell-username">
                <div className="sp-ellipsis text-xs sm:text-sm font-medium">
                  <span className="text-blue-500">alice_trader</span>
                </div>
              </div>
              <div className="justify-self-start sm:justify-self-center">
                <span className="px-2 py-1 rounded text-xs font-bold bg-green-600 text-white">YES</span>
              </div>
              <div className="sp-cell-num text-xs sm:text-sm text-gray-300">250</div>
              <div className="hidden sm:block sp-cell-num text-gray-300">0.724</div>
              <div className="col-span-3 sm:col-span-1 text-right sp-subline">
                Jan 8, 2025, 3:45 PM
              </div>
            </div>
            
            <div className="sp-grid-bets-row mt-2">
              <div className="sp-cell-username">
                <div className="sp-ellipsis text-xs sm:text-sm font-medium">
                  <span className="text-blue-500">bob_predictor_longusername</span>
                </div>
              </div>
              <div className="justify-self-start sm:justify-self-center">
                <span className="px-2 py-1 rounded text-xs font-bold bg-red-600 text-white">NO</span>
              </div>
              <div className="sp-cell-num text-xs sm:text-sm text-gray-300">100</div>
              <div className="hidden sm:block sp-cell-num text-gray-300">0.276</div>
              <div className="col-span-3 sm:col-span-1 text-right sp-subline">
                Jan 8, 2025, 2:30 PM
              </div>
            </div>
          </div>
          <div className='text-xs text-gray-400 mt-2'>
            <strong>Mobile (≤639px):</strong> 3 columns - Username, Outcome, Amount (timestamp spans full width)<br/>
            <strong>Desktop (≥640px):</strong> 5 columns - adds After probability & dedicated Placed column
          </div>
        </div>

        {/* Leaderboard Grid Demo */}
        <div className='mb-8'>
          <h3 className='text-xl font-semibold text-white mb-4'>Leaderboard Grid</h3>
          <div className='bg-gray-900 p-4 rounded-lg'>
            <div className="sp-grid-leaderboard-header">
              <div>Rank</div>
              <div>User</div>
              <div>Position</div>
              <div className="text-right">Profit</div>
              <div className="text-right">Current Value</div>
              <div className="text-right">Total Spent</div>
              <div>Shares</div>
            </div>
            
            <div className="sp-grid-leaderboard-row mt-2">
              <div className="flex items-center justify-start">
                <div className="text-white font-bold text-lg mr-2">🥇</div>
                <div className="sm:hidden sp-cell-username">
                  <div className="sp-ellipsis text-xs font-medium">
                    <span className="text-blue-500">alice_trader</span>
                  </div>
                </div>
              </div>
              <div className="hidden sm:block sp-cell-username">
                <div className="sp-ellipsis font-medium">
                  <span className="text-blue-500">alice_trader</span>
                </div>
              </div>
              <div className="hidden sm:block">
                <span className="px-2 py-1 rounded text-xs font-bold bg-green-600 text-white">YES</span>
              </div>
              <div className="text-right">
                <div className="font-bold text-sm text-green-400">+1,250</div>
                <div className="sm:hidden sp-subline">Pos YES • 15Y 3N</div>
              </div>
              <div className="hidden sm:block sp-cell-num text-gray-300">2,100</div>
              <div className="hidden sm:block sp-cell-num text-gray-300">850</div>
              <div className="hidden sm:block text-gray-300 text-xs">
                <div>YES: 15</div>
                <div>NO: 3</div>
              </div>
            </div>
            
            <div className="sp-grid-leaderboard-row mt-2">
              <div className="flex items-center justify-start">
                <div className="text-white font-bold text-lg mr-2">🥈</div>
                <div className="sm:hidden sp-cell-username">
                  <div className="sp-ellipsis text-xs font-medium">
                    <span className="text-blue-500">bob_predictor</span>
                  </div>
                </div>
              </div>
              <div className="hidden sm:block sp-cell-username">
                <div className="sp-ellipsis font-medium">
                  <span className="text-blue-500">bob_predictor</span>
                </div>
              </div>
              <div className="hidden sm:block">
                <span className="px-2 py-1 rounded text-xs font-bold bg-red-600 text-white">NO</span>
              </div>
              <div className="text-right">
                <div className="font-bold text-sm text-red-400">-75</div>
                <div className="sm:hidden sp-subline">Pos NO • 2Y 8N</div>
              </div>
              <div className="hidden sm:block sp-cell-num text-gray-300">675</div>
              <div className="hidden sm:block sp-cell-num text-gray-300">750</div>
              <div className="hidden sm:block text-gray-300 text-xs">
                <div>YES: 2</div>
                <div>NO: 8</div>
              </div>
            </div>
          </div>
          <div className='text-xs text-gray-400 mt-2'>
            <strong>Mobile (≤639px):</strong> 2 columns - Rank+Username combined, Profit+subline info<br/>
            <strong>Desktop (≥640px):</strong> 7 columns - separate Rank, User, Position, Profit, Current Value, Total Spent, Shares
          </div>
        </div>

        {/* CSS Classes Reference */}
        <div className='mb-8'>
          <h3 className='text-xl font-semibold text-white mb-4'>CSS Classes</h3>
          <div className='bg-gray-900 p-4 rounded-lg font-mono text-xs text-gray-300'>
            <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
              <div>
                <div className='text-emerald-400 mb-2'>/* Bets Grid */</div>
                <div>.sp-grid-bets-header</div>
                <div>.sp-grid-bets-row</div>
                <div className='mt-4 text-emerald-400 mb-2'>/* Leaderboard Grid */</div>
                <div>.sp-grid-leaderboard-header</div>
                <div>.sp-grid-leaderboard-row</div>
              </div>
              <div>
                <div className='text-emerald-400 mb-2'>/* Shared Utilities */</div>
                <div>.sp-cell-username</div>
                <div>.sp-cell-num</div>
                <div>.sp-chip</div>
                <div>.sp-subline</div>
                <div>.sp-tight</div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      {/* Color Palette Section */}
      <div className='bg-primary-background p-6'>
        <h2 className='text-2xl font-bold text-white mb-6'>SocialPredict Color Palette</h2>
        <div className='grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-8'>
          {/* Primary Colors */}
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-primary-background rounded mb-2 border border-gray-500'></div>
            <div className='text-white text-sm font-medium'>primary-background</div>
            <div className='text-gray-400 text-xs'>#0e121d</div>
          </div>
          
          {/* Gray Colors */}
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-custom-gray-verylight rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>custom-gray-verylight</div>
            <div className='text-gray-400 text-xs'>#DBD4D3</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-custom-gray-light rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>custom-gray-light</div>
            <div className='text-gray-400 text-xs'>#67697C</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-custom-gray-dark rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>custom-gray-dark</div>
            <div className='text-gray-400 text-xs'>#303030</div>
          </div>
          
          {/* Button Colors */}
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-green-btn rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>green-btn</div>
            <div className='text-gray-400 text-xs'>#054A29</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-red-btn rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>red-btn</div>
            <div className='text-gray-400 text-xs'>#D00000</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-gold-btn rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>gold-btn</div>
            <div className='text-gray-400 text-xs'>#FFC107</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-neutral-btn rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>neutral-btn</div>
            <div className='text-gray-400 text-xs'>#8A1C7C</div>
          </div>
          
          {/* Special Colors */}
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-beige rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>beige</div>
            <div className='text-gray-400 text-xs'>#F9D3A5</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-primary-pink rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>primary-pink</div>
            <div className='text-gray-400 text-xs'>#F72585</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-info-blue rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>info-blue</div>
            <div className='text-gray-400 text-xs'>#17a2b8</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-warning-orange rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>warning-orange</div>
            <div className='text-gray-400 text-xs'>#ffc107</div>
          </div>
          
          {/* Outcome Badge Colors */}
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-green-600 rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>green-600 (YES)</div>
            <div className='text-gray-400 text-xs'>#059669</div>
            <div className='text-gray-500 text-xs mt-1'>Used for YES outcome badges</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-red-600 rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>red-600 (NO)</div>
            <div className='text-gray-400 text-xs'>#dc2626</div>
            <div className='text-gray-500 text-xs mt-1'>Used for NO outcome badges</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-red-btn rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>red-btn (Button)</div>
            <div className='text-gray-400 text-xs'>#D00000</div>
            <div className='text-gray-500 text-xs mt-1'>Used for buttons & charts</div>
          </div>
        </div>
        
        {/* Chart Colors Section */}
        <h3 className='text-xl font-bold text-white mb-4 mt-8'>Chart Colors</h3>
        <div className='grid grid-cols-1 md:grid-cols-3 gap-4 mb-8'>
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-info-blue rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>Default Probability</div>
            <div className='text-gray-400 text-xs'>info-blue: #17a2b8</div>
            <div className='text-gray-500 text-xs mt-1'>Used for single probability line</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-green-btn rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>YES Probability</div>
            <div className='text-gray-400 text-xs'>green-btn: #054A29</div>
            <div className='text-gray-500 text-xs mt-1'>Used when showing dual probabilities</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='w-full h-16 bg-red-btn rounded mb-2'></div>
            <div className='text-white text-sm font-medium'>NO Probability</div>
            <div className='text-gray-400 text-xs'>red-btn: #D00000</div>
            <div className='text-gray-500 text-xs mt-1'>Used when showing dual probabilities</div>
          </div>
        </div>
      </div>
      
      <table className='min-w-full divide-y divide-gray-200 bg-primary-background'>
        <thead className='bg-gray-50'>
          <tr>
            <th
              scope='col'
              className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'
            >
              Component
            </th>
            <th
              scope='col'
              className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'
            >
              Description
            </th>
            <th
              scope='col'
              className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'
            >
              Import
            </th>
          </tr>
        </thead>
        <tbody className='bg-primary-background divide-y divide-gray-200'>
          <tr>
            <td className='px-6 py-4 text-white'>
              <div className='flex items-center'>
                <Header />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>Header</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import Header from '../../components/header/Header';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4 '>
              <div className='flex items-center'>
                <Sidebar />
              </div>
            </td>
            <td className='px-6 py-4  text-sm text-gray-500'>Sidebar</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import Sidebar from '../../components/sidebar/Sidebar';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <BetYesButton
                  isSelected={isSelected}
                  onClick={() => setIsSelected(!isSelected)}
                />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>Bet YES Button</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import BetYesButton from '../../components/buttons/trade/BetButtons';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <BetNoButton
                  isSelected={isSelected}
                  onClick={() => setIsSelected(!isSelected)}
                />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>Bet NO Button</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import BetNoButton from '../../components/buttons/BetButtons';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <ResolveButton
                  isSelected={isSelected}
                  onClick={() => setIsSelected(!isSelected)}
                />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>
              Neutral Button (Resolve)
            </td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import NeutralButton from '../../components/buttons/marketDetails/ResolveButtons';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <SelectNoButton
                  isSelected={isSelected}
                  onClick={() => setIsSelected(!isSelected)}
                />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>
              Select No Button (Resolutions)
            </td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import SelectNoButton from '../../components/buttons/marketDetails/ResolveButtons';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <SelectYesButton
                  isSelected={isSelected}
                  onClick={() => setIsSelected(!isSelected)}
                />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>
              Select Yes Button (Resolutions)
            </td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import SelectYesButton from '../../components/buttons/marketDetails/ResolveButtons';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <ConfirmResolveButton
                  isSelected={isSelected}
                  onClick={() => setIsSelected(!isSelected)}
                />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>
              ConfirmResolveButton (Resolutions)
            </td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import ConfirmResolveButton from '../../components/buttons/marketDetails/ResolveButtons';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <SiteButton
                  isSelected={isSelected}
                  onClick={() => setIsSelected(!isSelected)}
                />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>SiteButton</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import SiteButton from '../../components/buttons/SiteButton';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex items-center gap-4'>
                <RegularInput />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>RegularInput</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import { RegularInput } from '../../components/inputs/InputBar';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex items-center gap-4'>
                <SuccessInput />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>SuccessInput</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import { SuccessInput } from '../../components/inputs/InputBar';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex items-center gap-4'>
                <ErrorInput />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>ErrorInput</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import { ErrorInput } from '../../components/inputs/InputBar';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex items-center gap-4'>
                <PersonInput />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>PersonInput</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import { PersonInput } from '../../components/inputs/InputBar';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex items-center gap-4'>
                <LockInput />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>LockInput</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import { LockInput } from '../../components/inputs/InputBar';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex items-center gap-4'>
                <RegularInputBox />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>RegularInputBox</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import { RegularInputBox } from '../../components/inputs/InputBox';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <LoginModalButton />
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>
              Login Modal Button
            </td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import LoginModalButton from '../../components/modals/LoginModalClick';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <DatetimeSelector />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>
              Datetime Selector
            </td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import DatetimeSelector from '../../components/datetimeSelector/DatetimeSelector';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <SiteTabs tabs={tabsData} />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>Tabs</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import SiteTabs from '../../components/tabs/SiteTabs';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <ExpandableText
                  text="Will artificial intelligence completely replace human-operated prediction markets?"
                  maxLength={35}
                  className="text-blue-400"
                  buttonClassName="text-xs text-yellow-400 ml-1"
                />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>Expandable Text</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import ExpandableText from '../../components/utils/ExpandableText';`}</code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <ExpandableLink
                  text="Will artificial intelligence completely replace human-operated prediction markets by 2025?"
                  to="/markets/example"
                  maxLength={35}
                  linkClassName="text-blue-400 hover:text-blue-300"
                  buttonClassName="text-xs text-yellow-400 ml-1"
                />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>Expandable Link (Table Use)</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import ExpandableLink from '../../components/utils/ExpandableLink';`}</code>
            </td>
          </tr>
        </tbody>
      </table>
      <table className='min-w-full divide-y divide-gray-200 bg-primary-background'>
        <thead className='bg-gray-50'>
          <tr>
            <th
              scope='col'
              className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'
            >
              Description
            </th>
            <th
              scope='col'
              className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'
            >
              Import
            </th>
          </tr>
        </thead>
        <tbody className='bg-primary-background divide-y divide-gray-200'>
          <tr>
            <td className='px-6 py-4 text-sm text-gray-500'>MarketsTable</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>
                import MarketsTable from '../../components/tables/MarketTables';
              </code>
            </td>
          </tr>
          <tr>
            <td className='px-6 py-4'>
              <div className='flex flex-wrap items-center gap-4'>
                <SharesBadge type="YES" count={8} />
                <SharesBadge type="NO" count={12} />
              </div>
            </td>
            <td className='px-6 py-4 text-sm text-gray-500'>Shares Badges (Portfolio Display)</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>{`import { SharesBadge } from '../../components/buttons/trade/SellButtons';`}</code>
            </td>
          </tr>
        </tbody>
      </table>
      
      {/* Shares Badge Section */}
      <div className='bg-primary-background p-6'>
        <h3 className='text-xl font-bold text-white mb-4 mt-8'>Portfolio Shares Styling</h3>
        <div className='grid grid-cols-1 md:grid-cols-2 gap-4 mb-8'>
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='flex justify-center mb-4'>
              <SharesBadge type="YES" count={8} />
            </div>
            <div className='text-white text-sm font-medium'>YES Shares Badge</div>
            <div className='text-gray-400 text-xs'>Green (#054A29) to Beige (#F9D3A5) gradient</div>
            <div className='text-gray-400 text-xs'>Gold border (#FFC107) with coin emoji 🪙</div>
          </div>
          
          <div className='bg-primary-background p-4 rounded-lg border border-gray-600'>
            <div className='flex justify-center mb-4'>
              <SharesBadge type="NO" count={12} />
            </div>
            <div className='text-white text-sm font-medium'>NO Shares Badge</div>
            <div className='text-gray-400 text-xs'>Red (#D00000) to Beige (#F9D3A5) gradient</div>
            <div className='text-gray-400 text-xs'>Gold border (#FFC107) with coin emoji 🪙</div>
          </div>
        </div>
      </div>
      
      {/* Expandable Text Section */}
      <div className='bg-primary-background p-6'>
        <h3 className='text-xl font-bold text-white mb-4 mt-8'>Expandable Text Component</h3>
        <div className='grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8'>
          <div className='bg-gray-800 p-4 rounded-lg border border-gray-600'>
            <div className='text-white text-sm font-medium mb-2'>Short Text (No expansion needed)</div>
            <div className='text-gray-400 text-xs mb-3'>Text shorter than maxLength displays normally</div>
            <ExpandableText
              text="Will AI replace human prediction markets?"
              maxLength={50}
              className="text-blue-400"
            />
          </div>
          
          <div className='bg-gray-800 p-4 rounded-lg border border-gray-600'>
            <div className='text-white text-sm font-medium mb-2'>Long Text (Expandable)</div>
            <div className='text-gray-400 text-xs mb-3'>Text longer than maxLength gets truncated with expand button</div>
            <ExpandableText
              text="Will artificial intelligence completely replace human-operated prediction markets and forecasting platforms by the end of 2025, considering current technological advances?"
              maxLength={50}
              className="text-blue-400"
              expandedClassName="mt-2 p-3 bg-gray-700 rounded border border-gray-500"
              buttonClassName="text-xs text-yellow-400 hover:text-yellow-300 transition-colors ml-2"
              expandIcon="🔍"
            />
          </div>
        </div>
        
        <div className='bg-gray-900 p-4 rounded-lg'>
          <h4 className='text-lg font-semibold text-white mb-3'>Usage Examples</h4>
          <div className='space-y-4 text-sm'>
            <div>
              <div className='text-emerald-400 mb-1'>Basic Usage:</div>
              <code className='text-xs text-gray-300 bg-gray-800 px-2 py-1 rounded block'>
                {`<ExpandableText text="Long question title..." maxLength={45} />`}
              </code>
            </div>
            <div>
              <div className='text-emerald-400 mb-1'>With Custom Styling:</div>
              <code className='text-xs text-gray-300 bg-gray-800 px-2 py-1 rounded block whitespace-pre-wrap'>
{`<ExpandableText
  text="Long text here..."
  maxLength={60}
  className="text-blue-400"
  expandedClassName="mt-2 p-2 bg-gray-700 rounded"
  buttonClassName="text-xs text-yellow-400 ml-1"
  expandIcon="📖"
/>`}
              </code>
            </div>
            <div>
              <div className='text-emerald-400 mb-1'>Import Statement:</div>
              <code className='text-xs text-gray-300 bg-gray-800 px-2 py-1 rounded block'>
                {`import ExpandableText from '../utils/ExpandableText';`}
              </code>
            </div>
          </div>
        </div>
        
        <div className='bg-gray-900 p-4 rounded-lg mt-4'>
          <h4 className='text-lg font-semibold text-white mb-3'>Props</h4>
          <div className='grid grid-cols-1 md:grid-cols-2 gap-4 text-sm'>
            <div>
              <div className='text-blue-400 font-mono'>text</div>
              <div className='text-gray-400 text-xs'>Required string - The text to display</div>
            </div>
            <div>
              <div className='text-blue-400 font-mono'>maxLength</div>
              <div className='text-gray-400 text-xs'>Number (default: 50) - Characters before truncation</div>
            </div>
            <div>
              <div className='text-blue-400 font-mono'>className</div>
              <div className='text-gray-400 text-xs'>String - CSS classes for text container</div>
            </div>
            <div>
              <div className='text-blue-400 font-mono'>expandedClassName</div>
              <div className='text-gray-400 text-xs'>String - CSS classes for expanded view</div>
            </div>
            <div>
              <div className='text-blue-400 font-mono'>buttonClassName</div>
              <div className='text-gray-400 text-xs'>String - CSS classes for expand button</div>
            </div>
            <div>
              <div className='text-blue-400 font-mono'>expandIcon</div>
              <div className='text-gray-400 text-xs'>String (default: '📐') - Icon for expand button</div>
            </div>
          </div>
        </div>
      </div>
      
      <MarketsTable></MarketsTable>
      <table className='min-w-full divide-y divide-gray-200 bg-primary-background'>
        <thead className='bg-gray-50'>
          <tr>
            <th
              scope='col'
              className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'
            >
              Description
            </th>
            <th
              scope='col'
              className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'
            >
              Import
            </th>
          </tr>
        </thead>
        <tbody className='bg-primary-background divide-y divide-gray-200'>
          <tr>
            <td className='px-6 py-4 text-sm text-gray-500'>
              MarketChartDetailsTable
            </td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>
                import MarketDetailsTable from
                '../../components/marketDetails/MarketDetailsLayout';
              </code>
            </td>
          </tr>
        </tbody>
      </table>
      <hr></hr>
      <p></p>
      <table className='min-w-full divide-y divide-gray-200 bg-primary-background'>
        <thead className='bg-gray-50'>
          <tr>
            <th
              scope='col'
              className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'
            >
              Description
            </th>
            <th
              scope='col'
              className='px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider'
            >
              Import
            </th>
          </tr>
        </thead>
        <tbody className='bg-primary-background divide-y divide-gray-200'>
          <tr>
            <td className='px-6 py-4 text-sm text-gray-500'>MarketChart</td>
            <td className='px-6 py-4 text-sm font-mono text-gray-500'>
              <code>
                import MarketChart from '../../components/charts/MarketChart';
              </code>
            </td>
          </tr>
        </tbody>
      </table>
      <MarketChart
        data={TestMarketData.probabilityChanges}
        title={TestMarketData.market.questionTitle}
        className='shadow-md border border-custom-gray-light'
      />

      <div className='flex justify-star items-start flex-col flex-start'>
        <p>Loading spinner</p>

        <LoadingSpinner />
      </div>
    </div>
  );
};

export default Style;
