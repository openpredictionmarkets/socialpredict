import React from 'react';
import SiteTabs from './SiteTabs';
import BetsActivityLayout from '../layouts/activity/bets/BetsActivity';

const ActivityTabs = ({ marketId }) => {
    const tabsData = [
        { label: 'Comments', content: <div>Comments Go here...</div> },
        { label: 'Positions', content: <div>Positions Go here...</div> },
        { label: 'Bets', content: <BetsActivityLayout marketId={marketId} /> },
    ];

    return <SiteTabs tabs={tabsData} />;
};

export default ActivityTabs;