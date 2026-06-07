import React from 'react';
import SiteTabs from '../../components/tabs/SiteTabs';
import HomeEditor from './HomeEditor';
import MarketDiscoveryLayoutEditor from './MarketDiscoveryLayoutEditor';
import ReportingVisibilityEditor from './ReportingVisibilityEditor';
import SocialShareEditor from './SocialShareEditor';

function CmsDashboard() {
  const tabsData = [
    {
      label: 'Market Discovery Layout',
      content: <MarketDiscoveryLayoutEditor />,
    },
    {
      label: 'Home Page',
      content: <HomeEditor />,
    },
    {
      label: 'Reporting Visibility',
      content: <ReportingVisibilityEditor />,
    },
    {
      label: 'Social Share',
      content: <SocialShareEditor />,
    },
  ];

  return (
    <section className="bg-primary-background text-white">
      <div className="p-6 pb-5">
        <p className="text-xs uppercase tracking-[0.22em] text-primary-pink">Admin CMS</p>
        <h1 className="mt-2 text-2xl font-bold">Content And Discovery Management</h1>
        <p className="mt-2 max-w-3xl text-sm text-gray-300">
          Manage public homepage content, market discovery layout planning, and social share defaults.
        </p>
      </div>
      <SiteTabs tabs={tabsData} defaultTab="Market Discovery Layout" />
    </section>
  );
}

export default CmsDashboard;
